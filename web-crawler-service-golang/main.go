package main

import (
	"context"
	"encoding/json"
	"log"
	"web-crawler-service/crawler"

	amqp "github.com/rabbitmq/amqp091-go"
)

const crawlQueue = "crawl_queue"

type IndexedList struct {
	Webpages []Webpage
}

type CrawlList struct {
	docs []string
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
	 rabbitmq library creates a new go routine for listening
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

func channelHandler(msg amqp.Delivery, chann *amqp.Channel) {
	docs := []string{"https://fzaid.vercel.app"}
	testJson, err := json.Marshal(docs)
	if err != nil {
		log.Print("Unable to encode test doc array.\n")
	}
	webpageIndex := parseIncomingData(testJson)
	chann.Ack(msg.DeliveryTag, false)
	go crawler.Handler(webpageIndex.docs)
}

func parseIncomingData(data []byte) CrawlList {
	webpages := CrawlList{}
	json.Unmarshal(data, &webpages)
	return webpages
}
