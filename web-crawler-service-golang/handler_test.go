package main

import (
	"fmt"
	"sync"
	"testing"
)

const Pool = 1

func TestHandler(t *testing.T) {
	docs := []string{"https://fzaid.vercel.app/"}
	handlerThreads := handler(docs)
	if handlerThreads != 1 {
		t.Errorf("Result was incorrect, got %d , want %d.", handlerThreads, Pool)
	}
}

func handler(docs []string) int {
	aggregateChan := make(chan string)
	var wg sync.WaitGroup
	// block loop until there is more space in the pool
	go crawler(docs, aggregateChan, Pool, &wg)
	go func() {
		wg.Wait()
		close(aggregateChan)
	}()
	for data := range aggregateChan {
		fmt.Printf("Crawled: %s", data)
	}
	return 1
}

func crawler(docs []string, aggregateChan chan string, threadCount int, wg *sync.WaitGroup) {
	threadPoolChan := make(chan struct{}, threadCount)
	go func() {
		for _, doc := range docs {
			wg.Add(1)
			// blocks the loop if poolchan is full
			threadPoolChan <- struct{}{}
			go func() {
				defer wg.Done()
				// release/decrement pool
				defer func() { <-threadPoolChan }()
				crawl(doc, aggregateChan)
			}()
		}
	}()
}

func crawl(w string, bufferChannel chan string) {
	bufferChannel <- w
}
