package main

import (
	"crawler/internal/rabbitmq"
	"crawler/internal/types"
	webdriver "crawler/internal/webdriver"
	utilities "crawler/utilities"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/tebeka/selenium"
)

var elementSelector = []string{
	"a",
	"p",
	"span",
	"pre",
	"h1",
	"h2",
	"h3",
	"h4",
	"td",
	"ul",
	"code",
	"div",
}

type Spawner struct {
	ThreadPool int
	URLs       []string
}

type Crawler struct {
	URL string
	WD  *selenium.WebDriver
}

type DequeuedUrl struct {
	RemainingInQueue int
	Url              string
}

var indexedList map[string]types.IndexedWebpage

const (
	CRAWL_FAIL    = 1
	CRAWL_SUCCESS = 0
)

type ThreadToken struct{}

func NewCrawler(entryPoint string) (*Crawler, error) {
	c := &Crawler{
		URL: entryPoint,
	}
	wd, err := webdriver.CreateClient()
	if err != nil {
		log.Print(err.Error())
		log.Printf("ERROR: Unable to create a new connection with Chrome Web Driver.\n")
		return nil, err
	}
	c.WD = wd
	return c, nil
}

func NewSpawner(threadpool int, URLs []string) *Spawner {
	return &Spawner{
		ThreadPool: threadpool,
		URLs:       URLs,
	}
}

func (s *Spawner) SpawnCrawlers() types.CrawlResults {
	// Holds results of each crawled url
	crawlResultsChan := make(chan types.CrawlResult, len(s.URLs))

	// A semaphore to limit threads used
	threadSlot := make(chan struct{}, s.ThreadPool)

	// create parent context and pass to Crawl method

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, entryPoint := range s.URLs {
			wg.Add(1)
			threadSlot <- ThreadToken{}
			log.Printf("NOTIF: Thread token insert\n")
			go func() {

				defer func() {
					<-threadSlot
					wg.Done()
				}()
				crawler, err := NewCrawler(entryPoint)
				if err != nil {
					log.Print(err.Error())
					log.Printf("NOTIF: Thread token release due to error.\n")
					return
				}

				result, err := crawler.Crawl()
				if err != nil {
					log.Print(err.Error())
					return
				}
				(*crawler.WD).Quit()
				crawlResultsChan <- result
				log.Printf("NOTIF: Thread Token release\n")
			}()
		}
	}()

	log.Printf("NOTIF: Wait for crawlers\n")
	wg.Wait()
	close(crawlResultsChan)

	log.Println("NOTIF: All Process have finished.")
	return types.CrawlResults{
		Message:          "Crawled and indexed webpages",
		ThreadsUsed:      s.ThreadPool,
		URLSeedCount:     len(s.URLs),
		CrawlResultsChan: crawlResultsChan,
	}
}

