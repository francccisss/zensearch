package main

import (
	"context"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

const crawlQueue = "crawl_queue"

type Webpages struct {
	Webpages []Webpage
}

type Webpage struct {
	Title       string
	Contents    string
	Webpage_url string
}

func main() {

	var (
		ctx context.Context
	)

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

	aliveMainThread := make(chan struct{})
	go channelHandler(ctx, crawlChannel)
	<-aliveMainThread
}

func channelHandler(ctx context.Context, chann *amqp.Channel) {

	_, err := chann.Consume("", crawlQueue, false, false, false, false, nil)
	if err != nil {
		log.Panicf("Unable to create a crawl channel.")
	}

	for {
		// select demultiplexing inputs
		// exit go routine when some error occurs
	}

}

func parseIncomingData(data []byte) {

}
