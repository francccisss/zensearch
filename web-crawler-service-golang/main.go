package main

import (
	"encoding/json"
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
	rabbitmqclient.SetNewConnection("recevierConn", conn)
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
	crawler := Crawler{URLs: webpageIndex.Docs}
	go crawler.Start()
}

func parseIncomingData(data []byte) CrawlList {
	var webpages CrawlList
	json.Unmarshal(data, &webpages)
	return webpages
}
