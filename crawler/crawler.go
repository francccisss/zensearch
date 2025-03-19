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
				crawler, err := NewCrawler(entryPoint)
				if err != nil {
					log.Print(err.Error())
					log.Printf("NOTIF: Thread token release due to error.\n")
					<-threadSlot
					return
				}
				defer func() {
					<-threadSlot
					log.Printf("NOTIF: Thread Token release\n")
				}()
				defer wg.Done()
				result, err := crawler.Crawl()
				if err != nil {
					log.Print(err.Error())
					log.Printf("NOTIF: Thread token release due to error.\n")
					return
				}
				defer (*crawler.WD).Quit()
				crawlResultsChan <- result
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
		WD:           c.WD,
		PagesVisited: map[string]string{},
		Queue: Queue{
			array: []string{c.URL}, // inialize Queue with URLSeed
		},
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

	err = pageNavigator.ProcessSeed(c.URL)

	// clean up memory resource since it will linger in memory in the heap
	// once this function is removed from the data segment.

	// TODO need to store these in the database
	// defer clear(pageNavigator.PagesVisited)
	// defer clear(pageNavigator.IndexedWebpages)

	var cResult types.CrawlResult
	message := "Successfully Crawled & Indexed website"
	if err != nil {
		// Error for when crawler is not able to crawl and index the remaining webpages.
		message = "Something went wrong while crawling the webpage"
		fmt.Printf("ERROR: Crawler returned with errors from navigating %s\n", c.URL)
		fmt.Println(err.Error())
		cResult = types.CrawlResult{
			URLSeed:     c.URL,
			CrawlStatus: CRAWL_FAIL,
			Message:     message,
		}
		return cResult, nil
	}

	fmt.Printf("NOTIF: Crawler returned with no errors from navigating %s\n", c.URL)
	cResult = types.CrawlResult{
		URLSeed:     c.URL,
		CrawlStatus: CRAWL_SUCCESS,
		Message:     message,
	}
	return cResult, nil
}

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
