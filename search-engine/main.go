package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"search-engine/constants"
	"search-engine/internal/bm25"
	"search-engine/internal/rabbitmq"
	"search-engine/internal/segment_serializer"
	"search-engine/utilities"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// TODO SYSTEM ERRORS SHOULD RESTART THE SERVICE... I DONT KNOW HOW TO DO IT

// Maximum segment size in bytes

func main() {

	err := rabbitmq.EstablishConnection(7)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	conn, err := rabbitmq.GetConnection("conn")
	if err != nil {
		fmt.Println("Connection does not exist")
		os.Exit(1)
	}
	fmt.Println("Search engine established TCP Connection with RabbitMQ")

	// DECLARING CHANNELS
	mainChannel, err := conn.Channel()
	failOnError(err, "Failed to create a main Channel")
	dbQueryChannel, err := conn.Channel()
	failOnError(err, "Failed to create a database Channel")

	// SET PREFETCH FOR CUMULATIVE ACKS
	dbQueryChannel.Qos(constants.CMLTV_ACK, 0, false)

	rabbitmq.SetNewChannel("dbChannel", dbQueryChannel)
	rabbitmq.SetNewChannel("mainChannel", mainChannel)

	// DECLARING CHANNELS

	defer func() {
		conn.Close()
		mainChannel.Close()
		dbQueryChannel.Close()
	}()

	// DECLARING QUEUES
	mainChannel.QueueDeclare(rabbitmq.EXPRESS_SENGINE_QUERY_QUEUE, false, false, false, false, nil)
	failOnError(err, "Failed to create search queue")
	mainChannel.QueueDeclare(rabbitmq.SENGINE_EXPRESS_QUERY_CBQ, false, false, false, false, nil)
	failOnError(err, "Failed to create publish queue")

	dbQueryChannel.QueueDeclare(rabbitmq.SENGINE_DB_REQUEST_QUEUE, false, false, false, false, nil)
	failOnError(err, "Failed to create query queue")
	dbQueryChannel.QueueDeclare(rabbitmq.DB_SENGINE_REQUEST_CBQ, false, false, false, false, nil)
	failOnError(err, "Failed to create db response queue")
	// DECLARING QUEUES

	msgs, err := mainChannel.Consume(
		rabbitmq.EXPRESS_SENGINE_QUERY_QUEUE,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Panicf("Unable to listen to %s", rabbitmq.EXPRESS_SENGINE_QUERY_QUEUE)
	}

	searchQueryChan := make(chan string)
	incomingSegmentsChan := make(chan amqp.Delivery)
	webpageBytesChan := make(chan bytes.Buffer)
	var currentSearchQuery string

	// Receiving User's Query
	go func(searchQueryChan chan string) {
		for userSearch := range msgs {
			searchQuery := string(userSearch.Body)

			searchQueryChan <- searchQuery
			mainChannel.Ack(userSearch.DeliveryTag, true)

			fmt.Printf("User's Query: %s\n", searchQuery)
			currentSearchQuery = searchQuery
		}
	}(searchQueryChan)

	// Consumes and pushes segments to the `incomingSegmentsChan` channel
	go func(chann *amqp.Channel) {

		dbMsg, err := chann.Consume(
			rabbitmq.DB_SENGINE_REQUEST_CBQ,
			"",
			false,
			false,
			false,
			false,
			nil,
		)

		if err != nil {
			log.Panicf("Unable to listen to %s", rabbitmq.DB_SENGINE_REQUEST_CBQ)
		}

		// Consume and send segment to segment channel
		for incomingSegment := range dbMsg {
			incomingSegmentsChan <- incomingSegment
		}

	}(dbQueryChannel)

	// the consumed incoming segments will be processed here and
	go func() {
		for searchQuery := range searchQueryChan {

			fmt.Print("Query database\n")
			// Queries database to send segments to search engine
			go rabbitmq.QueryDatabase(searchQuery)

			fmt.Print("Spawn segment listener\n")

			// Listens for incoming segments from the database Query channel consumer
			go segments.HandleIncomingSegments(dbQueryChannel, incomingSegmentsChan, webpageBytesChan)
		}
	}()

	// Handling search engine logic for parsing webpage to json, ranking and data segmentation for transpotation
	go func() {

		// TODO THROW ERRORS TO FRONT END
		for webpageBuffer := range webpageBytesChan {
			// Parsing webpages

			timeStart := time.Now()
			webpages, err := utilities.ParseWebpages(webpageBuffer.Bytes())
			if err != nil {
				fmt.Println(err.Error())
				log.Println("Unable to parse webpages")
				continue
			}
			fmt.Printf("Time elapsed parsing: %dms\n", time.Until(timeStart).Abs().Milliseconds())

			// Ranking webpages
			timeStart = time.Now()

			calculatedRatings := bm25.CalculateBMRatings(currentSearchQuery, webpages)
			rankedWebpages := bm25.RankBM25Ratings(calculatedRatings)

			fmt.Printf("Total ranked webpages: %d\n", len(*rankedWebpages))
			fmt.Printf("Time elapsed ranking: %dms\n", time.Until(timeStart).Abs().Milliseconds())

			// create segments in this section after ranking
			timeStart = time.Now()
			segments, err := segments.CreateSegments(rankedWebpages, constants.MSS)
			if err != nil {
				fmt.Println(err.Error())
				log.Println("Unable to create segments")
				continue
			}

			fmt.Printf("Time elapsed data segmentation: %dms\n", time.Until(timeStart).Abs().Milliseconds())
			go rabbitmq.PublishScoreRanking(segments)

		}

	}()

	// need to signal this loop to stop if error or graceful exits
	loop := make(chan bool)
	loop <- true

}

// TODO Instead of panicking, create a recursive retry and then close application
func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err.Error())
	}
}
