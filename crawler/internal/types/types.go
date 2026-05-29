package types

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

type Result interface {
	sendResults()
}

type CrawlList struct {
	Docs []string
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

type DBResponse struct {
	IsSuccess bool
	Message   string
	URLSeed   string
}
