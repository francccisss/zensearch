package main

import (
	"crawler/internal/rabbitmq"
	"encoding/json"
	"fmt"
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

type CrawlList struct {
	Docs []string
}

type CrawlMessageStatus struct {
	IsSuccess bool
	Message   string
	URLSeed   string
}

type DBResponse struct {
	IsSuccess bool
	Message   string
	URLSeed   string
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
		log.Printf("Unable to create a db channel.")
	}
	defer dbChannel.Close()
	expressChannel, err := conn.Channel()
	if err != nil {
		log.Printf("Unable to create a express channel.")
	}
	defer expressChannel.Close()

	frontierChannel, err := conn.Channel()
	if err != nil {
		log.Printf("Unable to create a express channel.")
	}

	dbChannel.QueueDeclare(rabbitmq.CRAWLER_DB_INDEXING_QUEUE, false, false, false, false, nil)
	dbChannel.QueueDeclare(rabbitmq.DB_CRAWLER_INDEXING_CBQ, false, false, false, false, nil)

	rabbitmq.SetNewChannel("dbChannel", dbChannel)

	expressChannel.QueueDeclare(rabbitmq.EXPRESS_CRAWLER_QUEUE, false, false, false, false, nil)
	expressChannel.QueueDeclare(rabbitmq.CRAWLER_EXPRESS_CBQ, false, false, false, false, nil)
	rabbitmq.SetNewChannel("expressChannel", expressChannel)

	frontierChannel.QueueDeclare("crawler_db_storeurls_queue", false, false, false, false, nil)
	frontierChannel.QueueDeclare("db_crawler_dequeue_url_cbq", false, false, false, false, nil)

	frontierChannel.QueueDeclare("crawler_db_len_queue", false, false, false, false, nil)
	frontierChannel.QueueDeclare("get_queue_len_queue", false, false, false, false, nil)

	rabbitmq.SetNewChannel("frontierChannel", frontierChannel)

	expressMsg, err := expressChannel.Consume(rabbitmq.EXPRESS_CRAWLER_QUEUE, "", false, false, false, false, nil)
	if err != nil {
		log.Panicf("Unable to listen to express server")
	}
	for msg := range expressMsg {
		// add context??
		go handleIncomingUrls(msg, expressChannel)
	}
	log.Println("NOTIF: Crawler Exit.")
}

func handleIncomingUrls(msg amqp.Delivery, chann *amqp.Channel) {
	defer chann.Ack(msg.DeliveryTag, false)
	webpageIndex := parseIncomingData(msg.Body)
	fmt.Printf("Docs: %+v\n", webpageIndex.Docs)
	go SpawnCrawlers(webpageIndex.Docs)
}

func parseIncomingData(data []byte) CrawlList {
	var webpages CrawlList
	json.Unmarshal(data, &webpages)
	return webpages
}

// Send message back to express to notify that either crawl failed or was success
func SendCrawlMessageStatus(crawlStatus CrawlMessageStatus) error {

	expressChannel, err := rabbitmq.GetChannel("expressChannel")
	b, err := json.Marshal(crawlStatus)
	if err != nil {
		fmt.Println("ERROR: unable to marshal message status")
		return err
	}
	err = expressChannel.Publish("",
		rabbitmq.CRAWLER_EXPRESS_CBQ,
		false, false,
		amqp.Publishing{
			ContentType: "application/json",
			Type:        "store-indexed-webpages",
			Body:        b,
		})
	if err != nil {
		fmt.Println("ERROR: Unable send crawl message status to express ")
		return err
	}
	return nil
}
