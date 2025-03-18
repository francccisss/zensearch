package types

import ()

type Header struct {
	Title string
	URL   string
}

type IndexedWebpage struct {
	Header   Header
	Contents string
}

type IndexedResult struct {
	CrawlResult
	Webpages []IndexedWebpage
}

type CrawlResult struct {
	URLSeed     string // Main entry point where the crawler starts from
	Message     string
	CrawlStatus int
	TotalPages  int
}

// type for returning the result of all spawned crawler
type CrawlResults struct {
	URLSeedCount     int // The user defined distinct urls to be crawled
	Message          string
	ThreadsUsed      int
	CrawlResultsChan chan CrawlResult
}

type Result interface {
	sendResults()
}

func (cr CrawlResult) sendResults()   {}
func (ir IndexedResult) sendResults() {}
