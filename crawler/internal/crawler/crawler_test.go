package crawler

import (
	"context"
	"crawler/internal/rabbitmq"
	"fmt"
	"os"
	"testing"
	"time"

	"gopkg.in/yaml.v2"
)

var CRAWL_QUERY = []string{"https://zenread.pro/"}

func TestCrawler(t *testing.T) {

	client := MockConnection(t)
	// ephemeral queues
	defer func() {
		client.PublishChannel.QueueDelete(client.Definitions.Queues.ES_CR_REQUEST_CBQ, false, false, true)
		client.PublishChannel.QueueDelete(client.Definitions.Queues.CR_DB_INDEXING_CBQ, false, false, true)
		client.PublishChannel.QueueDelete(client.Definitions.Queues.CR_DB_ENQUEUE_CBQ, false, false, true)
		client.PublishChannel.QueueDelete(client.Definitions.Queues.CR_DB_DEQUEUE_CBQ, false, false, true)
		client.PublishChannel.QueueDelete(client.Definitions.Queues.CR_DB_GETLEN_CBQ, false, false, true)
		client.Connection.Close()
	}()
	cm, err := NewCrawlerManager(client, len(CRAWL_QUERY))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), time.Second*3)
	defer cancel()
	err = cm.SpawnCrawlers(CRAWL_QUERY)
	if err != nil {
		t.Fatal(err)
	}
	err = cm.Crawl(ctx)
	if err != nil {
		t.Fatal(err)
	}

}

func MockConnection(t *testing.T) *rabbitmq.RabbitMQClient {
	defBuf, err := os.ReadFile("../../../rabbitmq.yml")
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
			ES_CR_REQUEST_QUEUE: rbqDef.RBQueues.ExpressServerQueues.ES_CR_REQUEST_QUEUE,
			ES_CR_REQUEST_CBQ:   rbqDef.RBQueues.ExpressServerQueues.ES_CR_REQUEST_CBQ,

			CR_DB_INDEXING_QUEUE: rbqDef.RBQueues.CrawlerQueues.CR_DB_INDEXING_QUEUE,
			CR_DB_INDEXING_CBQ:   rbqDef.RBQueues.CrawlerQueues.CR_DB_INDEXING_CBQ,

			CR_DB_ENQUEUE_QUEUE: rbqDef.RBQueues.CrawlerQueues.CR_DB_ENQUEUE_QUEUE,
			CR_DB_ENQUEUE_CBQ:   rbqDef.RBQueues.CrawlerQueues.CR_DB_ENQUEUE_CBQ,

			CR_DB_DEQUEUE_QUEUE: rbqDef.RBQueues.CrawlerQueues.CR_DB_DEQUEUE_QUEUE,
			CR_DB_DEQUEUE_CBQ:   rbqDef.RBQueues.CrawlerQueues.CR_DB_DEQUEUE_CBQ,

			CR_DB_GETLEN_QUEUE: rbqDef.RBQueues.CrawlerQueues.CR_DB_GETLEN_QUEUE,
			CR_DB_GETLEN_CBQ:   rbqDef.RBQueues.CrawlerQueues.CR_DB_GETLEN_CBQ,
		},
	}

	fmt.Println(crawlerDef.Queues.CR_DB_INDEXING_QUEUE)
	client := rabbitmq.NewRabbitMQClient(crawlerDef)

	err = client.EstablishConnection(7)

	if err != nil {
		t.Fatal(err.Error())
	}

	err = client.SetDefinitions()
	if err != nil {
		t.Fatal(err.Error())
	}
	fmt.Println("Crawler established TCP Connection with RabbitMQ")
	return &client
}
