package main

import (
	"crawler/internal/rabbitmq"
	"crawler/internal/types"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/rabbitmq/amqp091-go"
)

func TestCrawlerNotif(t *testing.T) {

	err := rabbitmq.EstablishConnection(7)

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

	response := make(chan DBResponse)
	go ChannelListener(dbChannel, response)
	result := types.IndexedResult{
		CrawlResult: types.CrawlResult{
			URLSeed:     "fzaid.vercel.app",
			Message:     "failed to crawl URLSeed",
			CrawlStatus: CRAWL_FAIL,
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

func ChannelListener(chann *amqp091.Channel, resultChan chan DBResponse) {
	dbMsg, err := chann.Consume("", rabbitmq.DB_CRAWLER_INDEXING_NOTIF_CBQ, false, false, false, false, nil)
	if err != nil {
		panic("Unable to listen to db server")
	}
	fmt.Printf("NOTIF: listenting to %s\n", rabbitmq.DB_CRAWLER_INDEXING_NOTIF_CBQ)
	for msg := range dbMsg {
		response := &DBResponse{}
		err := json.Unmarshal(msg.Body, response)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err.Error())
			continue
		}
		fmt.Printf("NOTIF: DBResponse=%+v\n", response)
		fmt.Println("NOTIF: Notify express server")
		resultChan <- *response
		chann.Ack(msg.DeliveryTag, true)
		switch response.IsSuccess {
		case false:
			// send fail message to express server
			break
		case true:
			// send success message to express server
			break
		}
	}

}
