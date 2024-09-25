package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	rabbitmqclient "web-crawler-service-golang/pkg/rabbitmq_client"
	webdriver "web-crawler-service-golang/pkg/webdriver"
	utilities "web-crawler-service-golang/utilities/links"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/tebeka/selenium"
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
	URLs []string
}

type CrawlTask struct {
	URL    string
	ctx    context.Context
	wd     *selenium.WebDriver
	docLen int
}

var indexSelector = []string{
	"a",
	"p",
	"span",
	"code",
	"pre",
	"h1",
	"h2",
	"h3",
	"h4",
}

type PageIndexer struct {
	wd *selenium.WebDriver
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
	threadPool   = 10
	crawlFail    = 0
	crawlSuccess = 1
	// removes links to web objects that does not return an html page.
	linkFilter = `a:not([href$=".zip"]):not([href$=".pdf"]):not([href$=".exe"]):not([href$=".jpg"]):not([href$=".png"]):not([href$=".tar.gz"]):not([href$=".rar"]):not([href$=".7z"]):not([href$=".mp3"]):not([href$=".mp4"]):not([href$=".mkv"]):not([href$=".tar"]):not([href$=".xz"]):not([href$=".msi"])`
)

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

	dbChannel.QueueDeclare("database_push_queue", false, false, false, false, nil)
	dataBuffer, err := json.Marshal(resultMessage)
	dbChannel.Publish("", "database_push_queue", false, false, amqp.Publishing{
		Type:          "text/plain",
		Body:          []byte(dataBuffer), // TODO convert to buffer array / byte
		CorrelationId: jobID,
	})
	return nil
}

/*
  TODO need to close down chrome driver session after every crawl
  be it success or fail.
*/

func (c Crawler) Start() Results {
	aggregateChan := make(chan PageResult, len(c.URLs))
	semaphore := make(chan struct{}, threadPool)

	var (
		wg  sync.WaitGroup
		ctx context.Context
	)

	// create crawlers
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, doc := range c.URLs {
			wg.Add(1)
			semaphore <- struct{}{}
			log.Printf("NOTIF: Semaphore token insert\n")
			go func(doc string) {
				defer func() {
					<-semaphore
					log.Printf("NOTIF: Semaphore token release\n")
				}()
				defer wg.Done()
				wd, err := webdriver.CreateClient()
				if err != nil {
					log.Print(err.Error())
					log.Printf("ERROR: Unable to create a new connection with Chrome Web Driver.\n")
					return
				}
				crawler := CrawlTask{ctx: ctx, URL: doc, wd: wd, docLen: len(c.URLs)}
				status, err := crawler.Crawl()
				if err != nil {
					log.Print(err.Error())
					log.Printf("NOTIF: Semaphore token release due to error.\n")
					return
				}
				aggregateChan <- status
			}(doc)
		}
	}()

	log.Printf("NOTIF: Wait for crawlers\n")
	wg.Wait()
	close(aggregateChan)

	for crawler := range aggregateChan {
		log.Printf("Crawled URL: %s\n", crawler.URL)
		log.Printf("Crawl Message: %s\n", crawler.Message)
	}

	log.Println("NOTIF: All Process have finished.")

	return Results{
		Message:     "Crawled and indexed webpages",
		ThreadsUsed: threadPool,
		URLCount:    len(c.URLs),
		PageResult:  aggregateChan,
	}
}

