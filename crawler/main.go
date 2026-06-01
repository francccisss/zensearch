package main

import (
	"context"
	"crawler/internal/crawler"
	"crawler/internal/rabbitmq"
	"crawler/internal/types"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

func main() {

	defBuf, err := os.ReadFile("../rabbitmq.yml")
	if err != nil {
		panic(err)
	}

	if len(defBuf) == 0 {
		panic("Empty config file")
	}

	var rbqDef rabbitmq.RabbitMQDefinitions

	err = yaml.Unmarshal(defBuf, &rbqDef)
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

	err = client.SetDefinitions()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Println("Crawler established TCP Connection with RabbitMQ")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		expressMsg, err := client.EventsChannel.Consume(client.Definitions.Queues.ES_CR_REQUEST_QUEUE, "", false, false, false, false, nil)
		if err != nil {
			log.Println("Unable to listen to express server")
			break
		}
		msg := <-expressMsg
		var list types.CrawlList
		err = json.Unmarshal(msg.Body, &list)
		if err != nil {
			log.Panic(err)
		}
		client.EventsChannel.Ack(msg.DeliveryTag, false)
		go func(ctx context.Context) {

			fmt.Printf("Docs: %+v\n", list.Docs)
			crawlerManager, err := crawler.NewCrawlerManager(&client, len(list.Docs))
			if err != nil {
				log.Fatal(err)
			}
			err = crawlerManager.SpawnCrawlers(ctx, list.Docs)
			// TODO: Need to make the crawler send an error crawl message aside from individual ones
			// could use a different queue that the express server to consume from and handle
			if err != nil {
				panic(err)
			}

		}(ctx)
	}
	log.Println("NOTIF: Crawler Exit.")
}
