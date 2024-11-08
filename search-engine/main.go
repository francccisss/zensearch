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

// subsequent requests are not being pushed
func main() {
	searchQuery := ""

	conn, err := amqp.Dial("amqp://rabbitmq:5672/")
	failOnError(err, "Failed to create a new TCP Connection")
	fmt.Printf("Established TCP Connection with RabbitMQ\n")

	// DECLARING CHANNELS
	mainChannel, err := conn.Channel()
	failOnError(err, "Failed to create a new Channel")
	dbQueryChannel, err := conn.Channel()
	failOnError(err, "Failed to create a new Channel")
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

	dbMsg, err := dbQueryChannel.Consume(
		rabbitmq.DB_RESPONSE_QUEUE,
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		fmt.Printf("Unable to listen to %s", rabbitmq.DB_RESPONSE_QUEUE)
	}
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

	var (
		segmentCounter      uint32 = 0
		expectedSequenceNum uint32 = 0
	)

	webpageBytes := []byte{}
	for {
		select {
		case userSearch := <-msgs:
			{
				searchQuery = string(userSearch.Body)
				if searchQuery == "" {
					fmt.Print("Search Query is empty\n")
					continue
				}
				rabbitmq.QueryDatabase(string(userSearch.Body), dbQueryChannel)
				mainChannel.Ack(userSearch.DeliveryTag, true)
				fmt.Print("Process Done.\n")
			}
		case segment := <-dbMsg:
			{
				if searchQuery == "" {
					fmt.Print("Search Query is empty\n")
					continue
				}
				fmt.Print("Data from Database service retrieved\n")

				segmentHeader, err := GetSegmentHeader(segment.Body)
				if err != nil {
					fmt.Println("Unable to extract segment header")
				}

				segmentPayload, err := GetSegmentPayload(segment.Body)
				if err != nil {
					fmt.Println("Unable to extract segment payload")
					continue
				}

				// for retransmission/requeuing
				if segmentHeader.SequenceNum != expectedSequenceNum {
					dbQueryChannel.Nack(segment.DeliveryTag, true, true)
					fmt.Printf("Expected Sequence number %d, got %d\n",
						expectedSequenceNum, segmentHeader.SequenceNum)
					continue
				}

				segmentCounter++
				expectedSequenceNum++

				dbQueryChannel.Ack(segment.DeliveryTag, false)
				webpageBytes = append(webpageBytes, segmentPayload...)
				fmt.Printf("Byte Length: %d\n", len(webpageBytes))

				if segmentCounter == segmentHeader.TotalSegments {
					log.Printf("Received all of the segments from Database %d", segmentCounter)
					fmt.Printf("Total Byte Length: %d\n", len(webpageBytes))

					webpages, err := ParseWebpages(webpageBytes)
					if err != nil {
						fmt.Printf(err.Error())
						log.Panicf("Unable to parse webpages")
						continue
					}

					calculatedRatings := bm25.CalculateBMRatings(searchQuery, webpages, bm25.AvgDocLen(webpages))
					rankedWebpages := bm25.RankBM25Ratings(calculatedRatings)
					for _, webpage := range *rankedWebpages {
						fmt.Printf("URL: %s\n", webpage.Url)
						fmt.Printf("BM25 Score: %f\n\n", webpage.TokenRating.Bm25rating)
					}
					fmt.Printf("Search Query for composite query: %s\n\n", searchQuery)
					rabbitmq.PublishScoreRanking(rankedWebpages, mainChannel)

					// reset everything
					expectedSequenceNum = 0
					segmentCounter = 0
					webpageBytes = nil

					break
				}
			}
		}
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
		return &[]utilities.WebpageTFIDF{}, err
	}
	return &webpages, nil
}