func (ct CrawlTask) Crawl() (PageResult, error) {
	defer log.Printf("NOTIF: Finished Crawling\n")
	defer (*ct.wd).Close()

	// Initialization

	log.Printf("NOTIF: Start Crawling %s\n", ct.URL)
	indexer := PageIndexer{wd: ct.wd}
	hostname, err := utilities.GetOrigin(ct.URL)
	if err != nil {
		fmt.Println(err.Error())
	}
	entry := WebpageEntry{
		URL:             ct.URL,
		IndexedWebpages: make([]IndexedWebpage, 0, ct.docLen),
		hostname:        hostname,
		Title:           "",
	}
	pageTraverser := PageTraverser{
		entry:        &entry,
		currentUrl:   ct.URL,
		indexer:      &indexer,
		pagesVisited: map[string]string{},
	}
	_ = (*ct.wd).Get(ct.URL)
	title, _ := (*ct.wd).Title()
	entry.Title = title
	err = pageTraverser.traversePages()
	if err != nil {
		fmt.Println("ERROR: Well something went wrong with the last stack.")
	}
	result := PageResult{
		URL:         ct.URL,
		CrawlStatus: crawlFail,
		Message:     "Successfully Crawled & Indexed website",
		TotalPages:  len(entry.IndexedWebpages),
	}
	// if > 1 threads a have finished processing
	// and pass webpages into the save function, and both of them
	// are being sent at the same time
	// create differet correlation IDs

	saveIndexedWebpages(hostname, &entry)

	return result, nil

}

type PageTraverser struct {
	entry        *WebpageEntry
	indexer      *PageIndexer
	pagesVisited map[string]string
	currentUrl   string
}

func (pt *PageTraverser) traversePages() error {

	if _, visited := pt.pagesVisited[pt.currentUrl]; visited {
		// its so that we can grab unique links and append to children of the current page
		fmt.Println("NOTIF: Page already visited")
		return nil
	}
	err := (*pt.indexer.wd).Get(pt.currentUrl)
	if err != nil {
		return fmt.Errorf("ERROR: Unable to visit the current link.")
	}
	pt.pagesVisited[pt.currentUrl] = pt.currentUrl

	links, err := (*pt.indexer.wd).FindElements(selenium.ByCSSSelector, linkFilter)

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
		// need to filter out links that is not the same as origin
		ref, _ := link.GetAttribute("href")
		cleanedRef, _, _ := strings.Cut(ref, "#")
		childHostname, err := utilities.GetOrigin(cleanedRef)

		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		if _, visited := pt.pagesVisited[cleanedRef]; !visited && pt.entry.hostname == childHostname {
			// its so that we can grab unique links and append to children of the current page
			children = append(children, cleanedRef)
		}
	}
	/*
		Start indexing the current page and pushing into the indexed webpages slice
	*/
	indexedWebpage, err := pt.indexer.Index()
	if err != nil {
		log.Println("ERROR: Not handled yet")
	}
	pt.entry.IndexedWebpages = append(pt.entry.IndexedWebpages, indexedWebpage)

	/*
	 no child to traverse to then return to caller, the caller function will
	 then go to its next child in the children array.
	*/

	if len(children) == 0 {
		return nil
	}
	for _, child := range children {
		pt.currentUrl = child
		err := pt.traversePages()
		// if error occured from traversing or any error has occured
		// just move to the next child
		if err != nil {
			continue
		}

	}
	return nil
}
func (p PageIndexer) Index() (IndexedWebpage, error) {

	/*
		Iterating through the indexSelector, where each selector, we creates
		a new go routine, so using a buffered channel with the exact length of
		the indexSelector would make more sense.

		If ever we want to throttle the operation we can create a semaphore by
		limiting the buffered channel, if resource is an issue.
	*/

	textContentChan := make(chan string, len(indexSelector))
	var wg sync.WaitGroup

	// Start wait group after go routine is processed on a different thread
	wg.Add(1)

	// Go routine generator
	go func() {
		defer func() {
			wg.Done()
			log.Println("NOTIF: Text Extracted from elements.")
		}()
		for _, selector := range indexSelector {
			wg.Add(1)
			go func(selector string) {
				defer wg.Done()
				defer fmt.Printf("NOTIF: Selector: %s done\n", selector)
				textContents, err := textContentByIndexSelector(p.wd, selector)
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
	title, err := (*p.wd).Title()
	if err != nil {
		log.Printf("ERROR: No title for this page")
	}

	url, err := (*p.wd).CurrentURL()
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

func textContentByIndexSelector(wd *selenium.WebDriver, selector string) ([]string, error) {
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
