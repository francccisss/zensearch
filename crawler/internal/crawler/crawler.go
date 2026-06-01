package crawler

import (
	"context"
	"crawler/internal/rabbitmq"
	"crawler/internal/types"
	webdriver "crawler/internal/webdriver"
	utilities "crawler/utilities"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
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
	FrontierQueue FrontierQueue
}

type CrawlerManager struct {
	RBQClient     *rabbitmq.RabbitMQClient
	WD            *selenium.WebDriver
	FrontierQueue FrontierQueue
}

type CrawlMessageStatus struct {
	IsSuccess bool
	Message   string
	URLSeed   string
}

func NewCrawlerManager(rbqClient *rabbitmq.RabbitMQClient, limit int) (*CrawlerManager, error) {
	wd, err := webdriver.CreateClient()
	if err != nil {
		return nil, err
	}
	cm := &CrawlerManager{
		WD:        wd,
		RBQClient: rbqClient,
	}
	return cm, nil
}

func (cr *CrawlerManager) NewCrawler(entryPoint string, fq FrontierQueue) *Crawler {
	c := &Crawler{
		URL:           entryPoint,
		FrontierQueue: fq,
		WD:            cr.WD,
	}
	return c
}

func (crm *CrawlerManager) SpawnCrawlers(ctx context.Context, URLs []string) error {

	var wg sync.WaitGroup

	crawlerResultsChan := make(chan CrawlMessageStatus, len(URLs))

	fq := crm.NewFrontierQueue()
	for _, entryPoint := range URLs {
		crawler := crm.NewCrawler(entryPoint, fq)
		go func() {
			defer func() {
				wg.Done()
				(*crawler.WD).Quit()
			}()
			err := crawler.Crawl(crm.SendIndexedWebpage)

			messageStatus := CrawlMessageStatus{
				IsSuccess: true,
				Message:   "Succesfully indexed and stored webpages",
				URLSeed:   entryPoint,
			}
			if err != nil {
				messageStatus.Message = err.Error()
				messageStatus.IsSuccess = false
			}
			crawlerResultsChan <- messageStatus
			fmt.Printf("Thread Token release\n")
		}()
	}

	for result := range crawlerResultsChan {
		err := crm.SendCrawlMessageStatus(result)
		fmt.Println(err)
	}

	fmt.Println("All Process have finished.")
	return nil
}

