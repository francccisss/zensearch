package main

import (
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/tebeka/selenium"
	"log"
	"strings"
	"sync"
	rabbitmqclient "web-crawler-service-golang/pkg/rabbitmq_client"
	webdriver "web-crawler-service-golang/pkg/webdriver"
	utilities "web-crawler-service-golang/utilities/links"
)

type Header struct {
	Title string
	Url   string
}

type WebpageEntry struct {
	URL             string
	IndexedWebpages []IndexedWebpage
	hostname        string
	Title           string
}
type IndexedWebpage struct {
	Header   Header
	Contents string
}

type Crawler struct {
	URL string
	ctx context.Context
	wd  *selenium.WebDriver
}

var elementSelector = []string{
	"a",
	"p",
	"span",
	"pre",
	"h1",
	"h2",
	"h3",
	"h4",
}

type Results struct {
	URLCount    int
	URLsFailed  []string
	Message     string
	ThreadsUsed int
	PageResult  <-chan PageResult
}

var indexedList map[string]IndexedWebpage

const (
	maxRetries   = 7
	threadPool   = 10
	crawlFail    = 0
	crawlSuccess = 1
	// removes links to web objects that does not return an html page.
	linkFilter = `a:not([href$=".zip"]):not([href$=".svg"]):not([href$=".scss"]):not([href$=".css"]):not([href$=".pdf"]):not([href$=".exe"]):not([href$=".jpg"]):not([href$=".png"]):not([href$=".tar.gz"]):not([href$=".rar"]):not([href$=".7z"]):not([href$=".mp3"]):not([href$=".mp4"]):not([href$=".mkv"]):not([href$=".tar"]):not([href$=".xz"]):not([href$=".msi"])`
)

type PageNavigator struct {
	entry        *WebpageEntry
	wd           *selenium.WebDriver
	pagesVisited map[string]string
	currentUrl   string
}

type PageResult struct {
	URL         string
	Message     string
	CrawlStatus int
	TotalPages  int
}

type Message struct {
	Webpages []IndexedWebpage
	Header
}

type ErrorMessage struct {
	Message string
	Url     string
}

type Spawner struct {
	threadPool int
	URLs       []string
}
type Spawn struct {
	threadSlot           chan struct{}
	wg                   *sync.WaitGroup
	ctx                  context.Context
	aggregateResultsChan chan PageResult
}

type ThreadToken struct{}

func (sp *Spawn) SpawnCrawler(doc string) {
	defer func() {
		<-sp.threadSlot
		log.Printf("NOTIF: Thread Token release\n")
	}()
	defer sp.wg.Done()
	wd, err := webdriver.CreateClient()
	if err != nil {
		log.Print(err.Error())
		log.Printf("ERROR: Unable to create a new connection with Chrome Web Driver.\n")
		return
	}
	crawler := Crawler{ctx: sp.ctx, URL: doc, wd: wd}
	result, err := crawler.Crawl()
	if err != nil {
		log.Print(err.Error())
		log.Printf("NOTIF: Thread token release due to error.\n")
		return
	}
	sp.aggregateResultsChan <- result
}

func (s *Spawner) SpawnCrawlers() Results {
	aggregateResultsChan := make(chan PageResult, len(s.URLs))
	threadSlot := make(chan struct{}, s.threadPool)

	var (
		wg  sync.WaitGroup
		ctx context.Context
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, doc := range s.URLs {
			wg.Add(1)
			threadSlot <- ThreadToken{}
			log.Printf("NOTIF: Thread token insert\n")
			spawn := Spawn{aggregateResultsChan: aggregateResultsChan, threadSlot: threadSlot, wg: &wg, ctx: ctx}
			go spawn.SpawnCrawler(doc)
		}
	}()

	log.Printf("NOTIF: Wait for crawlers\n")
	wg.Wait()
	close(aggregateResultsChan)
	for crawler := range aggregateResultsChan {
		log.Printf("Crawled URL: %s\n", crawler.URL)
		log.Printf("Crawl Message: %s\n", crawler.Message)
	}

	log.Println("NOTIF: All Process have finished.")

	return Results{
		Message:     "Crawled and indexed webpages",
		ThreadsUsed: threadPool,
		URLCount:    len(s.URLs),
		PageResult:  aggregateResultsChan,
	}
}

