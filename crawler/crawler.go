package main

import (
	"context"
	rabbitmqclient "crawler/internal/rabbitmq"
	webdriver "crawler/internal/webdriver"
	utilities "crawler/utilities"
	"encoding/json"
	"fmt"
	"log"
	"sync"

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

var elementSelector = []string{
	"a",
	"p",
	"span",
	"pre",
	"h1",
	"h2",
	"h3",
	"h4",
	"td",
	"ul",
	"code",
	"div",
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
)

type PageResult struct {
	URL         string
	Message     string
	CrawlStatus int
	TotalPages  int
}

type MessageResult struct {
	Webpages []IndexedWebpage
	Header
	Message     string
	CrawlStatus int
}

type ErrorMessage struct {
	Message     string
	Url         string
	CrawlStatus int
}

type Spawner struct {
	threadPool int
	URLs       []string
}

type Crawler struct {
	URL string
	ctx context.Context
	wd  *selenium.WebDriver
}

type ThreadToken struct{}

func NewCrawler(entryPoint string, ctx context.Context) (*Crawler, error) {
	c := &Crawler{
		URL: entryPoint,
		ctx: ctx,
	}
	wd, err := webdriver.CreateClient()
	if err != nil {
		log.Print(err.Error())
		log.Printf("ERROR: Unable to create a new connection with Chrome Web Driver.\n")
		return nil, err
	}
	c.wd = wd
	return c, nil
}

func NewSpawner(threadpool int, URLs []string) *Spawner {
	return &Spawner{
		threadPool: threadpool,
		URLs:       URLs,
	}
}

func (s *Spawner) SpawnCrawlers() Results {
	resultsChan := make(chan PageResult, len(s.URLs))
	threadSlot := make(chan struct{}, s.threadPool)

	var (
		wg  sync.WaitGroup
		ctx context.Context
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, entryPoint := range s.URLs {
			wg.Add(1)
			threadSlot <- ThreadToken{}
			log.Printf("NOTIF: Thread token insert\n")
			go func() {
				crawler, err := NewCrawler(entryPoint, ctx)
				if err != nil {
					log.Print(err.Error())
					log.Printf("NOTIF: Thread token release due to error.\n")
					<-threadSlot
					return
				}
				defer func() {
					<-threadSlot
					log.Printf("NOTIF: Thread Token release\n")
				}()
				defer wg.Done()
				result, err := crawler.Crawl()
				if err != nil {
					log.Print(err.Error())
					log.Printf("NOTIF: Thread token release due to error.\n")
					return
				}
				defer (*crawler.wd).Quit()
				resultsChan <- result
			}()
		}
	}()

	log.Printf("NOTIF: Wait for crawlers\n")
	wg.Wait()
	close(resultsChan)
	for crawler := range resultsChan {
		log.Printf("Crawled URL: %s\n", crawler.URL)
		log.Printf("Crawl Message: %s\n", crawler.Message)
	}

	log.Println("NOTIF: All Process have finished.")

	return Results{
		Message:     "Crawled and indexed webpages",
		ThreadsUsed: threadPool,
		URLCount:    len(s.URLs),
		PageResult:  resultsChan,
	}
}

func (c Crawler) Crawl() (PageResult, error) {
	defer log.Printf("NOTIF: Finished Crawling\n")
	defer (*c.wd).Close()

	log.Printf("NOTIF: Start Crawling %s\n", c.URL)
	hostname, _, err := utilities.GetHostname(c.URL)
	disallowedPaths, err := utilities.ExtractRobotsTxt(c.URL)
	if err != nil {
		fmt.Println("ERROR: Unable to extract robots.txt")
		fmt.Println(err.Error())
	}

	// bro I only understand english :D just remove the ones that you want to be included
	languagePaths := []string{"/es/", "/ko/", "/tr/", "/th/", "/it/", "/uk/", "/sk/", "/fr/", "/de/", "/zh/", "/ja/", "/ru/", "/ar/", "/pt/", "/hi/", "/zh/", "/zh-tw/", "/zh-c/", "/zh-cn/", "/pt-br/", "/uz/"}
	disallowedPaths = append(disallowedPaths, languagePaths...)
	fmt.Printf("DISALLOWED PATHS: %+v\n", disallowedPaths)
	entry := WebpageEntry{
		URL:             c.URL,
		IndexedWebpages: make([]IndexedWebpage, 0, 10),
		hostname:        hostname,
		Title:           "",
	}
	pageNavigator := PageNavigator{
		entry:        &entry,
		wd:           c.wd,
		pagesVisited: map[string]string{},
		queue: Queue{
			array: []string{c.URL},
		},
		disallowedPaths: disallowedPaths,
	}

	maxRetries := 7
	err = pageNavigator.navigatePageWithRetries(maxRetries, c.URL)
	errorMessage := &ErrorMessage{
		CrawlStatus: crawlFail,
		Url:         hostname,
		Message:     "Unable to crawl the source Url",
	}
	if err != nil {
		fmt.Printf("ERROR: Unable to navigate to source url %s\n", c.URL)
		fmt.Printf(err.Error())
		// Error for when the crawler was not able to start crawling from the source.
		err = sendResult(errorMessage.sendErrorOnWebpageCrawl, "crawl_poll_queue", "", "")
		if err != nil {
			fmt.Printf(err.Error())
		}
		return PageResult{}, fmt.Errorf("ERROR: Unable to navigate to the website entry point.\n")
	}
	title, err := (*c.wd).Title()
	if err == nil {
		entry.Title = title
	}

	/*
	 to prevent duplicates if user adds a url that does not have a suffix of `/`
	 the hashmap will consider it as not the same, and we cant use strings.Contain().
	 I know its ugly.
	*/

	if c.URL[len(c.URL)-1] != '/' {
		c.URL += "/"
	}

	err = pageNavigator.navigatePages(c.URL)

	// clean up memory resource since it will linger in memory in the heap
	// once this function is removed from the stack.
	defer clear(entry.IndexedWebpages)
	defer clear(pageNavigator.pagesVisited)

	/*
	   TODO Improve error handling code, it looks ugly.
	*/
	var result PageResult
	if err != nil {
		// Error for when crawler is not able to crawl and index the remaining webpages.
		errorMessage.Message = "Something went wrong while crawling the webpage"
		fmt.Printf("ERROR: Crawler returned with errors from navigating %s\n", c.URL)
		fmt.Printf("ERROR MESSAGE: \n")
		fmt.Println(err.Error())
		sendResult(errorMessage.sendErrorOnWebpageCrawl, "crawl_poll_queue", "", "")
		// err = sendResult(entry.saveIndexedWebpages, "db_indexing_crawler", "crawl_poll_queue", "Crawler was stopped but was able to index the website.")
		result = PageResult{
			URL:         c.URL,
			CrawlStatus: crawlFail,
			Message:     errorMessage.Message,
			TotalPages:  len(entry.IndexedWebpages),
		}
		return result, nil
	}

	fmt.Printf("NOTIF: Crawler returned with no errors from navigating %s\n", c.URL)
	result = PageResult{
		URL:         c.URL,
		CrawlStatus: crawlSuccess,
		Message:     "Successfully Crawled & Indexed website",
		TotalPages:  len(entry.IndexedWebpages),
	}
	err = sendResult(entry.saveIndexedWebpages, "db_indexing_crawler", "crawl_poll_queue", "Successfully Crawled and Indexed Website.")

	return result, nil
}

/*
Returns an array of text contents from an array of common elements specified
by the current selector eg: p, a, span etc.
*/
func sendResult(constructMessage func(message string) ([]byte, error), routingKey string, callbackQueue string, message string) error {
	conn, err := rabbitmqclient.GetConnection("conn")
	if err != nil {
		fmt.Print(err.Error())
		log.Panicf("ERROR: Unable to get %s connection.\n", "conn")
	}
	channel, err := conn.Channel()
	if err != nil {
		log.Printf("ERROR: Unable to create a new channel.\n")
		return err
	}

	defer channel.Close()

	messageBuffer, err := constructMessage(message)
	if err != nil {
		return err
	}

	// need to  check declaration and publication before returning nil
	_, err = channel.QueueDeclare(routingKey, false, false, false, false, nil)
	if err != nil {
		log.Printf("ERROR: Unable to declare %s channel.\n", routingKey)
		return err
	}
	err = channel.Publish("", routingKey, false, false, amqp.Publishing{
		Type:    "text/plain",
		Body:    []byte(messageBuffer),
		ReplyTo: callbackQueue,
	})
	if err != nil {
		log.Printf("ERROR: Unable to publish message to %s channel.\n", routingKey)
		return err
	}
	return nil
}

func (e ErrorMessage) sendErrorOnWebpageCrawl(message string) ([]byte, error) {
	dataBuffer, err := json.Marshal(e)
	if err != nil {
		return []byte{}, fmt.Errorf("ERROR: Unable to marshal Error Message.")
	}
	return dataBuffer, nil
}

func (e *WebpageEntry) saveIndexedWebpages(message string) ([]byte, error) {

	dataBuffer, err := json.Marshal(MessageResult{
		Webpages: e.IndexedWebpages,
		Header: Header{
			Title: e.Title,
			Url:   e.hostname,
		},
		Message:     message,
		CrawlStatus: crawlSuccess,
	},
	)
	if err != nil {
		return []byte{}, fmt.Errorf("ERROR: Unable to marshal message")
	}
	return dataBuffer, nil
}
