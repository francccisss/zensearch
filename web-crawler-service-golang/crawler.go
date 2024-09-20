package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"web-crawler-service-golang/pkg/selenium"
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

func Crawler(docs []string) int {

	// Start Web Driver Server
	selenium.CreateWebDriverServer()
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
			st, err := crawl(ctx, doc)

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

func crawl(ctx context.Context, w string) (string, error) {
	defer log.Printf("NOTIF: Finished Crawling\n")

	wd, err := selenium.CreateClient()
	if err != nil {
		return "", err
	}
	err = (*wd).Get(w)
	log.Printf("NOTIF: Start Crawling %s\n", w)
	if err != nil {
		return "", err
	}
	title, err := (*wd).Title()
	if err != nil {
		log.Printf("ERROR: No title for this page")
	}
	// just return the data from crawl activity
	return title, nil
}
