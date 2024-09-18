package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
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
	threadPool = 1
)

var indexedList map[string]Webpage

func Crawler(docs []string) int {
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

	/*
	 if threadpool is n and docs is 1
	 it can then be consumed by this for loop right after
	 the first loop is done, because since there is only one doc
	 it can then be consumed by this loop instantly while it is still processing

	 but if there are multiple docs and 1 thread,
	 the aggregateChan is full and blocks until consumed,
	 so because this is blocked, we cant release the semaphore
	 and the loop is not able to finish be case we have one we need to finish all the docs in the loop

	 need to consume the aggregateChan right after every iteration
	*/

	return 1
}

func crawl(ctx context.Context, w string) string {
	defer log.Printf("Finished Crawling\n")
	log.Printf("Start Crawling %s\n", w)
	time.Sleep(2 * time.Second)
	return w
}

func (c *CrawlHandler) start(aggregateChan chan string, wg *sync.WaitGroup) {
}

/*
each thread of a crawler returns an array of webpages?
or for each webpages that is crawled, store them in to the channel?

latter saves memory, but more steps to process
steps: crawl -> index -> store -> transport to channel -> store to map

former keeps everything in memory, so might take too much resource,
fewer steps in terms of transporting to channel and storing in map,
*/
