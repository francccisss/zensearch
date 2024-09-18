package main

import (
	"fmt"
	"sync"
	"testing"
)

const Pool = 1

func TestHandler(t *testing.T) {
	docs := []string{"https://fzaid.vercel.app/"}
	handlerThreads := Testhandler(docs)
	if handlerThreads != 1 {
		t.Errorf("Result was incorrect, got %d , want %d.", handlerThreads, Pool)
	}
}

func Testhandler(docs []string) int {
	aggregateChan := make(chan string)
	var wg sync.WaitGroup
	// block loop until there is more space in the pool

	go Testcrawler(docs, aggregateChan, Pool, &wg) // creates threads by Pool size
	go func() {
		wg.Wait()
		close(aggregateChan)
	}()
	for data := range aggregateChan {
		fmt.Printf("Crawled: %s", data)
	}
	return 1
}

func Testcrawler(docs []string, aggregateChan chan string, threadCount int, wg *sync.WaitGroup) {
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
				Testcrawl(doc, aggregateChan)
			}()
		}
	}()
}

func Testcrawl(w string, bufferChannel chan string) {
	bufferChannel <- w
}
