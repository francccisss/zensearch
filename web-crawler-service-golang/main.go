package main

import (
	"encoding/json"
	"fmt"
	"log"
	rabbitmqclient "web-crawler-service-golang/pkg/rabbitmq_client"
	webdriver "web-crawler-service-golang/pkg/webdriver"

	amqp "github.com/rabbitmq/amqp091-go"
)

const crawlQueue = "crawl_queue"

type IndexedList struct {
	Webpages []site
}

type CrawlList struct {
	Docs []string
}

type site struct {
	Title       string
	Contents    string
	Webpage_url string
}

func main() {

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Panicf("Unable to establish a tcp connection with message broker.")
	}
	rabbitmqclient.SetNewConnection("receiverConn", conn)
	if err != nil {
		log.Panicf("Unable to establish a tcp connection with message broker.")
	}
	defer conn.Close()

	crawlChannel, err := conn.Channel()
	if err != nil {
		log.Panicf("Unable to create a crawl channel.")
	}

	crawlChannel.QueueDeclare(crawlQueue, false, false, false, false, nil)
	delivery, err := crawlChannel.Consume("", crawlQueue, false, false, false, false, nil)

	defer crawlChannel.Close()
	if err != nil {
		log.Panicf("Unable to assert crawl message queue.")
	}

	service, err := webdriver.CreateWebDriverServer()
	defer (*service).Stop()
	if err != nil {
		log.Print("INFO: Retry web driver server or the application.\n")
		log.Print(err.Error())
	}

	/*
	 rabbitmq library creates a new go routine for listening to new requests,
	 this function is for handling incoming messages from
	 the rabbitmq listener
	*/

	// TODO create a push queue back to the client to notify
	// that crawling is done, after indexing pages maybe within the saveIndexedWebpages()

	go func() {
		// body will be an array of webpages to crawl
		for msg := range delivery {
			go handleConnections(msg, crawlChannel)
		}
	}()

	aliveMainThread := make(chan struct{})
	<-aliveMainThread

	log.Println("NOTIF: Crawler Exit.")

}

/*
  For every message in the queue that is received and handled,
  create a new Webdriver Server, then process the users CrawlList,
  after process is finished, close down Webdriver Server

  This may add some overhead due to initializing and killing the server
  but it is insignificant when users bulk their requests.
*/

func handleConnections(msg amqp.Delivery, chann *amqp.Channel) {
	defer chann.Ack(msg.DeliveryTag, false)
	webpageIndex := parseIncomingData(msg.Body)
	fmt.Printf("Docs: %+v\n", webpageIndex.Docs)
	spawner := NewSpawner(10, webpageIndex.Docs)
	go spawner.SpawnCrawlers()
}
func NewSpawner(threadpool int, URLs []string) *Spawner {
	return &Spawner{
		threadPool: threadpool,
		URLs:       URLs,
	}
}

func parseIncomingData(data []byte) CrawlList {
	var webpages CrawlList
	json.Unmarshal(data, &webpages)
	return webpages
}
