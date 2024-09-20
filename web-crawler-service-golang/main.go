package main

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

const crawlQueue = "crawl_queue"

type IndexedList struct {
	Webpages []Webpage
}

type CrawlList struct {
	Docs []string
}

type Webpage struct {
	Title       string
	Contents    string
	Webpage_url string
}

func main() {

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
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
		log.Panicf("Unable to create a crawl channel.")
	}

	/*
	 rabbitmq library creates a new go routine for listening to new requests,
	 this function is for handling incoming messages from
	 the rabbitmq listener
	*/

	go func() {
		// body will be an array of webpages to crawl
		for msg := range delivery {
			go channelHandler(msg, crawlChannel)
		}
	}()

	aliveMainThread := make(chan struct{})
	<-aliveMainThread
}

/*
  For every message in the queue that is received and handled,
  create a new Webdriver Server, then process the users CrawlList,
  after process is finished, close down Webdriver Server

  This may add some overhead due to initializing and killing the server
  but it is insignificant when users bulk their requests.
*/

func channelHandler(msg amqp.Delivery, chann *amqp.Channel) {
	webpageIndex := parseIncomingData(msg.Body)
	defer chann.Ack(msg.DeliveryTag, false)
	go Crawler(webpageIndex.Docs)
}

func parseIncomingData(data []byte) CrawlList {
	var webpages CrawlList
	json.Unmarshal(data, &webpages)
	return webpages
}
