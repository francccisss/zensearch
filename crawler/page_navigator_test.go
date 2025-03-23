package main

import (
	"crawler/internal/rabbitmq"
	"fmt"
	"testing"
)

const (
	CRAWLER_DB_STOREURLS_QUEUE = "crawler_db_storeurls_queue"
	CRAWLER_DB_CLEARURLS_QUEUE = "crawler_db_clearurls_queue"

	CRAWLER_DB_DEQUEUE_URL_QUEUE = "crawler_db_dequeue_url_queue"
	DB_CRAWLER_DEQUEUE_URL_CBQ   = "db_crawler_dequeue_url_cbq"
)

func TestStoringURLQueue(t *testing.T) {

	var err = rabbitmq.EstablishConnection(7)
	if err != nil {
		t.Fatalf(err.Error())
	}

	conn, err := rabbitmq.GetConnection("conn")
	if err != nil {
		fmt.Println("Connection does not exist")
		t.Fatal(err)
	}
	fmt.Println("Crawler established TCP Connection with RabbitMQ")

	defer conn.Close()

	dbChannel, err := conn.Channel()
	if err != nil {
		fmt.Printf("Unable to create a crawl channel.\n")
		t.Fatal(err)
	}
	_, err = dbChannel.QueueDeclare(CRAWLER_DB_STOREURLS_QUEUE, false, false, false, false, nil)

	if err != nil {
		fmt.Printf("Unable to assert queue=%s\n", CRAWLER_DB_STOREURLS_QUEUE)
		t.Fatal(err)
	}
	rabbitmq.SetNewChannel("dbChannel", dbChannel)
	defer dbChannel.Close()

	s := make(chan struct{})
	seeds := []string{"https://fzaid.vercel.app"}
	spawner := NewSpawner(10, seeds)
	spawner.SpawnCrawlers()
	<-s
	fmt.Println("TEST: end test")
}
