package crawler

import (
	"fmt"
	"sync"
	"testing"
)

type CrawlerHandler struct {
}

const Pool = 1

func TestHandler(t *testing.T) {
	docs := []string{"https://fzaid.vercel.app/"}
	handlerThreads := handler(docs)
	if handlerThreads != Pool {
		t.Errorf("Result was incorrect, got %d , want %d.", handlerThreads, Pool)
	}
}

type Semaphore struct {
	status bool
	pool   int
}

func (s *Semaphore) checkPool() {
	if s.pool == 0 {
		s.status = true
		return
	}
	s.status = false
}
func (s *Semaphore) Up() {
	s.pool++
}
func (s *Semaphore) Down() {
	s.pool--
}

func handler(docs []string) int {
	aggregateChan := make(chan string)
	var wg sync.WaitGroup
	s := Semaphore{pool: Pool}
	for _, doc := range docs {
		wg.Add(1)
		s.Down()
		go func() {
			defer wg.Done()
			spawner(doc, aggregateChan)
		}()
	}
	go func() {
		wg.Wait()
		close(aggregateChan)
	}()
	for data := range aggregateChan {
		fmt.Printf("Crawled: %s", data)
		s.Up()
	}
	return s.pool
}

func spawner(w string, bufferChannel chan string) {
	bufferChannel <- w
}
