package main

import (
	"crawler/internal/rabbitmq"
	"crawler/internal/types"
	"fmt"
	"testing"
)

var err = rabbitmq.EstablishConnection(7)

func TestDBtoCrawlerNotif(t *testing.T) {

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
		fmt.Printf("Unable to create a crawl channel.\n")
	}
	rabbitmq.SetNewChannel("dbChannel", dbChannel)
	defer dbChannel.Close()

	dbChannel.QueueDeclare(rabbitmq.DB_CRAWLER_INDEXING_NOTIF_CBQ, false, false, false, false, nil)
	dbChannel.QueueDeclare(rabbitmq.CRAWLER_EXPRESS_CBQ, false, false, false, false, nil)

	response := make(chan DBResponse)
	go DBChannelListener(dbChannel, response)
	result := types.IndexedResult{
		CrawlResult: types.CrawlResult{
			URLSeed:     "fzaid.vercel.app",
			Message:     "Successfully indexed and stored webpages",
			CrawlStatus: CRAWL_SUCCESS,
			TotalPages:  21,
		},
		Webpages: []types.IndexedWebpage{
			{
				Header: types.Header{
					Title: "menu",
					URL:   "fzaid.vercel.app/menu",
				},
				Contents: "Doobeedobeedapdap",
			},
		},
	}
	err = SendResults(result)
	if err != nil {
		t.Fatal(err)
	}
	r := <-response
	fmt.Println(r)
}

func TestSendMessageToExpress(t *testing.T) {
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

	expressChannel, err := conn.Channel()
	if err != nil {
		fmt.Printf("Unable to create a crawl channel.\n")
	}

	rabbitmq.SetNewChannel("expressChannel", expressChannel)
	defer expressChannel.Close()

	expressChannel.QueueDeclare(rabbitmq.EXPRESS_CRAWLER_QUEUE, false, false, false, false, nil)
	expressChannel.QueueDeclare(rabbitmq.CRAWLER_EXPRESS_CBQ, false, false, false, false, nil)

	messageStatus := CrawlMessageStatus{
		IsSuccess: true,
		Message:   "Successfully crawled URLSeed", // need response directly from database
		URLSeed:   "fzaid.vercel.app",
	}
	fmt.Println("TEST: Sending Message to Express")
	err = SendCrawlMessageStatus(messageStatus, expressChannel, rabbitmq.CRAWLER_EXPRESS_CBQ)
	if err != nil {
		fmt.Printf("ERROR: Unable to send message status through %s\n", rabbitmq.CRAWLER_EXPRESS_CBQ)
		fmt.Printf("ERROR: %s", err)
		t.FailNow()
	}
}
