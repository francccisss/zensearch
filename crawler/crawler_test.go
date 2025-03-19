package main

import (
	"crawler/internal/rabbitmq"
	"crawler/internal/types"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

var err = rabbitmq.EstablishConnection(7)

func TestCrawlerIndexing(t *testing.T) {
	sm := make(chan struct{}, 1)

	osSignalChan := make(chan os.Signal, 1)

	signal.Notify(osSignalChan, syscall.SIGINT, syscall.SIGTERM)

	if err != nil {
		fmt.Println(err.Error())
		t.Fatal(err)
	}
	conn, err := rabbitmq.GetConnection("conn")
	go MockDatabase(sm)
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
	_, err = dbChannel.QueueDeclare(rabbitmq.DB_CRAWLER_INDEXING_NOTIF_CBQ, false, false, false, false, nil)

	if err != nil {
		fmt.Printf("Unable to assert queue=%s\n", rabbitmq.DB_CRAWLER_INDEXING_NOTIF_CBQ)
		t.Fatal(err)
	}
	rabbitmq.SetNewChannel("dbChannel", dbChannel)
	defer dbChannel.Close()

	response := make(chan DBResponse, 10)
	go DBTestChannelListener(dbChannel, response)
	go func() {
		for {
			select {
			case r := <-response:
				fmt.Println(r)
			case <-osSignalChan:
				fmt.Println("Close gracefully")
				return
			}
		}
	}()

	seeds := []string{"https://fzaid.vercel.app/"}

	// seeds := []string{"https://gobyexample.com/"}
	// seeds := []string{"https://gobyexample.com/", "https://fzaid.vercel.app/"}
	fmt.Printf("Crawling seeds: %+v\n", seeds)
	spawner := NewSpawner(10, seeds)
	crawlResults := spawner.SpawnCrawlers()
	fmt.Printf("TEST: crawl_results=%+v", crawlResults)

	fmt.Printf("\n\n------RESULTS------\n")
	for result := range crawlResults.CrawlResultsChan {
		fmt.Printf("| SEED: %s |\n", result.URLSeed)
		fmt.Printf("TEST: crawl_status=%d\n", result.CrawlStatus)
		fmt.Printf("TEST: message=%s\n", result.Message)

	}
	sm <- struct{}{}
	fmt.Println("TEST: test end")
}

func MockDatabase(sm <-chan struct{}) {
	conn, err := rabbitmq.GetConnection("conn")
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
	mockDBChann, err := conn.Channel()
	if err != nil {
		fmt.Println("TEST DBMOCK: ERROR Unable to create a mock DB channel")
		panic(err)
	}
	_, err = mockDBChann.QueueDeclare(rabbitmq.CRAWLER_DB_INDEXING_NOTIF_QUEUE, false, false, false, false, nil)

	if err != nil {
		fmt.Printf("Unable to assert queue=%s\n", rabbitmq.CRAWLER_DB_INDEXING_NOTIF_QUEUE)
		panic(err)
	}
	crawlerMsg, err := mockDBChann.Consume("", rabbitmq.CRAWLER_DB_INDEXING_NOTIF_QUEUE, false, false, false, false, nil)

	database := []types.IndexedWebpage{}
	for {
		// for msg := range crawlerMsg {
		select {
		case <-sm:
			fmt.Println("\n\n------CLOSING MOCKDB------")
			for _, page := range database {
				fmt.Printf("TEST DBMOCK: title: %+s\n", page.Header.Title)
				fmt.Printf("TEST DBMOCK: content_length: %d\n", len(page.Contents))
			}
			return
		case msg := <-crawlerMsg:
			webpage := &types.IndexedWebpage{}
			err := json.Unmarshal(msg.Body, webpage)
			if err != nil {
				mockDBChann.Nack(msg.DeliveryTag, false, false)
				panic(err)
			}
			mockDBChann.Ack(msg.DeliveryTag, false)
			fmt.Println("TEST DBMOCK: Appended new webpage to database")
			database = append(database, *webpage)
			fmt.Printf("TEST DBMOCK: webpages_count=%d\n", len(database))
		}

	}
}

func TestSendMessageToExpress(t *testing.T) {
	URLSeeds := []struct {
		seed    string
		msg     string
		success bool
	}{
		{
			seed:    "https://fasdas",
			msg:     "Something went wrong, unable to crawl URLSeed",
			success: false,
		},
		{
			seed:    "https://fzaid.vercel.app",
			msg:     "Something went wrong, unable to crawl URLSeed",
			success: false,
		},
	}
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

	var wg sync.WaitGroup
	for _, url := range URLSeeds {
		wg.Add(1)
		go func() {

			defer wg.Done()
			messageStatus := CrawlMessageStatus{
				IsSuccess: url.success,
				Message:   url.msg, // need response directly from database
				// if testing make sure URLSeed matches the crawled url on the client
				URLSeed: url.seed,
			}
			fmt.Println("TEST: Sending Message to Express")
			err = SendCrawlMessageStatus(messageStatus, expressChannel, rabbitmq.CRAWLER_EXPRESS_CBQ)
			if err != nil {
				fmt.Printf("ERROR: Unable to send message status through %s\n", rabbitmq.CRAWLER_EXPRESS_CBQ)
				fmt.Printf("ERROR: %s", err)
				panic(err)
			}
		}()
	}
	wg.Wait()
	fmt.Println("TEST: Done")
}

func DBTestChannelListener(chann *amqp.Channel, resultChan chan DBResponse) {
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
