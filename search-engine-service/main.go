package main

import (
	"encoding/json"
	"fmt"
	"log"
	// "search-engine-service/database"
	// tfidf "search-engine-service/tf-idf"
	amqp "github.com/rabbitmq/amqp091-go"
	"search-engine-service/utilities"
)

const QUERYQUEUE = "database_query_queue"
const DB_RESPONSE_QUEUE = "database_response_queue"
const SEARCHQUEUE = "search_queue"

// subsequent requests are not being pushed
func main() {

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to create a new TCP Connection")
	fmt.Printf("Established TCP Connection with RabbitMQ\n")

	// DECLARING CHANNELS
	consumerChannel, err := conn.Channel()
	failOnError(err, "Failed to create a new Channel")
	dbQueryChannel, err := conn.Channel()
	failOnError(err, "Failed to create a new Channel")
	// DECLARING CHANNELS

	// DECLARING QUEUES
	consumerChannel.QueueDeclare(SEARCHQUEUE,
		false, false, false, false, nil)
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
	msgs, err := consumerChannel.Consume(
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
		case searchQuery := <-msgs:
			{
				pushQuery(string(searchQuery.Body), dbQueryChannel)
				fmt.Print("Process Done.\n")
			}
		case data := <-queriedData:
			{
				fmt.Print("Data from Database service retrieved\n")
				fmt.Printf("Data %s\n", string(data.Body))
				// parseWebpageQuery(data.Body)
			}

		}
	}
	fmt.Print("Main Exit.")
}

func pushQuery(searchQuery string, ch *amqp.Channel) {
	err := ch.Publish(
		"",
		QUERYQUEUE,
		false, false, amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte("queryWebpages"),
		},
	)
	fmt.Printf("Push message to database service.\n")
	if err != nil {
		log.Panicf(err.Error())
	}
	log.Printf("End of Query\n")
}

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
