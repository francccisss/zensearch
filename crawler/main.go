package main

import (
	rabbitmqclient "crawler/internal/rabbitmq"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
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

	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	log.Println("Successfully connected to rabbitmq")
	if err != nil {
		log.Panicf("Unable to establish a tcp connection with message broker.")
	}
	rabbitmqclient.SetNewConnection("receiverConn", conn)
	if err != nil {
		log.Printf("Unable to create a new connection.")
	}
	defer conn.Close()

	crawlChannel, err := conn.Channel()
	if err != nil {
		log.Printf("Unable to create a crawl channel.")
	}

	crawlChannel.QueueDeclare(crawlQueue, false, false, false, false, nil)
	delivery, err := crawlChannel.Consume("", crawlQueue, false, false, false, false, nil)

	defer crawlChannel.Close()
	if err != nil {
		log.Panicf("Unable to assert crawl message queue.")
	}
	log.Println("Crawl Channel Created")

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
