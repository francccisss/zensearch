package types

import "github.com/tebeka/selenium"

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
	Webpage IndexedWebpage
}

type CrawlResult struct {
	URLSeed     string // Main entry point where the crawler starts from
	Message     string
	CrawlStatus int
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

type Crawler struct {
	URL string
	WD  *selenium.WebDriver
}

type DequeuedUrl struct {
	RemainingInQueue int
	Url              string
}

type FrontierQueue interface {
	Enqueue()
	ListenDequeuedUrl()
	Dequeue(root string)
	Len()
	GetChan() chan DequeuedUrl
}

func (cr CrawlResult) sendResults()   {}
func (ir IndexedResult) sendResults() {}
