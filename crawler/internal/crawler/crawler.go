package crawler

import (
	"context"
	"crawler/internal/rabbitmq"
	"crawler/internal/types"
	webdriver "crawler/internal/webdriver"
	utilities "crawler/utilities"
	"encoding/json"
	"fmt"
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

type crawler struct {
	URL           string
	FrontierQueue FrontierQueue
	PageNavigator *PageNavigator
}

type CrawlerManager struct {
	RBQClient        *rabbitmq.RabbitMQClient
	WD               *selenium.WebDriver
	FrontierQueue    FrontierQueue
	CrawlerList      []*crawler
	ConsumerChannels map[string]<-chan amqp.Delivery
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
		WD:          wd,
		RBQClient:   rbqClient,
		CrawlerList: make([]*crawler, 0, limit),
	}
	return cm, nil
}

func (crm *CrawlerManager) newCrawler(entryPoint string) *crawler {
	fq := crm.NewFrontierQueue()
	fmt.Printf("Start Crawling %s\n", entryPoint)

	// ROBOTS.TXT HANDLING
	hostname, _, _ := utilities.GetHostname(entryPoint)
	disallowedPaths, err := utilities.ExtractRobotsTxt(entryPoint)
	if err != nil {
		fmt.Println("Unable to extract robots.txt")
		fmt.Println(err.Error())
	}
	languagePaths := []string{"/es/", "/ko/", "/tr/", "/th/", "/it/", "/uk/", "/sk/", "/fr/", "/de/", "/zh/", "/ja/", "/ru/", "/ar/", "/pt/", "/hi/", "/zh/", "/zh-tw/", "/zh-c/", "/zh-cn/", "/pt-br/", "/uz/"}
	disallowedPaths = append(disallowedPaths, languagePaths...)
	fmt.Printf("DISALLOWED PATHS: %+v\n", disallowedPaths)
	// ROBOTS.TXT HANDLING

	pg := PageNavigator{
		WD:              crm.WD,
		DisallowedPaths: disallowedPaths,
		Hostname:        hostname,
		FQ:              &fq,
	}

	/*
	 to prevent duplicates if user adds a url that does not have a suffix of `/`
	 the hashmap will consider it as not the same, and we cant use strings.Contain().
	 I know its ugly.
	*/

	c := &crawler{
		URL:           entryPoint,
		FrontierQueue: fq,
		PageNavigator: &pg,
	}
	if c.URL[len(c.URL)-1] != '/' {
		c.URL += "/"
	}

	return c
}

// give each crawler their own go routine from a thread pool
func (crm *CrawlerManager) SpawnCrawlers(URLs []string) error {

	for _, entryPoint := range URLs {
		crawler := crm.newCrawler(entryPoint)
		go crawler.FrontierQueue.ListenDequeuedUrl()
		crm.CrawlerList = append(crm.CrawlerList, crawler)
	}
	fmt.Printf("%d Crawlers Created\n", len(URLs))
	return nil

}

