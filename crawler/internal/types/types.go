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

type CrawlResult struct {
	URL         string
	Message     string
	CrawlStatus int
	TotalPages  int
}

type IndexedResult struct {
	Webpages []IndexedWebpage
	Header
	Message     string
	CrawlStatus int
}

type ErrorMessage struct {
	Message     string
	URL         string
	CrawlStatus int
}

type Results struct {
	URLCount    int
	URLsFailed  []string
	Message     string
	ThreadsUsed int
	CrawlResult <-chan CrawlResult
}
