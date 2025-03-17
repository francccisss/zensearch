package main

import (
	"crawler/internal/rabbitmq"
	"encoding/json"
	"fmt"
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

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

	err := rabbitmq.EstablishConnection(7)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	conn, err := rabbitmq.GetConnection("conn")
	if err != nil {
		fmt.Println("Connection does not exist")
		os.Exit(1)
	}
	fmt.Println("Crawler established TCP Connection with RabbitMQ")

	defer conn.Close()

	crawlChannel, err := conn.Channel()
	if err != nil {
		log.Printf("Unable to create a crawl channel.")
	}

	crawlChannel.QueueDeclare(rabbitmq.CRAWLER_DB_INDEXING_QUEUE, false, false, false, false, nil)
	delivery, err := crawlChannel.Consume("", rabbitmq.CRAWLER_DB_INDEXING_QUEUE, false, false, false, false, nil)

	defer crawlChannel.Close()
	if err != nil {
		log.Panicf("Unable to assert crawl message queue.")
	}
	log.Println("Crawl Channel Created")

	go func() {
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

func parseIncomingData(data []byte) CrawlList {
	var webpages CrawlList
	json.Unmarshal(data, &webpages)
	return webpages
}
