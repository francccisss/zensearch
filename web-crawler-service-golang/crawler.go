package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	webdriver "web-crawler-service-golang/pkg"

	"github.com/tebeka/selenium"
)

type webpage struct {
	Title       string
	Contents    string
	Webpage_url string
}

const (
	threadPool = 10
)

var indexedList map[string]Webpage

func CrawlHandler(docs []string) int {

	// Start Web Driver Server
	aggregateChan := make(chan string)
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, threadPool)
	var ctx context.Context

	go func() {
		for data := range aggregateChan {
			log.Printf("CRAWLED: %s\n", data)
		}
	}()

	for _, doc := range docs {
		wg.Add(1)
		semaphore <- struct{}{}
		log.Printf("NOTIF: Semaphore token insert\n")
		go func(doc string) {
			fmt.Printf("Doc: %s\n", doc)
			defer wg.Done()
			defer func() {
				<-semaphore
				log.Printf("NOTIF: Semaphore token release\n")
			}()
			st, err := Crawl(ctx, doc)

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

func Crawl(ctx context.Context, w string) (string, error) {
	defer log.Printf("NOTIF: Finished Crawling\n")
	wd, err := webdriver.CreateClient()
	if err != nil {
		return "", err
	}
	// need to close this on timeout
	err = (*wd).Get(w)
	log.Printf("NOTIF: Start Crawling %s\n", w)
	if err != nil {
		return "", err
	}
	title, err := (*wd).Title()
	if err != nil {
		log.Printf("ERROR: No title for this page")
	}
	IndexPage(wd)
	// just return the data from crawl activity
	return title, nil
}

var indexSelector = []string{
	"a",
	"p",
	"span",
	// "code",
	// "pre",
	"h1",
	// "h2",
	// "h3",
	// "h4",
}

func IndexPage(wd *selenium.WebDriver) {
	// need to create multiple requests for each element
	elChan := make(chan []string)
	var wg sync.WaitGroup

	// extracting a single page which means, we can concat all of the strings,
	// each running on separate go routines, and return all of the concated elements, and do a final concatenation

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
					// empty string
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
		// out of range if using this, need to create a new slice to point to
		elementTextContents = append(elementTextContents, text)
		// bruh why overcomplicate things
	}
	fmt.Println(elementTextContents)
	return elementTextContents, nil
}
