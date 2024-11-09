package main

import (
	"encoding/json"
	"fmt"
	"log"
	"search-engine/bm25"
	"search-engine/rabbitmq"
	"search-engine/utilities"

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

			fmt.Printf("Search query retrieved: `%s`", searchQuery)

			webpageBytes, err := ListenIncomingSegments(searchQuery)
			fmt.Printf("Total Byte Length: %d\n", len(webpageBytes))

			if err != nil {
				fmt.Printf("Something went wrong while listening to incoming data segments from database\n")
				log.Printf(err.Error())
			} // DONT TOUCH THIS
			webpages, err := ParseWebpages(webpageBytes)

			if err != nil {
				fmt.Printf(err.Error())
				log.Panicf("Unable to parse webpages")
			}
			// create segments in this section after ranking
			calculatedRatings := bm25.CalculateBMRatings(searchQuery, webpages, bm25.AvgDocLen(webpages))
			rankedWebpages := bm25.RankBM25Ratings(calculatedRatings)
			fmt.Printf("Search Query for composite query: %s\n\n", searchQuery)

			segments, err := CreateSegments(rankedWebpages, MSS)
			if err != nil {
				fmt.Println(err.Error())
				log.Panicf("Unable to create segments")
			}

			b := []byte{}
			for _, segment := range segments {
				b = append(b, segment...)
			}
			we, err := ParseWebpages(b)
			if err != nil {
				fmt.Println("Unable to parse webpages")
				log.Panicf(err.Error())
			}
			fmt.Println(we)
			// rabbitmq.PublishScoreRanking(rankedWebpages)
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
			rabbitmq.QueryDatabase(searchQuery)
		}
	}(searchQueryChan)

	// need to signal this loop to stop if error or graceful exits
	for {

	}

}

// maybe use message for cache validation later on for optimization

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err.Error())
	}
}

func ParseWebpages(data []byte) (*[]utilities.WebpageTFIDF, error) {
	var webpages []utilities.WebpageTFIDF
	err := json.Unmarshal(data, &webpages)
	if err != nil {
		return nil, err
	}
	return &webpages, nil
}