func (c Crawler) Crawl(SaveWebpageHandler func(types.IndexedResult) error) error {
	defer fmt.Printf("Finished Crawling\n")

	fmt.Printf("Start Crawling %s\n", c.URL)

	// ROBOTS.TXT HANDLING
	hostname, _, _ := utilities.GetHostname(c.URL)
	disallowedPaths, err := utilities.ExtractRobotsTxt(c.URL)
	if err != nil {
		fmt.Println("Unable to extract robots.txt")
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
		Hostname:        hostname,
		FQ:              &c.FrontierQueue,
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

	// Blocks thread

	// FIX: This throws away new list and instead continues on the old ones that already
	// exists in the frontier queue from a previously failed session
	queueLength, err := c.FrontierQueue.Len(hostname)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// BOOT STRAPPING FRONTIER QUEUE
	if queueLength == 0 {
		fmt.Println("CRAWLER TEST: QUEUE IS EMPTY")
		ex := ExtractedUrls{
			Root:  hostname,
			Nodes: []string{c.URL},
		}
		fmt.Printf("CRAWLER TEST: HOSTNAME OF SEED %s\n", hostname)
		// Sends the URL seed to the frontier queue
		err = c.FrontierQueue.Enqueue(ex)
		if err != nil {
			// Error for when crawler is not able to crawl and index the seed URL.
			fmt.Printf("unable to store Urls to database service %s\n", c.URL)
			fmt.Println(err.Error())
			return err
		}

		// TODO: race condition issue xd
		// need to wait for database service to receive enqueued url
		time.Sleep(time.Second * 3)
	}

	fmt.Println("CRAWLER TEST: DEQUEUEING")

	err = c.FrontierQueue.Dequeue(hostname)
	if err != nil {
		fmt.Println(err)
		return err
	}

	for dq := range dqUrlChan {
		fmt.Printf("DEQUEUE DATA: %+v\n", dq)

		if dq.RemainingInQueue == 0 {
			fmt.Println("No more urls in queue, cleaning up")
			close(dqUrlChan)
			break
		}

		fmt.Printf("TEST CRAWLER: PROCESSING DEQUEUED URL: %s\n", dq.Url)
		// TODO: Process pages concurrently
		go func() {
			retries := 0
			for retries < MAX_RETRIES {
				res, err := pageNavigator.ProcessUrl(dq.Url)
				if err != nil {
					fmt.Printf("unable to naviagate to %s retrying\n", dq.Url)
					retries++
					continue
				}
				err = SaveWebpageHandler(res)
				if err != nil {
					log.Fatal(err.Error())
				}
				return
			}
			fmt.Printf("Unable to navigate to %s after %d retries, skipping url\n", dq.Url, retries)
		}()

		// if queue is empty it should return an object with blanked url string
		// and a length of 0 and sets the current node to is_visited
		err := c.FrontierQueue.Dequeue(hostname)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	fmt.Printf("Crawler returned with no errors from navigating %s\n", c.URL)
	return nil
}

// TODO: Make SendIndexedWebpage use Go Routine instead
// ```typescript
//
//	(alias) type IndexedWebpage = {
//	    Message: string;
//	    CrawlStatus: number;
//	    Webpage: {
//	        Header: header;
//	        Contents: string;
//	    };
//	    URLSeed: string;
//	}
//
// ````
func (crm *CrawlerManager) SendIndexedWebpage(result types.IndexedResult) error {

	b, err := json.Marshal(result)
	if err != nil {
		fmt.Println("unable to marshal indexed results")
		return err
	}

	returnChan := make(chan amqp.Return)

	err = crm.RBQClient.PublishChannel.Publish("",
		rabbitmq.CRAWLER_DB_INDEXING_QUEUE,
		false, false,
		amqp.Publishing{
			ContentType: "application/json",
			Type:        "store-indexed-webpages",
			Body:        b,
			ReplyTo:     rabbitmq.DB_CRAWLER_INDEXING_CBQ,
		})
	go func() {
		del, err := crm.RBQClient.EventsChannel.Consume(rabbitmq.DB_CRAWLER_INDEXING_CBQ, "", false, false, false, false, nil)
		if err != nil {
			fmt.Printf("Error on DB_CRAWLER_INDEXING_CBQ = '%s'\n", err)
			return
		}
		msg := <-del
		err = crm.RBQClient.EventsChannel.Ack(msg.DeliveryTag, false)
		if err != nil {
			fmt.Printf("Error from ACK on DB_CRAWLER_INDEXING_CBQ = '%s'\n", err)
			return
		}
	}()
	crm.RBQClient.EventsChannel.NotifyReturn(returnChan)
	select {
	case r := <-returnChan:
		fmt.Printf("Unable to deliver message to designated queue %s\n", rabbitmq.CRAWLER_DB_INDEXING_QUEUE)
		return fmt.Errorf("code=%d message=%s\n", r.ReplyCode, r.ReplyText)
	case <-time.After(1 * time.Second):
		fmt.Println("No return error from messge broker")
		return nil
	}

}

// Send message back to express to notify that either crawl failed or was success
func (crm *CrawlerManager) SendCrawlMessageStatus(crawlStatus CrawlMessageStatus) error {

	b, err := json.Marshal(crawlStatus)
	if err != nil {
		fmt.Println("unable to marshal message status")
		return err
	}
	err = crm.RBQClient.PublishChannel.Publish("",
		rabbitmq.CRAWLER_EXPRESS_CBQ,
		false, false,
		amqp.Publishing{
			ContentType: "application/json",
			Type:        "store-indexed-webpages",
			Body:        b,
		})
	if err != nil {
		fmt.Println("Unable send crawl message status to express ")
		return err
	}
	return nil
}
