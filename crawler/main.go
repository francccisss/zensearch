package main

import (
	"crawler/internal/rabbitmq"
	"encoding/json"
	"fmt"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
	"gopkg.in/yaml.v2"
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

	defBuf, err := os.ReadFile("../rabbitmq.yml")
	if err != nil {
		panic(err)
	}

	if len(defBuf) == 0 {
		panic("Empty config file")
	}

	var rbqDef rabbitmq.RabbitMQDefinitions

	err = yaml.Unmarshal(defBuf, rbqDef)
	if err != nil {
		panic(err)
	}

	crawlerDef := rabbitmq.CrawlerDefinitions{
		Exchange: rbqDef.RBExchange,
		RoutingKeys: rabbitmq.RoutingKeys{
			ES_CR_REQUEST:  rbqDef.RBRoutingKeys.ExpressServerKeys.ES_CR_REQUEST,
			CR_DB_INDEXING: rbqDef.RBRoutingKeys.CrawlerKeys.CR_DB_INDEXING,
			CR_DB_ENQUEUE:  rbqDef.RBRoutingKeys.CrawlerKeys.CR_DB_ENQUEUE,
			CR_DB_DEQUEUE:  rbqDef.RBRoutingKeys.CrawlerKeys.CR_DB_DEQUEUE,
			CR_DB_GETLEN:   rbqDef.RBRoutingKeys.CrawlerKeys.CR_DB_GETLEN,
		},
		Queues: rabbitmq.Queues{
			ES_CR_REQUEST_QUEUE:  rbqDef.RBQueues.ExpressServerQueues.ES_CR_REQUEST_QUEUE,
			ES_CR_REQUEST_CBQ:    rbqDef.RBQueues.ExpressServerQueues.ES_CR_REQUEST_CBQ,
			CR_DB_INDEXING_QUEUE: rbqDef.RBQueues.CrawlerQueues.CR_DB_INDEXING_QUEUE,
			CR_DB_INDEXING_CBQ:   rbqDef.RBQueues.CrawlerQueues.CR_DB_INDEXING_CBQ,
			CR_DB_ENQUEUE_QUEUE:  rbqDef.RBQueues.CrawlerQueues.CR_DB_ENQUEUE_QUEUE,
			CR_DB_ENQUEUE_CBQ:    rbqDef.RBQueues.CrawlerQueues.CR_DB_ENQUEUE_CBQ,
			CR_DB_DEQUEUE_QUEUE:  rbqDef.RBQueues.CrawlerQueues.CR_DB_DEQUEUE_QUEUE,
			CR_DB_DEQUEUE_CBQ:    rbqDef.RBQueues.CrawlerQueues.CR_DB_DEQUEUE_CBQ,
			CR_DB_GETLEN_QUEUE:   rbqDef.RBQueues.CrawlerQueues.CR_DB_GETLEN_QUEUE,
			CR_DB_GETLEN_CBQ:     rbqDef.RBQueues.CrawlerQueues.CR_DB_GETLEN_CBQ,
		},
	}

	client := rabbitmq.NewRabbitMQClient(crawlerDef)

	err = client.EstablishConnection(7)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("Crawler established TCP Connection with RabbitMQ")

	// expressMsg, err := expressChannel.Consume(rabbitmq.EXPRESS_CRAWLER_QUEUE, "", false, false, false, false, nil)
	// if err != nil {
	// 	log.Panicf("Unable to listen to express server")
	// }
	// for msg := range expressMsg {
	// 	// add context??
	// 	go handleIncomingUrls(msg, expressChannel)
	// }
	// log.Println("NOTIF: Crawler Exit.")
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
