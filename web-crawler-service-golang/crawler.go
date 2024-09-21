package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	webdriver "web-crawler-service-golang/pkg"

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

type CrawlHandler struct {
	URLs []string
}

type Crawler struct {
	URL string
	ctx context.Context
}

const threadPool = 10

var indexedList map[string]Webpage

func (c CrawlHandler) CrawlHandler() int {
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

	for _, doc := range c.URLs {
		wg.Add(1)
		semaphore <- struct{}{}
		log.Printf("NOTIF: Semaphore token insert\n")
		go func(doc string) {
			fmt.Printf("Doc: %s\n", doc)
			crawler := Crawler{ctx: ctx, URL: doc}
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

func (c Crawler) Crawl() (string, error) {
	defer log.Printf("NOTIF: Finished Crawling\n")
	wd, err := webdriver.CreateClient()
	if err != nil {
		return "", err
	}
	// need to close this on timeout
	err = (*wd).Get(c.URL)
	log.Printf("NOTIF: Start Crawling %s\n", c.URL)
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
	elChan := make(chan []string)
	var wg sync.WaitGroup
	go func() {
		for elementContents := range elChan {
			fmt.Println(elementContents)
		}
	}()
	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
			log.Println("NOTIF: Indexing Done.")
		}()
		for _, selector := range indexSelector {
			wg.Add(1)
			go func(selector string) {
				defer wg.Done()
				contents, err := extractElements(wd, selector)
				defer fmt.Printf("NOTIF: Selector: %s done\n", selector)
				if err != nil {
					elChan <- []string{}
					wg.Done()
				}
				elChan <- contents
			}(selector)
		}
	}()

	fmt.Println("NOTIF: Waiting for page indexer")
	wg.Wait()
	close(elChan)
	fmt.Println("NOTIF: Page indexed")
}

func extractElements(wd *selenium.WebDriver, selector string) ([]string, error) {
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

func joinContents(wc []string) (string, error) {
	return "", nil
}
