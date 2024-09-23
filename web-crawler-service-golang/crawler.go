package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	webdriver "web-crawler-service-golang/pkg/webdriver"

	"github.com/tebeka/selenium"
)

type Header struct {
	Title       string
	Webpage_url string
}

type IndexedWebpage struct {
	Header
	Contents string
}

type Crawler struct {
	URLs []string
}

type CrawlTask struct {
	URL string
	ctx context.Context
	wd  *selenium.WebDriver
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

const threadPool = 10

var indexedList map[string]IndexedWebpage

const (
	crawlFail    = 0
	crawlSuccess = 1
)

type PageResult struct {
	URL         string
	Message     string
	crawlStatus int
}

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
				crawler := CrawlTask{ctx: ctx, URL: doc, wd: wd}
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
		log.Printf("CRAWLED: %+v\n", crawler)
	}

	log.Println("NOTIF: All Process have finished.")

	return Results{
		Message:     "Crawled and indexed webpages",
		ThreadsUsed: threadPool,
		URLCount:    len(c.URLs),
		PageResult:  aggregateChan,
	}
}

// 0 error
// 1 done

/*
TODO
Implement a tree search algorithm, where every link on the current visited webpage
represents a node in a tree.

Should use recursion or stack to keep track of the nodes that the crawler is traversing.

For every webpage that we successfully crawl, we need to index each page using Index
and every indexed page should be stored in a data structure / Array of type T where T is an IndexedWebpage type.

*/

func (ct CrawlTask) Crawl() (PageResult, error) {

	defer log.Printf("NOTIF: Finished Crawling\n")
	// need to close this on timeout
	err := (*ct.wd).Get(ct.URL)
	log.Printf("NOTIF: Start Crawling %s\n", ct.URL)
	if err != nil {
		return PageResult{
			URL:         ct.URL,
			crawlStatus: crawlFail,
			Message:     "Unable to establish a tcp connection with the provided URL.",
		}, err
	}
	indexer := PageIndexer{wd: ct.wd}
	indexer.Index()
	return PageResult{
		URL:         ct.URL,
		crawlStatus: crawlSuccess,
		Message:     "Crawled and Indexed."}, nil
}

func (p PageIndexer) Index() (IndexedWebpage, error) {

	/*
		Iterating through the indexSelector, where each selector, we create
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

	fmt.Printf("PAGE CONTENTS: %s\n", pageContents)
	fmt.Println("NOTIF: Page indexed")

	newIndexedPage := IndexedWebpage{
		Contents: pageContents,
		Header: Header{
			Webpage_url: url,
			Title:       title,
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
	fmt.Println(elementTextContents)

	return elementTextContents, nil
}

func joinContents(tc []string) string {
	return strings.Join(tc, " ")
}