func (c Crawler) Crawl() (PageResult, error) {
	defer log.Printf("NOTIF: Finished Crawling\n")
	defer (*c.wd).Close()

	log.Printf("NOTIF: Start Crawling %s\n", c.URL)
	hostname, err := utilities.GetHostname(c.URL)
	if err != nil {
		fmt.Println(err.Error())
	}
	entry := WebpageEntry{
		URL:             c.URL,
		IndexedWebpages: make([]IndexedWebpage, 0, 10),
		hostname:        hostname,
		Title:           "",
	}
	pageNavigator := PageNavigator{
		entry:        &entry,
		currentUrl:   c.URL,
		wd:           c.wd,
		pagesVisited: map[string]string{},
	}
	err = pageNavigator.navigatePageWithRetries(maxRetries)
	if err != nil {
		sendErrorOnWebpageCrawl(c.URL)
		fmt.Println(err.Error())
		return PageResult{}, fmt.Errorf("ERROR: Unable to navigate to the website entry point.")
	}
	title, err := (*c.wd).Title()
	if err == nil {
		entry.Title = title
	}
	err = pageNavigator.navigatePages()

	/*
	 TODO when other pages are already indexed, but only a single page throws an error
	 then all progress with huge amounts of data will be lost, need to save these Results
	 despite an error occurs instead of returning an zero byte result.
	*/
	if err != nil {
		sendErrorOnWebpageCrawl(c.URL)
		fmt.Println("ERROR: Well something went wrong with the last stack.")
		return PageResult{
			URL:         c.URL,
			CrawlStatus: crawlFail,
			Message:     "An Error has occured while crawling the current url.",
			TotalPages:  len(entry.IndexedWebpages),
		}, nil
	}
	result := PageResult{
		URL:         c.URL,
		CrawlStatus: crawlSuccess,
		Message:     "Successfully Crawled & Indexed website",
		TotalPages:  len(entry.IndexedWebpages),
	}
	// saveIndexedWebpages(hostname, &entry)
	return result, nil
}

func (pn *PageNavigator) navigatePageWithRetries(retries int) error {
	if retries > 0 {
		err := (*pn.wd).Get(pn.currentUrl)
		if err != nil {
			return pn.navigatePageWithRetries(retries - 1)
		}
		return nil
	}
	return fmt.Errorf("ERROR: Unable to retrieve webpage after several retries.")
}

func (pn *PageNavigator) navigatePages() error {

	if _, visited := pn.pagesVisited[pn.currentUrl]; visited {
		// its so that we can grab unique links and append to children of the current page
		fmt.Println("NOTIF: Page already visited")
		return nil
	}
	err := pn.navigatePageWithRetries(maxRetries)
	if err != nil {
		return fmt.Errorf("ERROR: Unable to visit the current link.")
	}
	pn.pagesVisited[pn.currentUrl] = pn.currentUrl

	links, err := (*pn.wd).FindElements(selenium.ByCSSSelector, linkFilter)

	// no children/error
	if err != nil {
		log.Println("ERROR: Unable to find elements of type `a`.")
		return fmt.Errorf("ERROR: Unable to find elements of type `a`.")
	}

	/*
		Need to check such that we can ignore the already visited links
		and use the ones that doesnt exist and consider it
		as the child of the currently visited link
	*/

	children := make([]string, 0)
	for _, link := range links {
		// need to filter out links that is not the same as hostname
		ref, _ := link.GetAttribute("href")
		cleanedRef, _, _ := strings.Cut(ref, "#")
		childHostname, err := utilities.GetHostname(cleanedRef)

		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		if _, visited := pn.pagesVisited[cleanedRef]; !visited && pn.entry.hostname == childHostname {
			// if so that we can grab unique links and append to children of the current page
			// and ignore links not relative to the entry point link
			children = append(children, cleanedRef)
		}
	}

	indexedWebpage, err := pn.Index()
	if err != nil {
		log.Println("ERROR: Not handled yet")
	}
	pn.entry.IndexedWebpages = append(pn.entry.IndexedWebpages, indexedWebpage)

	/*
	 no child to traverse to then return to caller, the caller function will
	 then go to its next child in the children array.
	*/

	if len(children) == 0 {
		return nil
	}
	for _, child := range children {
		pn.currentUrl = child
		err := pn.navigatePages()
		// if error occured from traversing or any error has occured
		// just move to the next child
		if err != nil {
			continue
		}

	}
	return nil
}
func (pt PageNavigator) Index() (IndexedWebpage, error) {

	/*
		Iterating through the indexSelector, where each selector, we creates
		a new go routine, so using a buffered channel with the exact length of
		the indexSelector would make more sense.

		If ever we want to throttle the operation we can create a semaphore by
		limiting the buffered channel, if resource is an issue.
	*/

	textContentChan := make(chan string, len(elementSelector))
	var wg sync.WaitGroup

	// Start wait group after go routine is processed on a different thread
	wg.Add(1)

	// Go routine generator
	go func() {
		defer func() {
			wg.Done()
			log.Println("NOTIF: Text Extracted from elements.")
		}()
		for _, selector := range elementSelector {
			wg.Add(1)
			go func(selector string) {
				defer wg.Done()
				textContents, err := textContentByElementSelector(pt.wd, selector)
				if err != nil {
					textContentChan <- ""
					return
				}
				textContentChan <- joinContents(textContents)
			}(selector)
		}
	}()

	fmt.Println("NOTIF: Waiting for page indexer")
	wg.Wait()
	close(textContentChan)
	textChanSlice := make([]string, 0, 100)

	/*
		create a select statement to catch an error returned by
		invdividual go routine element selectors
	*/

	for elementContents := range textContentChan {
		textChanSlice = append(textChanSlice, elementContents)
	}

	pageContents := joinContents(textChanSlice)
	title, err := (*pt.wd).Title()
	if err != nil {
		log.Printf("ERROR: No title for this page")
	}

	url, err := (*pt.wd).CurrentURL()
	if err != nil {
		log.Printf("ERROR: No url for this page")
	}
	fmt.Printf("NOTIF: Page %s Indexed\n", url)

	newIndexedPage := IndexedWebpage{
		Contents: pageContents,
		Header: Header{
			Url:   url,
			Title: title,
		},
	}
	return newIndexedPage, nil
}

