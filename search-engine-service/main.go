package main

import (
	"encoding/json"
	"fmt"
	"log"
	"search-engine-service/bm25"
	"search-engine-service/rabbitmq"
	"search-engine-service/utilities"

	amqp "github.com/rabbitmq/amqp091-go"
)

// subsequent requests are not being pushed
func main() {
	searchQuery := ""

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
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

	queriedData, err := dbQueryChannel.Consume(
		rabbitmq.DB_RESPONSE_QUEUE,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Panicf(err.Error())
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
	failOnError(err, "Failed to register a consumer")

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
		case data := <-queriedData:
			{
				if searchQuery == "" {
					fmt.Print("Search Query is empty\n")
					continue
				}
				fmt.Print("Data from Database service retrieved\n")
				webpages := parseWebpageQuery(data.Body)
				fmt.Println(len(*webpages))

				calculatedRatings := bm25.CalculateBMRatings(searchQuery, webpages)
				rankedWebpages := bm25.RankBM25Ratings(calculatedRatings)
				for _, webpage := range *rankedWebpages {
					fmt.Printf("URL: %s\n", webpage.Url)
					fmt.Printf("TF Score: %f\n", webpage.TokenRating.TfRating)
					fmt.Printf("BM25 Score: %f\n\n", webpage.TokenRating.Bm25rating)
				}
				fmt.Printf("Search Query for composite query: %s\n\n", searchQuery)
				rabbitmq.PublishScoreRanking(rankedWebpages, mainChannel)
				dbQueryChannel.Ack(data.DeliveryTag, true)
			}
		}
	}
}

// maybe use message for cache validation later on for optimization

func parseWebpageQuery(data []byte) *[]utilities.WebpageTFIDF {
	var webpages []utilities.WebpageTFIDF
	err := json.Unmarshal(data, &webpages)
	failOnError(err, "Unable to Decode json data from database.")
	return &webpages
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err.Error())
	}
}
