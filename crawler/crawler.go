package main

import (
	"crawler/internal/rabbitmq"
	"crawler/internal/types"
	webdriver "crawler/internal/webdriver"
	utilities "crawler/utilities"
	"encoding/binary"
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
		fmt.Print(err.Error())
		fmt.Printf("ERROR: Unable to create a new connection with Chrome Web Driver.\n")
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

func (s *Spawner) SpawnCrawlers() {
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
			fmt.Printf("NOTIF: Thread token insert\n")
			go func() {

				defer func() {
					<-threadSlot
					wg.Done()
				}()
				crawler, err := NewCrawler(entryPoint)
				if err != nil {
					fmt.Printf(err.Error())
					fmt.Printf("NOTIF: Thread token release due to error.\n")
					return
				}
				err = crawler.Crawl()
				if err != nil {
					fmt.Println(err.Error())
					errMessageStatus := CrawlMessageStatus{
						IsSuccess: false,
						URLSeed:   entryPoint,
						Message:   err.Error(),
					}
					SendCrawlMessageStatus(errMessageStatus)
					return
				}

				(*crawler.WD).Quit()

				messageStatus := CrawlMessageStatus{
					IsSuccess: true,
					Message:   "Succesfully indexed and stored webpages",
					URLSeed:   entryPoint,
				}
				SendCrawlMessageStatus(messageStatus)
				fmt.Printf("NOTIF: Thread Token release\n")
			}()
		}
	}()

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
	}

	/*
	 to prevent duplicates if user adds a url that does not have a suffix of `/`
	 the hashmap will consider it as not the same, and we cant use strings.Contain().
	 I know its ugly.
	*/

	if c.URL[len(c.URL)-1] != '/' {
		c.URL += "/"
	}

	dqUrlChan := make(chan DequeuedUrl)
	go ListenDequeuedUrl(dqUrlChan)

	// check queue length, means that if it is > 0, then there are pending
	// nodes from the previous session, so if it > 0, we continue from
	// the current node in the queue, else  then we enqueue a new seed url

	// Visited links are already checked from the database service
	// so crawler does not have to check if the current url has already
	// been visited by it.

	queueLength, err := GetQueueLength(hostname)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if queueLength == 0 {
		fmt.Println("CRAWLER TEST: QUEUE IS EMPTY")
		ex := ExtractedUrls{
			Domain: hostname,
			Nodes:  []string{c.URL},
		}
		fmt.Printf("CRAWLER TEST: HOSTNAME OF SEED %s\n", hostname)
		// Sends the URL seed to the frontier queue
		err = EnqueueUrls(ex)
		if err != nil {
			// Error for when crawler is not able to crawl and index the seed URL.
			fmt.Printf("ERROR: unable to store Urls to database service %s\n", c.URL)
			fmt.Println(err.Error())
			return err
		}
	}

	fmt.Println("CRAWLER TEST: DEQUEUEING")
	err = DequeueUrl(hostname)
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
		err := DequeueUrl(hostname)
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

// Dequeues the url in the current queue, the domain corresponds to the
// crawler's current running job from a URL seed
func DequeueUrl(domain string) error {

	chann, err := rabbitmq.GetChannel("frontierChannel")
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = chann.Publish("",
		rabbitmq.CRAWLER_DB_DEQUEUE_URL_QUEUE,
		false, false,
		amqp091.Publishing{
			Body:    []byte(domain),
			ReplyTo: rabbitmq.DB_CRAWLER_DEQUEUE_URL_CBQ,
		})

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func ListenDequeuedUrl(dqChan chan DequeuedUrl) {
	fmt.Println("Listening to dequeued url")

	chann, err := rabbitmq.GetChannel("frontierChannel")
	if err != nil {
		fmt.Println(err)
		return
	}
	msg, err := chann.Consume(rabbitmq.DB_CRAWLER_DEQUEUE_URL_CBQ, "", false, false, false, false, nil)

	if err != nil {
		fmt.Println(err)
		return
	}

	for chanMsg := range msg {
		dq := &DequeuedUrl{}
		fmt.Println("CRAWLER TEST: received dequeued URL")
		err = json.Unmarshal(chanMsg.Body, dq)
		if err != nil {
			fmt.Println("ERROR: unable to unmarshal dequeued url")
			fmt.Println(err.Error())
			return
		}
		chann.Ack(chanMsg.DeliveryTag, false)
		dqChan <- *dq
	}
}

func EnqueueUrls(exUrls ExtractedUrls) error {

	fmt.Printf("CRAWLER TEST: ENQUEUING %+d URLS\n", len(exUrls.Nodes))
	const CRAWLER_DB_STOREURLS_QUEUE = "crawler_db_storeurls_queue"
	chann, err := rabbitmq.GetChannel("frontierChannel")
	if err != nil {
		return err
	}

	b, err := json.Marshal(exUrls)
	if err != nil {
		return err
	}
	err = chann.Publish("", CRAWLER_DB_STOREURLS_QUEUE, false, false, amqp091.Publishing{
		ContentType: "application/json",
		Body:        b,
	})
	if err != nil {
		return err
	}
	return nil
}

func GetQueueLength(hostname string) (uint32, error) {
	fmt.Printf("CRAWLER TEST: GETTING QUEUE LENGTH FOR %s\n", hostname)
	const CRAWLER_DB_GET_LEN_QUEUE = "crawler_db_len_queue"
	chann, err := rabbitmq.GetChannel("frontierChannel")
	if err != nil {
		return 0, err
	}
	chann.QueueDeclare(CRAWLER_DB_GET_LEN_QUEUE, false, false, false, false, nil)
	chann.QueueDeclare("get_queue_len_queue", false, false, false, false, nil)

	err = chann.Publish("", CRAWLER_DB_GET_LEN_QUEUE, false, false, amqp091.Publishing{
		ContentType: "application/json",
		Body:        []byte(hostname),
		ReplyTo:     "get_queue_len_queue",
	})
	if err != nil {
		return 0, err
	}

	lenMsg, err := chann.Consume("get_queue_len_queue", "", false, false, false, false, nil)

	msg := <-lenMsg

	queueLen := binary.LittleEndian.Uint32(msg.Body)

	fmt.Printf("CRAWLER TEST: BODY BUF = %v\n", msg.Body)
	fmt.Printf("CRAWLER TEST: CURRENT QUEUE LEN: %d\n", queueLen)
	chann.Ack(msg.DeliveryTag, false)

	return queueLen, nil
}
