package main

import (
	"context"
	rabbitmq "crawler/internal/rabbitmq"
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
	CrawlResult <-chan CrawlResult
}

var indexedList map[string]IndexedWebpage

const (
	crawlFail    = 1
	crawlSuccess = 0
)

type CrawlResult struct {
	URL         string
	Message     string
	CrawlStatus int
	TotalPages  int
}

type IndexedResult struct {
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
	crawlResultsChan := make(chan CrawlResult, len(s.URLs))
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
				crawlResultsChan <- result
			}()
		}
	}()

	log.Printf("NOTIF: Wait for crawlers\n")
	wg.Wait()
	close(crawlResultsChan)
	for crawler := range crawlResultsChan {
		log.Printf("Crawled URL: %s\n", crawler.URL)
		log.Printf("Crawl Message: %s\n", crawler.Message)
	}

	log.Println("NOTIF: All Process have finished.")

	return Results{
		Message:     "Crawled and indexed webpages",
		ThreadsUsed: s.threadPool,
		URLCount:    len(s.URLs),
		CrawlResult: crawlResultsChan,
	}
}

func (c Crawler) Crawl() (CrawlResult, error) {
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
	if err != nil {
		fmt.Printf("ERROR: Unable to navigate to source url %s\n", c.URL)
		fmt.Println(err.Error())

		err = sendIndexedResult(IndexedResult{
			CrawlStatus: crawlFail,
			Header:      Header{Url: hostname, Title: entry.Title},
			Message:     "Unable to crawl the source Url",
			Webpages:    []IndexedWebpage{},
		},
		)
		if err != nil {
			fmt.Println(err.Error())
			return CrawlResult{}, fmt.Errorf("ERROR: Unable to send error result to db.\n")
		}
		return CrawlResult{}, fmt.Errorf("ERROR: Unable to navigate to the website entry point.\n")
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
	var result CrawlResult

	message := "Successfully Crawled & Indexed website"
	if err != nil {
		// Error for when crawler is not able to crawl and index the remaining webpages.
		message = "Something went wrong while crawling the webpage"
		fmt.Printf("ERROR: Crawler returned with errors from navigating %s\n", c.URL)
		fmt.Println(err.Error())
		result = CrawlResult{
			URL:         c.URL,
			CrawlStatus: crawlFail,
			Message:     message,
			TotalPages:  len(entry.IndexedWebpages),
		}

		err = sendIndexedResult(IndexedResult{
			CrawlStatus: crawlFail,
			Header:      Header{Url: c.URL, Title: entry.Title},
			Message:     message,
			Webpages:    []IndexedWebpage{},
		},
		)

		if err != nil {
			fmt.Println(err.Error())
			return CrawlResult{}, fmt.Errorf("ERROR: Unable to send error result to db.\n")
		}
		return result, nil
	}

	fmt.Printf("NOTIF: Crawler returned with no errors from navigating %s\n", c.URL)
	result = CrawlResult{
		URL:         c.URL,
		CrawlStatus: crawlSuccess,
		Message:     message,
		TotalPages:  len(entry.IndexedWebpages),
	}
	err = sendIndexedResult(IndexedResult{
		CrawlStatus: crawlFail,
		Header:      Header{Url: c.URL, Title: entry.Title},
		Message:     message,
		Webpages:    []IndexedWebpage{},
	},
	)

	if err != nil {
		fmt.Println(err.Error())
		return CrawlResult{}, fmt.Errorf("ERROR: Unable to send error result to db.\n")
	}
	return result, nil
}

/*
Returns an array of text contents from an array of common elements specified
by the current selector eg: p, a, span etc.
*/
func sendIndexedResult(result IndexedResult) error {
	conn, err := rabbitmq.GetConnection("conn")
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

	b, err := json.Marshal(result)
	if err != nil {
		log.Printf("ERROR: Unable to marshal result.\n")
		return err
	}

	// need to  check declaration and publication before returning nil
	_, err = channel.QueueDeclare(rabbitmq.CRAWLER_DB_INDEXING_QUEUE, false, false, false, false, nil)
	if err != nil {
		log.Printf("ERROR: Unable to declare %s channel.\n", rabbitmq.CRAWLER_DB_INDEXING_QUEUE)
		return err
	}
	err = channel.Publish("", rabbitmq.CRAWLER_DB_INDEXING_QUEUE, false, false, amqp.Publishing{
		Type:    "text/plain",
		Body:    []byte(b),
		ReplyTo: rabbitmq.DB_EXPRESS_INDEXING_CBQ,
	})
	if err != nil {
		log.Printf("ERROR: Unable to publish message to %s channel.\n", rabbitmq.CRAWLER_DB_INDEXING_QUEUE)
		return err
	}
	return nil
}

func (e *WebpageEntry) marshalIndexedWebpages(message string) ([]byte, error) {

	dataBuffer, err := json.Marshal(IndexedResult{
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