func textContentByElementSelector(wd *selenium.WebDriver, selector string) ([]string, error) {
	elementTextContents := make([]string, 0, 10)
	elements, err := (*wd).FindElements(selenium.ByCSSSelector, selector)
	if err != nil {
		log.Printf("ERROR: Elements does not satisfy css selector: %s", selector)
		return nil, err
	}
	for _, el := range elements {
		text, err := el.Text()
		if err != nil {
			continue
		}
		elementTextContents = append(elementTextContents, text)
	}

	return elementTextContents, nil
}

func joinContents(tc []string) string {
	return strings.Join(tc, " ")
}

func sendErrorOnWebpageCrawl(hostname string) error {

	/*
	 expressErrorChannel is used to pass a message from crawl service
	 to the express websocket server using the `crawl_poll_queue` (Might change routing key name)
	 routing key to notify users immediately that an error occured for the current
	 to notify users immediately that an error occured for the current url
	*/

	fmt.Println("ERROR Crawl")
	conn, err := rabbitmqclient.GetConnection("receiverConn")
	if err != nil {
		fmt.Print(err.Error())
		log.Panicln("ERROR: Unable to get connection.")
	}
	expressErrorChannel, err := conn.Channel()
	if err != nil {
		log.Printf("ERROR: Unable to create a database channel.")
		return fmt.Errorf(err.Error())
	}

	errorMessage := ErrorMessage{
		Message: "Error",
		Url:     hostname,
	}
	const crawlPollqueue = "crawl_poll_queue"

	expressErrorChannel.QueueDeclare(crawlPollqueue, false, false, false, false, nil)
	dataBuffer, err := json.Marshal(errorMessage)
	expressErrorChannel.Publish("", crawlPollqueue, false, false, amqp.Publishing{
		Type:    "text/plain",
		Body:    []byte(dataBuffer),
		ReplyTo: crawlPollqueue,
	})
	expressErrorChannel.Close()
	return nil
}

func saveIndexedWebpages(jobID string, entry *WebpageEntry) error {
	conn, err := rabbitmqclient.GetConnection("receiverConn")
	if err != nil {
		fmt.Print(err.Error())
		log.Panicln("ERROR: Unable to get connection.")
	}
	dbChannel, err := conn.Channel()
	if err != nil {
		log.Printf("ERROR: Unable to create a database channel.")
		return fmt.Errorf(err.Error())
	}

	resultMessage := Message{
		Webpages: entry.IndexedWebpages,
		Header: Header{
			Title: entry.Title,
			Url:   entry.hostname,
		},
	}
	const (
		dbIndexingQueue = "db_indexing_crawler"
		crawlPollqueue  = "crawl_poll_queue"
	)
	dbChannel.QueueDeclare(dbIndexingQueue, false, false, false, false, nil)
	dataBuffer, err := json.Marshal(resultMessage)
	dbChannel.Publish("", dbIndexingQueue, false, false, amqp.Publishing{
		Type:          "text/plain",
		Body:          []byte(dataBuffer),
		CorrelationId: jobID,
		ReplyTo:       crawlPollqueue,
	})
	dbChannel.Close()
	return nil
}
