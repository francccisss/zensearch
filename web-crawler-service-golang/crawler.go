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
type webpage struct {
	Header
	Contents string
}

type Crawler struct {
	URLs []string
}

type CrawlTask struct {
	URL string
	ctx context.Context
}

const threadPool = 10

var indexedList map[string]Webpage

func (c Crawler) Start() int {
	// Start Web Driver Server
	aggregateChan := make(chan string)
	semaphore := make(chan struct{}, threadPool)
	var (
		wg  sync.WaitGroup
		ctx context.Context
	)

	go func() {
		for data := range aggregateChan {
			log.Printf("CRAWLED: %s\n", data)
		}
	}()

	// Why do we limit go routines to save resource by throttling threads?
	// maybe we can let users control how many threads to crawl list of webpages

	for _, doc := range c.URLs {
		wg.Add(1)
		semaphore <- struct{}{}
		log.Printf("NOTIF: Semaphore token insert\n")
		go func(doc string) {
			fmt.Printf("Doc: %s\n", doc)
			crawler := CrawlTask{ctx: ctx, URL: doc}
			defer wg.Done()
			defer func() {
				<-semaphore
				log.Printf("NOTIF: Semaphore token release\n")
			}()
			st, err := crawler.Crawl()

			// need to get out if an error occurs
			if err != nil {
				defer func() {
					<-semaphore
				}()
				defer wg.Done()
				log.Print(err.Error())
				log.Printf("NOTIF: Semaphore token release due to error.\n")
			}

			aggregateChan <- st
		}(doc)
	}

	log.Printf("NOTIF: Wait for crawlers\n")
	wg.Wait()
	log.Printf("NOTIF: All Process has finished\n")
	close(aggregateChan)
	<-aggregateChan

	return 1
}

func (ct CrawlTask) Crawl() (string, error) {
	defer log.Printf("NOTIF: Finished Crawling\n")
	wd, err := webdriver.CreateClient()
	if err != nil {
		return "", err
	}
	// need to close this on timeout
	err = (*wd).Get(ct.URL)
	log.Printf("NOTIF: Start Crawling %s\n", ct.URL)
	if err != nil {
		return "", err
	}
	title, err := (*wd).Title()
	if err != nil {
		log.Printf("ERROR: No title for this page")
	}
	Index(wd)
	// just return the data from crawl activity
	return title, nil
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

func Index(wd *selenium.WebDriver) {

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
				textContents, err := textContentByIndexSelector(wd, selector)
				if err != nil {
					textContentChan <- ""
					/*
					   Idiomatic way is to call defer beforehand,
					   once this error checker is true, we can return
					   and defer function is called wg.Done() in particular
					*/
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
	for elementContents := range textContentChan {
		textChanSlice = append(textChanSlice, elementContents)
	}
	pageContents := joinContents(textChanSlice)
	fmt.Printf("PAGE CONTENTS: %s\n", pageContents)
	fmt.Println("NOTIF: Page indexed")
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
