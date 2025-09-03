package main

import (
	frontier "crawler/frontier_queue"
	"crawler/internal/rabbitmq"
	"crawler/internal/types"
	webdriver "crawler/internal/webdriver"
	utilities "crawler/utilities"
	"encoding/json"
	"fmt"
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
	URL           string
	WD            *selenium.WebDriver
	FrontierQueue frontier.FrontierQueue
}

type DequeuedUrl struct {
	RemainingInQueue int
	Url              string
}

func NewCrawler(entryPoint string, fq frontier.FrontierQueue) (*Crawler, error) {
	c := &Crawler{
		URL:           entryPoint,
		FrontierQueue: fq,
	}
	wd, err := webdriver.CreateClient()
	if err != nil {
		fmt.Print(err.Error())
		fmt.Printf("ERROR: Unable to create a new connection with Chrome Web Driver.\n")
		return nil, err
	}
	c.WD = wd
	return c, nil
}

func SpawnCrawlers(URLs []string) {
	crawlResultsChan := make(chan types.CrawlResult, len(URLs))

	var wg sync.WaitGroup

	for _, entryPoint := range URLs {
		wg.Add(1)
		go func() {
			fq := frontier.New()
			crawler, err := NewCrawler(entryPoint, fq)
			if err != nil {
				fmt.Printf(err.Error())
				fmt.Printf("NOTIF: Thread token release due to error.\n")
				return
			}
			defer func() {
				wg.Done()
				(*crawler.WD).Quit()
			}()
			err = crawler.Crawl()
			messageStatus := CrawlMessageStatus{
				IsSuccess: true,
				Message:   "Succesfully indexed and stored webpages",
				URLSeed:   entryPoint,
			}
			if err != nil {
				messageStatus = CrawlMessageStatus{
					IsSuccess: false,
					URLSeed:   entryPoint,
					Message:   err.Error(),
				}
				fmt.Println(err.Error())
			}

			SendCrawlMessageStatus(messageStatus)
			fmt.Printf("NOTIF: Thread Token release\n")
		}()
	}

	fmt.Printf("NOTIF: Wait for crawlers\n")
	wg.Wait()
	close(crawlResultsChan)

	fmt.Println("NOTIF: All Process have finished.")
}

func (c Crawler) Crawl() error {
	defer fmt.Printf("NOTIF: Finished Crawling\n")
	defer (*c.WD).Close()

	fmt.Printf("NOTIF: Start Crawling %s\n", c.URL)

	// ROBOTS.TXT HANDLING
	hostname, _, _ := utilities.GetHostname(c.URL)
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
		Urls:            []string{}, // initialize Queue with URLSeed
		DisallowedPaths: disallowedPaths,
		IndexedWebpages: make([]types.IndexedWebpage, 0, 50),
		Hostname:        hostname,
		fq:              &c.FrontierQueue,
	}

	/*
	 to prevent duplicates if user adds a url that does not have a suffix of `/`
	 the hashmap will consider it as not the same, and we cant use strings.Contain().
	 I know its ugly.
	*/

	if c.URL[len(c.URL)-1] != '/' {
		c.URL += "/"
	}

	dqUrlChan := c.FrontierQueue.GetChann()
	go c.FrontierQueue.ListenDequeuedUrl()

	// check queue length, means that if it is > 0, then there are pending
	// nodes from the previous session, so if it > 0, we continue from
	// the current node in the queue, else  then we enqueue a new seed url

	// Visited links are already checked from the database service
	// so crawler does not have to check if the current url has already
	// been visited by it.

	queueLength, err := c.FrontierQueue.Len(hostname)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if queueLength == 0 {
		fmt.Println("CRAWLER TEST: QUEUE IS EMPTY")
		ex := frontier.ExtractedUrls{
			Domain: hostname,
			Nodes:  []string{c.URL},
		}
		fmt.Printf("CRAWLER TEST: HOSTNAME OF SEED %s\n", hostname)
		// Sends the URL seed to the frontier queue
		err = c.FrontierQueue.Enqueue(ex)
		if err != nil {
			// Error for when crawler is not able to crawl and index the seed URL.
			fmt.Printf("ERROR: unable to store Urls to database service %s\n", c.URL)
			fmt.Println(err.Error())
			return err
		}
	}

	fmt.Println("CRAWLER TEST: DEQUEUEING")

	err = c.FrontierQueue.Dequeue(hostname)
	if err != nil {
		fmt.Println(err)
		return err
	}

	for dq := range dqUrlChan {
		fmt.Printf("DEQUEUE DATA: %+v\n", dq)
		retries := 0

		if dq.RemainingInQueue == 0 {
			fmt.Println("No more urls in queue, cleaning up")
			close(dqUrlChan)
			break
		}
		// if dq.RemainingInQueue == 0 && isRoot == false {
		// 	fmt.Println("No more urls in queue, cleaning up")
		// 	close(dqUrlChan)
		// 	break
		// }

		fmt.Printf("TEST CRAWLER: PROCESSING DEQUEUED URL: %s\n", dq.Url)
		err = pageNavigator.ProcessUrl(dq.Url)
		if err != nil {
			retries++
			for retries < MAX_RETRIES {
				err = pageNavigator.ProcessUrl(dq.Url)
				if err != nil {
					fmt.Printf("ERROR: unable to naviagate to %s retrying\n", dq.Url)
					retries++
				}
			}
			fmt.Printf("ERROR: unable to naviagate to %s after %d, skipping url\n", dq.Url, retries)
		}

		// if queue is empty it should return an object with blanked url string
		// and a length of 0 and sets the current node to is_visited
		err := c.FrontierQueue.Dequeue(hostname)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	fmt.Printf("NOTIF: Crawler returned with no errors from navigating %s\n", c.URL)
	return nil
}

func SendIndexedWebpage(result types.Result) error {

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
		rabbitmq.CRAWLER_DB_INDEXING_QUEUE,
		false, false,
		amqp091.Publishing{
			ContentType: "application/json",
			Type:        "store-indexed-webpages",
			Body:        b,
			ReplyTo:     rabbitmq.DB_CRAWLER_INDEXING_CBQ,
		})
	chann.NotifyReturn(returnChan)
	select {
	case r := <-returnChan:
		fmt.Printf("ERROR: Unable to deliver message to designated queue %s\n", rabbitmq.CRAWLER_DB_INDEXING_QUEUE)
		return fmt.Errorf("ERROR: code=%d message=%s\n", r.ReplyCode, r.ReplyText)
	case <-time.After(1 * time.Second):
		fmt.Println("NOTIF: No return error from messge broker")
		return nil
	}

}
