package main

import (
	"encoding/json"
	"fmt"
	"log"
	"search-engine/internal/bm25"
	"search-engine/internal/rabbitmq"
	"search-engine/internal/segments"

	amqp "github.com/rabbitmq/amqp091-go"
)

// TODO SYSTEM ERRORS SHOULD RESTART THE SERVICE... I DONT KNOW HOW TO DO IT

// Maximum segment size in bytes
const MSS = 100000

func main() {

	conn, err := amqp.Dial("amqp://rabbitmq:5672/")
	failOnError(err, "Failed to create a new TCP Connection")
	fmt.Printf("Established TCP Connection with RabbitMQ\n")

	// DECLARING CHANNELS
	mainChannel, err := conn.Channel()
	failOnError(err, "Failed to create a main Channel")
	dbQueryChannel, err := conn.Channel()
	failOnError(err, "Failed to create a database Channel")

	rabbitmq.SetNewChannel("dbChannel", dbQueryChannel)
	rabbitmq.SetNewChannel("mainChannel", mainChannel)

	// DECLARING CHANNELS

	defer func() {
		conn.Close()
		mainChannel.Close()
		dbQueryChannel.Close()
	}()

	// DECLARING QUEUES
	mainChannel.QueueDeclare(rabbitmq.SEARCH_QUEUE, false, false, false, false, nil)
	failOnError(err, "Failed to create search queue")
	mainChannel.QueueDeclare(rabbitmq.PUBLISH_QUEUE, false, false, false, false, nil)
	failOnError(err, "Failed to create publish queue")

	dbQueryChannel.QueueDeclare(rabbitmq.DB_QUERY_QUEUE, false, false, false, false, nil)
	failOnError(err, "Failed to create query queue")
	dbQueryChannel.QueueDeclare(rabbitmq.DB_RESPONSE_QUEUE, false, false, false, false, nil)
	failOnError(err, "Failed to create db response queue")
	// DECLARING QUEUES

	msgs, err := mainChannel.Consume(
		rabbitmq.SEARCH_QUEUE,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		fmt.Printf("Unable to listen to %s", rabbitmq.SEARCH_QUEUE)
	}

	searchQueryChan := make(chan string)

	// Concurrency Pipline
	go func(searchQueryChan <-chan string) {
		for {
			searchQuery := <-searchQueryChan

			// Should i use go routines? its still going to be an unbuffered channel anyways
			// so might as well just make everything synchronous

			// Segments retrieval
			fmt.Printf("Search query retrieved: `%s`\n", searchQuery)
			webpageBytes, err := Segments.ListenIncomingSegments(searchQuery)
			fmt.Printf("Total Bytes received: %d\n", len(webpageBytes))

			if err != nil {
				fmt.Printf("Something went wrong while listening to incoming data segments from database\n")
				log.Panicf(err.Error())
			}

			// For ranking webpages
			webpages, err := ParseWebpages(webpageBytes)
			if err != nil {
				fmt.Printf(err.Error())
				log.Panicf("Unable to parse webpages")
			}
			calculatedRatings := bm25.CalculateBMRatings(searchQuery, webpages, bm25.AvgDocLen(webpages))
			rankedWebpages := bm25.RankBM25Ratings(calculatedRatings)

			fmt.Printf("Total ranked webpages: %d\n", len(*rankedWebpages))

			// create segments in this section after ranking
			segments, err := Segments.CreateSegments(rankedWebpages, MSS)
			if err != nil {
				fmt.Println(err.Error())
				log.Panicf("Unable to create segments")
			}
			go rabbitmq.PublishScoreRanking(segments)
		}
	}(searchQueryChan)

	go func(searchQueryChan chan string) {
		for {
			userSearch := <-msgs
			searchQuery := string(userSearch.Body)
			if searchQuery == "" {
				log.Panicf("Search Query is empty\n")
			}

			fmt.Printf("User's Query: %s\n", searchQuery)
			mainChannel.Ack(userSearch.DeliveryTag, true)
			// block until we received a search query
			// if Process Done log is not called means
			// searchQueryChan is not working for some reason
			searchQueryChan <- searchQuery
			fmt.Print("Process Done.\n")
			go rabbitmq.QueryDatabase(searchQuery)
		}
	}(searchQueryChan)

	// need to signal this loop to stop if error or graceful exits
	loop := make(chan bool)
	<-loop

}

// maybe use message for cache validation later on for optimization

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err.Error())
	}
}

func ParseWebpages(data []byte) (*[]bm25.WebpageTFIDF, error) {
	var webpages []bm25.WebpageTFIDF
	err := json.Unmarshal(data, &webpages)
	if err != nil {
		return nil, err
	}
	return &webpages, nil
}
