package main

import (
	"crawler/internal/rabbitmq"
	"crawler/internal/types"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"sync"
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
	go MockDatabase(sm)
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
	_, err = dbChannel.QueueDeclare(rabbitmq.DB_CRAWLER_INDEXING_CBQ, false, false, false, false, nil)

	if err != nil {
		fmt.Printf("Unable to assert queue=%s\n", rabbitmq.DB_CRAWLER_INDEXING_CBQ)
		t.Fatal(err)
	}
	_, err = frontierChannel.QueueDeclare("db_crawler_dequeue_url_cbq", false, false, false, false, nil)

	if err != nil {
		fmt.Printf("Unable to assert queue=%s\n", "db_crawler_dequeue_url_cbq")
		t.Fatal(err)
	}

	// frontierChannel, err := conn.Channel()
	// _, err = frontierChannel.QueueDeclare("db_crawler_dequeue_url_cbq", false, false, false, false, nil)
	// rabbitmq.SetNewChannel("frontierChannel", frontierChannel)
	rabbitmq.SetNewChannel("dbChannel", dbChannel)
	defer dbChannel.Close()
	rabbitmq.SetNewChannel("frontierChannel", frontierChannel)
	defer frontierChannel.Close()

	response := make(chan DBResponse, 10)
	go DBTestChannelListener(dbChannel, response)
	go func() {
		for {
			select {
			case r := <-response:
				fmt.Println(r)
			}
		}
	}()

	seeds := []string{"https://magill.dev/"}

	// seeds := []string{"https://gobyexample.com/", "https://fzaid.vercel.app/"}
	fmt.Printf("Crawling seeds: %+v\n", seeds)
	spawner := NewSpawner(10, seeds)
	spawner.SpawnCrawlers()
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
	_, err = mockDBChann.QueueDeclare(rabbitmq.CRAWLER_DB_INDEXING_QUEUE, false, false, false, false, nil)

	if err != nil {
		fmt.Printf("Unable to assert queue=%s\n", rabbitmq.CRAWLER_DB_INDEXING_QUEUE)
		panic(err)
	}
	crawlerMsg, err := mockDBChann.Consume(rabbitmq.CRAWLER_DB_INDEXING_QUEUE, "", false, false, false, false, nil)

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

	var err = rabbitmq.EstablishConnection(7)
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
			err = SendCrawlMessageStatus(messageStatus)
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

func TestDequeueUrls(t *testing.T) {

	var err = rabbitmq.EstablishConnection(7)

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
	dbChannel.QueueDeclare(rabbitmq.DB_CRAWLER_INDEXING_CBQ, false, false, false, false, nil)

	frontierChannel.QueueDeclare(rabbitmq.DB_CRAWLER_DEQUEUE_URL_CBQ, false, false, false, false, nil)
	frontierChannel.QueueDeclare(rabbitmq.CRAWLER_DB_DEQUEUE_URL_QUEUE, false, false, false, false, nil)

	rabbitmq.SetNewChannel("dbChannel", dbChannel)
	defer dbChannel.Close()
	rabbitmq.SetNewChannel("frontierChannel", frontierChannel)
	defer frontierChannel.Close()

	crawler, err := NewCrawler("https://fzaid.vercel.app")
	if err != nil {
		t.Fatal(err)
	}

	crawler.Crawl()
	fmt.Println("TEST: Test done")
}
func DBTestChannelListener(chann *amqp.Channel, resultChan chan DBResponse) {
	fmt.Println("TEST: DB CHANNEL LISTENER")

	dbMsg, err := chann.Consume(rabbitmq.DB_CRAWLER_INDEXING_CBQ, "", false, false, false, false, nil)
	if err != nil {
		panic("Unable to listen to db server")
	}
	fmt.Printf("NOTIF: listenting to %s\n", rabbitmq.DB_CRAWLER_INDEXING_CBQ)
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