func (c Crawler) Crawl() (types.CrawlResult, error) {
	defer log.Printf("NOTIF: Finished Crawling\n")
	defer (*c.WD).Close()

	log.Printf("NOTIF: Start Crawling %s\n", c.URL)

	// ROBOTS.TXT HANDLING
	hostname, _, err := utilities.GetHostname(c.URL)
	disallowedPaths, err := utilities.ExtractRobotsTxt(c.URL)
	if err != nil {
		fmt.Println("ERROR: Unable to extract robots.txt")
		fmt.Println(err.Error())
	}
	languagePaths := []string{"/es/", "/ko/", "/tr/", "/th/", "/it/", "/uk/", "/sk/", "/fr/", "/de/", "/zh/", "/ja/", "/ru/", "/ar/", "/pt/", "/hi/", "/zh/", "/zh-tw/", "/zh-c/", "/zh-cn/", "/pt-br/", "/uz/"}
	disallowedPaths = append(disallowedPaths, languagePaths...)
	fmt.Printf("DISALLOWED PATHS: %+v\n", disallowedPaths)
	// ROBOTS.TXT HANDLING

	pageNavigator := PageNavigator{
		WD:              c.WD,
		PagesVisited:    map[string]string{},
		Urls:            []string{c.URL}, // inialize Queue with URLSeed
		DisallowedPaths: disallowedPaths,
		IndexedWebpages: make([]types.IndexedWebpage, 0, 50),
		Hostname:        hostname,
	}

	/*
	 to prevent duplicates if user adds a url that does not have a suffix of `/`
	 the hashmap will consider it as not the same, and we cant use strings.Contain().
	 I know its ugly.
	*/

	if c.URL[len(c.URL)-1] != '/' {
		c.URL += "/"
	}

	// sends the urls from current page to frontier queue
	ex := ExtractedUrls{
		Domain: hostname,
		Urls:   pageNavigator.Urls,
	}

	err = storeURLs(ex)

	var cResult types.CrawlResult
	if err != nil {
		// Error for when crawler is not able to crawl and index the remaining webpages.
		fmt.Printf("ERROR: Crawler returned with errors from navigating %s\n", c.URL)
		fmt.Println(err.Error())
		cResult = types.CrawlResult{
			URLSeed:     c.URL,
			CrawlStatus: CRAWL_FAIL,
			Message:     "Something went wrong while crawling the webpage",
		}
		return cResult, nil
	}

	for {
		retries := 0
		dequeuedUrl, err := DequeueUrl()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		url := <-dequeuedUrl

		dq := &DequeuedUrl{}
		err = json.Unmarshal(url.Body, dq)

		if err != nil {
			fmt.Println(err.Error())
			break
		}

		if dq.RemainingInQueue == 0 {
			fmt.Println("No more urls in queue, cleaning up")
			break
		}

		err = pageNavigator.ProcessUrl(dq.Url)
		if err != nil {
			retries++
			for retries < MAX_RETRIES {
				err = pageNavigator.ProcessUrl(dq.Url)
				if err != nil {
					retries++
				}
			}
		}
	}

	fmt.Printf("NOTIF: Crawler returned with no errors from navigating %s\n", c.URL)
	cResult = types.CrawlResult{
		URLSeed:     c.URL,
		CrawlStatus: CRAWL_SUCCESS,
		Message:     "Successfully Crawled & Indexed website",
	}
	return cResult, nil
}

// TODO Only send results once queue is empty
func SendResults(result types.Result) error {

	chann, err := rabbitmq.GetChannel("dbChannel")
	if err != nil {
		return err
	}

	b, err := json.Marshal(result)
	if err != nil {
		fmt.Println("ERROR: unable to marshal indexed results")
		return err
	}

	returnChan := make(chan amqp091.Return)
	err = chann.Publish("",
		rabbitmq.CRAWLER_DB_INDEXING_NOTIF_QUEUE,
		false, false,
		amqp091.Publishing{
			ContentType: "application/json",
			Type:        "store-indexed-webpages",
			Body:        b,
			ReplyTo:     rabbitmq.DB_CRAWLER_INDEXING_NOTIF_CBQ,
		})
	chann.NotifyReturn(returnChan)
	select {
	case r := <-returnChan:
		fmt.Printf("ERROR: Unable to deliver message to designated queue %s\n", rabbitmq.CRAWLER_DB_INDEXING_NOTIF_QUEUE)
		return fmt.Errorf("ERROR: code=%d message=%s\n", r.ReplyCode, r.ReplyText)
	case <-time.After(1 * time.Second):
		fmt.Println("NOTIF: No return error from messge broker")
		return nil
	}

}

func DequeueUrl() (<-chan amqp091.Delivery, error) {
	const CRAWLER_DB_DEQUEUE_URL_QUEUE = "crawler_db_dequeue_url_queue"

	chann, err := rabbitmq.GetChannel("dbChannel")
	if err != nil {
		return nil, err
	}

	msg, err := chann.Consume("", CRAWLER_DB_DEQUEUE_URL_QUEUE, false, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	return msg, nil
}