func (crm *CrawlerManager) Crawl(ctx context.Context) error {

	var wg sync.WaitGroup
	defer (*crm.WD).Quit()

	crawlerResultsChan := make(chan CrawlMessageStatus, len(crm.CrawlerList))
	for i := range crm.CrawlerList {

		wg.Go(func() {
			crawler := crm.CrawlerList[i]
			dqUrlChan := crawler.FrontierQueue.GetChann()

			messageStatus := CrawlMessageStatus{
				IsSuccess: true,
				Message:   "Succesfully indexed and stored webpages",
				URLSeed:   crawler.URL,
			}

			queueLength, err := crawler.FrontierQueue.Len(crawler.PageNavigator.Hostname)
			if err != nil {
				fmt.Println(err)
				messageStatus.IsSuccess = false
				messageStatus.Message = err.Error()
				crawlerResultsChan <- messageStatus
				return
			}

			fmt.Printf("Pre-fill queue len: %d\n", queueLength)
			// BOOT STRAPPING FRONTIER QUEUE
			if queueLength == 0 {
				// Sends the URL seed to the frontier queue
				fmt.Println("No Links to continue")
				dqBootStrap := DequeuedUrl{
					RemainingInQueue: 0,
					Url:              crawler.URL,
				}

				fmt.Println("Boot Strapping current url")
				res, err := crawler.crawl(dqBootStrap)
				if err != nil {
					fmt.Println(err)
					messageStatus.IsSuccess = false
					messageStatus.Message = err.Error()
					crawlerResultsChan <- messageStatus
					return
				}

				err = crm.SendIndexedWebpage(res)
				if err != nil {
					fmt.Println(err)
					messageStatus.IsSuccess = false
					messageStatus.Message = err.Error()
					crawlerResultsChan <- messageStatus
					return
				}
			}

			err = crawler.FrontierQueue.Dequeue(crawler.PageNavigator.Hostname)
			if err != nil {
				fmt.Println(err)
				messageStatus.IsSuccess = false
				messageStatus.Message = err.Error()
				crawlerResultsChan <- messageStatus
				return
			}

			for dq := range dqUrlChan {

				if dq.RemainingInQueue == 0 {
					fmt.Println("No more urls in queue, cleaning up")
					close(dqUrlChan)
					break
				}

				res, err := crawler.crawl(dq)
				if err != nil {
					fmt.Println(err)
					continue
				}

				fmt.Println("WILL BE INDEXING")
				err = crm.SendIndexedWebpage(res)
				fmt.Println("Sent Indexed webpage to database service.")
				if err != nil {
					fmt.Println(err)
					messageStatus.IsSuccess = false
					messageStatus.Message = err.Error()
					crawlerResultsChan <- messageStatus
					return
				}

				err = crawler.FrontierQueue.Dequeue(crawler.PageNavigator.Hostname)

				if err != nil {
					fmt.Println(err)
					messageStatus.IsSuccess = false
					messageStatus.Message = err.Error()
					crawlerResultsChan <- messageStatus
					return
				}

			}
			crawlerResultsChan <- messageStatus
			fmt.Printf("Thread Token release\n")
		})

	}
	wg.Wait()

	close(crawlerResultsChan)
	for result := range crawlerResultsChan {
		fmt.Println(result)
		// err := crm.SendCrawlMessageStatus(result)
		// fmt.Println(err)
	}
	return nil
}

// TODO: Make this more faster by dequeuing N urls
// and processing them using go routines, then waiting for
// each routine to finish using wait groups, to then process
// the next batch of urls
func (c crawler) crawl(dq DequeuedUrl) (types.IndexedResult, error) {
	defer fmt.Printf("Finished Crawling\n")

	fmt.Printf("DEQUEUE DATA: %+v\n", dq)

	fmt.Printf("TEST CRAWLER: PROCESSING DEQUEUED URL: %s\n", dq.Url)
	retries := 0
	for retries < MAX_RETRIES {
		res, err := c.PageNavigator.ProcessUrl(dq.Url)
		if err != nil {
			fmt.Printf("unable to naviagate to %s retrying\n", dq.Url)
			retries++
			continue
		}
		fmt.Printf("Crawler returned with no errors from navigating %s\n", c.URL)
		return res, nil
	}

	return types.IndexedResult{}, fmt.Errorf("Unable to fetch process page from %s", c.URL)
}

func (crm *CrawlerManager) SendIndexedWebpage(result types.IndexedResult) error {

	fmt.Println("AM I SENDING ANYTHING")
	b, err := json.Marshal(result)
	if err != nil {
		fmt.Println("unable to marshal indexed results")
		return err
	}

	returnChan := make(chan amqp.Return)

	err = crm.RBQClient.PublishChannel.Publish(crm.RBQClient.Definitions.Exchange.Crawler,
		crm.RBQClient.Definitions.RoutingKeys.CR_DB_INDEXING,
		false, false,
		amqp.Publishing{
			ContentType: "application/json",
			ReplyTo:     crm.RBQClient.Definitions.Queues.CR_DB_INDEXING_CBQ,
			Body:        b,
		})
	if err != nil {
		return err
	}
	crm.RBQClient.EventsChannel.NotifyReturn(returnChan)
	select {
	case r := <-returnChan:
		fmt.Printf("Unable to deliver message to designated queue %s\n", crm.RBQClient.Definitions.Queues.CR_DB_INDEXING_CBQ)
		return fmt.Errorf("code=%d message=%s\n", r.ReplyCode, r.ReplyText)
	case <-time.After(1 * time.Second):
		fmt.Println("No error from messge broker")
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
	err = crm.RBQClient.PublishChannel.Publish(crm.RBQClient.Definitions.Exchange.General,
		crm.RBQClient.Definitions.Queues.ES_CR_REQUEST_CBQ,
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
