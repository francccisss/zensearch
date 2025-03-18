package main

import (
	"crawler/internal/rabbitmq"
	"crawler/internal/types"
	"encoding/json"
	"fmt"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
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
	dbChannel.QueueDeclare(rabbitmq.DB_CRAWLER_INDEXING_NOTIF_CBQ, false, false, false, false, nil)
	rabbitmq.SetNewChannel("dbChannel", dbChannel)
	defer dbChannel.Close()

	response := make(chan DBResponse)
	go DBMockChannelListener(dbChannel, response)
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
	fmt.Println("TEST: Waiting for response")
	r := <-response
	fmt.Println(r)

}

// func TestSendMessageToExpress(t *testing.T) {
// 	URLSeeds := []struct {
// 		seed    string
// 		msg     string
// 		success bool
// 	}{
// 		{
// 			seed:    "https://fasdas",
// 			msg:     "Something went wrong, unable to crawl URLSeed",
// 			success: false,
// 		},
// 		{
// 			seed:    "https://fzaid.vercel.app",
// 			msg:     "Something went wrong, unable to crawl URLSeed",
// 			success: false,
// 		},
// 	}
// 	if err != nil {
// 		fmt.Println(err.Error())
// 		t.Fatal(err)
// 	}
// 	conn, err := rabbitmq.GetConnection("conn")
// 	if err != nil {
// 		fmt.Println("Connection does not exist")
// 		t.Fatal(err)
// 	}
// 	fmt.Println("Crawler established TCP Connection with RabbitMQ")
//
// 	defer conn.Close()
//
// 	expressChannel, err := conn.Channel()
// 	if err != nil {
// 		fmt.Printf("Unable to create a crawl channel.\n")
// 	}
//
// 	rabbitmq.SetNewChannel("expressChannel", expressChannel)
// 	defer expressChannel.Close()
//
// 	expressChannel.QueueDeclare(rabbitmq.EXPRESS_CRAWLER_QUEUE, false, false, false, false, nil)
// 	expressChannel.QueueDeclare(rabbitmq.CRAWLER_EXPRESS_CBQ, false, false, false, false, nil)
//
// 	var wg sync.WaitGroup
// 	for _, url := range URLSeeds {
// 		wg.Add(1)
// 		go func() {
//
// 			defer wg.Done()
// 			messageStatus := CrawlMessageStatus{
// 				IsSuccess: url.success,
// 				Message:   url.msg, // need response directly from database
// 				// if testing make sure URLSeed matches the crawled url on the client
// 				URLSeed: url.seed,
// 			}
// 			fmt.Println("TEST: Sending Message to Express")
// 			err = SendCrawlMessageStatus(messageStatus, expressChannel, rabbitmq.CRAWLER_EXPRESS_CBQ)
// 			if err != nil {
// 				fmt.Printf("ERROR: Unable to send message status through %s\n", rabbitmq.CRAWLER_EXPRESS_CBQ)
// 				fmt.Printf("ERROR: %s", err)
// 				panic(err)
// 			}
// 		}()
// 	}
// 	wg.Wait()
// 	fmt.Println("TEST: Done")
// }

func DBMockChannelListener(chann *amqp.Channel, resultChan chan DBResponse) {
	fmt.Println("TEST: DB CHANNEL LISTENER")

	dbMsg, err := chann.Consume("", rabbitmq.DB_CRAWLER_INDEXING_NOTIF_CBQ, false, false, false, false, nil)
	if err != nil {
		panic("Unable to listen to db server")
	}
	fmt.Printf("NOTIF: listenting to %s\n", rabbitmq.DB_CRAWLER_INDEXING_NOTIF_CBQ)
	for msg := range dbMsg {

		response := &DBResponse{}
		err := json.Unmarshal(msg.Body, response)
		fmt.Println("TEST: Received DB Response")
		if err != nil {
			fmt.Printf("ERROR: %s\n", err.Error())
			chann.Nack(msg.DeliveryTag, false, true)
			continue
		}
		chann.Ack(msg.DeliveryTag, false)
		fmt.Printf("NOTIF: DBResponse=%+v\n", response)
		resultChan <- *response

		fmt.Println("NOTIF: Notify express server")
		switch response.IsSuccess {
		case false:
			// send fail message to express server when error
			// storing webpages on database service
			messageStatus := CrawlMessageStatus{
				IsSuccess: response.IsSuccess,
				Message:   response.Message, // need response directly from database
				URLSeed:   response.URLSeed,
			}
			fmt.Println(messageStatus)
			break
		case true:
			messageStatus := CrawlMessageStatus{
				IsSuccess: response.IsSuccess,
				Message:   "Succesfully indexed and stored webpages",
				URLSeed:   response.URLSeed,
			}
			fmt.Println(messageStatus)
			break
		}
	}
}
