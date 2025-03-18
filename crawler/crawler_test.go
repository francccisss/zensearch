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
	dbChannel.QueueDeclare(rabbitmq.DB_CRAWLER_INDEXING_NOTIF_CBQ, false, false, false, false, nil)
	rabbitmq.SetNewChannel("dbChannel", dbChannel)
	defer dbChannel.Close()

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
