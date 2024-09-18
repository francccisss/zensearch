package main

import (
	"context"
	"fmt"
	"sync"
)

/*
 TODO do this
 Responsible for handling crawler and webpage indexing
 x - Should handle the invocation of multple crawl jobs.
 - Handle errors from crawlers.
 - Killing and spawning crawlers.
 x - Responsible for aggregating and passing data to and from channel buffer.
 x - Makes sure that crawlers dont interleave in a context switch when passing data into the channel array buffer.
*/

type webpage struct {
	Title       string
	Contents    string
	Webpage_url string
}

type CrawlHandler struct {
	threadCount int
	docs        []string
}

const (
	threadPool = 4
)

var indexedList map[string]Webpage

func Crawler(docs []string) int {
	aggregateChan := make(chan string)
	crawler := CrawlHandler{threadCount: threadPool, docs: docs}
	var wg sync.WaitGroup
	go crawler.start(aggregateChan, &wg)
	go func() {
		wg.Wait()
		close(aggregateChan)
	}()
	for data := range aggregateChan {
		fmt.Printf("Crawled: %s", data)
	}
	return 1
}

func (c *CrawlHandler) start(aggregateChan chan string, wg *sync.WaitGroup) {
	// Semaphore
	// creates threads by Pool size
	semaphore := make(chan struct{}, c.threadCount)
	var ctx context.Context
	for _, doc := range c.docs {
		wg.Add(1)
		// blocks the loop if poolchan is full
		semaphore <- struct{}{}
		go func(doc string) {
			defer wg.Done()
			// release/decrement pool
			defer func() { <-semaphore }()
			// pass in indexed data from crawl activity
			aggregateChan <- crawl(ctx, doc)
		}(doc)
	}
}

/*
each thread of a crawler returns an array of webpages?
or for each webpages that is crawled, store them in to the channel?

latter saves memory, but more steps to process
steps: crawl -> index -> store -> transport to channel -> store to map

former keeps everything in memory, so might take too much resource,
fewer steps in terms of transporting to channel and storing in map,
*/

func crawl(ctx context.Context, w string) string {
	fmt.Printf("Start Crawling %s\n", w)
	return w
}
