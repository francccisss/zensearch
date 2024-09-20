package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
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
			log.Printf("Crawled: %s\n", data)
		}
	}()

	for _, doc := range docs {
		wg.Add(1)
		semaphore <- struct{}{}
		log.Printf("Semaphore token insert\n")
		go func(doc string) {
			fmt.Printf("Doc: %s\n", doc)
			defer wg.Done()
			defer func() {
				<-semaphore
				log.Printf("Semaphore token release\n")
			}()
			// blocks because no consumer so defer is not called
			aggregateChan <- crawl(ctx, doc)
		}(doc)
	}

	log.Printf("Wait for crawlers\n")
	wg.Wait()
	log.Printf("All Process has finished\n")
	close(aggregateChan)
	aggregatedData := <-aggregateChan
	log.Printf("%s\n", aggregatedData)

	return 1
}

func crawl(ctx context.Context, w string) (string, error) {
	wd, err := selenium.CreateClient()
	if err != nil {
		fmt.Errorf(err.Error())
		return "", err
	}

	defer log.Printf("Finished Crawling\n")
	log.Printf("Start Crawling %s\n", w)
	return "", nil
}
