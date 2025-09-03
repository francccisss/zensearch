package main

import (
	"crawler/internal/rabbitmq"
	"fmt"
	"testing"
)

func TestCrawlerIndexing(t *testing.T) {
	sm := make(chan struct{}, 1)

	var err = rabbitmq.EstablishConnection(7)
	// signal.Notify(osSignalChan, syscall.SIGINT, syscall.SIGTERM)

	if err != nil {
		fmt.Println(err.Error())
		t.Fatal(err)
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
		fmt.Printf("Unable to create a db channel.\n")
		t.Fatal(err)
	}
	frontierChannel, err := conn.Channel()
	if err != nil {
		fmt.Printf("Unable to create a frontier channel.\n")
		t.Fatal(err)
	}

	dbChannel.QueueDeclare(rabbitmq.DB_CRAWLER_INDEXING_CBQ, false, false, false, false, nil)

	frontierChannel.QueueDeclare("crawler_db_storeurls_queue", false, false, false, false, nil)
	frontierChannel.QueueDeclare("crawler_db_len_queue", false, false, false, false, nil)
	frontierChannel.QueueDeclare("get_queue_len_queue", false, false, false, false, nil)
	frontierChannel.QueueDeclare("db_crawler_dequeue_url_cbq", false, false, false, false, nil)

	if err != nil {
		fmt.Printf("Unable to assert queue=%s\n", "db_crawler_dequeue_url_cbq")
		t.Fatal(err)
	}

	rabbitmq.SetNewChannel("dbChannel", dbChannel)
	defer dbChannel.Close()
	rabbitmq.SetNewChannel("frontierChannel", frontierChannel)
	defer frontierChannel.Close()

	seeds := []string{"https://gobyexample.com/"}
	fmt.Printf("Crawling seeds: %+v\n", seeds)
	SpawnCrawlers(seeds)
	sm <- struct{}{}
	fmt.Println("TEST: test end")
}
