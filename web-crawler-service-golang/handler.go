package main

import "context"

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

const threadPool = 4

var indexedList map[string]Webpage

func Handler(docs []string) {
	aggregateChan := make(chan webpage)
	var docIndex int // init to 0 anyways
	threadCount := threadPool

	// just to spawn threads
	for threadCount > 0 {
		if docIndex > len(docs) {
			break
		}
		go spawnCrawler(docs[docIndex], aggregateChan)
		docIndex++
		threadCount--
	}

	for data := range aggregateChan {
		save(data)
		// restore thread count
		threadCount++
	}
}

func save(w webpage) {

}

func spawnCrawler(w string, bufferChannel chan webpage) {

	/*
	 each crawler returns an array of webpages?
	 or for each webpages that is crawled, store them in to the channel?

	 latter saves memory, but more steps to process
	 steps: crawl -> index -> store -> transport to channel -> store to map

	 former keeps everything in memory, so might take too much resource,
	 fewer steps in terms of transporting to channel and storing in map,
	*/

	var ctx context.Context
	Crawl(ctx, w)
	// pushes data into the bufferChannel
}
