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

type DBResponse struct {
	IsSuccess bool
	Message   string
	Url       string
}

// TODO create type to send to express server

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

	dbChannel, err := conn.Channel()
	if err != nil {
		log.Printf("Unable to create a crawl channel.")
	}
	defer dbChannel.Close()
	expressChannel, err := conn.Channel()
	if err != nil {
		log.Printf("Unable to create a crawl channel.")
	}
	defer expressChannel.Close()

	dbChannel.QueueDeclare(rabbitmq.DB_CRAWLER_INDEXING_NOTIF_CBQ, false, false, false, false, nil)
	rabbitmq.SetNewChannel("dbChannel", dbChannel)
	go func() {
		dbMsg, err := dbChannel.Consume("", rabbitmq.DB_CRAWLER_INDEXING_NOTIF_CBQ, false, false, false, false, nil)
		if err != nil {
			log.Panicf("Unable to listen to db server")
		}
		for msg := range dbMsg {
			response := &DBResponse{}
			err := json.Unmarshal(msg.Body, response)
			if err != nil {
				fmt.Printf("ERROR: %s\n", err.Error())
				continue
			}

			dbChannel.Ack(msg.DeliveryTag, true)
			fmt.Printf("NOTIF: DBResponse=%+v\n", response)
			fmt.Println("NOTIF: Notify express server")
		}
	}()

	expressChannel.QueueDeclare(rabbitmq.EXPRESS_CRAWLER_QUEUE, false, false, false, false, nil)
	rabbitmq.SetNewChannel("expressChannel", expressChannel)
	expressMsg, err := expressChannel.Consume("", rabbitmq.EXPRESS_CRAWLER_QUEUE, false, false, false, false, nil)
	if err != nil {
		log.Panicf("Unable to listen to express server")
	}
	for msg := range expressMsg {
		go handleIncomingUrls(msg, expressChannel)
	}

	log.Println("NOTIF: Crawler Exit.")

}

func handleIncomingUrls(msg amqp.Delivery, chann *amqp.Channel) {
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
