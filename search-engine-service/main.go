package main

import (
	"encoding/json"
	"fmt"
	"log"
	"search-engine-service/rabbitmq"
	tfidf "search-engine-service/tf-idf"
	"search-engine-service/utilities"

	amqp "github.com/rabbitmq/amqp091-go"
)

const QUERYQUEUE = "database_query_queue"
const PUBLISH_QUEUE = "publish_ranking_queue"
const DB_RESPONSE_QUEUE = "database_response_queue"
const SEARCHQUEUE = "search_queue"

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

	// DECLARING QUEUES
	mainChannel.QueueDeclare(SEARCHQUEUE, false, false, false, false, nil)
	failOnError(err, "Failed to create search queue")
	dbQueryChannel.QueueDeclare(QUERYQUEUE, false, false, false, false, nil)
	failOnError(err, "Failed to create query queue")
	dbQueryChannel.QueueDeclare(DB_RESPONSE_QUEUE, false, false, false, false, nil)
	failOnError(err, "Failed to create query queue")
	// DECLARING QUEUES

	queriedData, err := dbQueryChannel.Consume(
		DB_RESPONSE_QUEUE,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Panicf(err.Error())
	}
	msgs, err := mainChannel.Consume(
		SEARCHQUEUE,
		"",
		true,
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
				// assign tfscore to each webpages.
				tfidf.CalculateTF(searchQuery, &webpages)
				IDF := tfidf.CalculateIDF(searchQuery, &webpages)
				rankedWebpages := tfidf.RankTFIDFRatings(IDF, &webpages)
				rabbitmq.PublishScoreRanking(rankedWebpages, mainChannel)
			}
		}
	}
}

// maybe use message for cache validation later on for optimization

func parseWebpageQuery(data []byte) []utilities.WebpageTFIDF {
	var webpages []utilities.WebpageTFIDF // I dont know why it doesnt work
	err := json.Unmarshal(data, &webpages)
	failOnError(err, "Unable to Decode json data from database.")
	return webpages
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err.Error())
	}
}