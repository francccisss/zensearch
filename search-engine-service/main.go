package main

import (
	"encoding/json"
	"fmt"
	"log"
	// "search-engine-service/database"
	// tfidf "search-engine-service/tf-idf"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"search-engine-service/utilities"
)

const QUERYQUEUE = "database_query_queue"
const RPCQUEUE = "rpc_database_queue"
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
	sQueue, err := consumerChannel.QueueDeclare(SEARCHQUEUE,
		false, false, false, false, nil,
	)
	failOnError(err, "Failed to create search queue")
	dbQueryChannel.QueueDeclare(QUERYQUEUE, false, false, false, false, nil)
	failOnError(err, "Failed to create query queue")
	// DECLARING QUEUES

	msgs, err := consumerChannel.Consume(
		sQueue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to register a consumer")

	for d := range msgs {
		corID := uuid.New().String()
		err := dbQueryChannel.Publish(
			"",
			QUERYQUEUE,
			false, false, amqp.Publishing{
				CorrelationId: corID,
				ReplyTo:       RPCQUEUE,
				ContentType:   "text/plain",
				Body:          []byte("queryWebpages"),
			},
		)
		fmt.Printf("Push message to database service.\n")
		if err != nil {
			log.Panicf(err.Error())
		}
		processSearchQuery(string(d.Body), conn)
		fmt.Print("Process Done.\n")
	}
	fmt.Print("Main Exit.")
}

func processSearchQuery(searchQuery string, conn *amqp.Connection) {

	rpcQueryChannel, err := conn.Channel()
	queriedData, err := rpcQueryChannel.Consume(
		RPCQUEUE,
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
	data := <-queriedData

	log.Printf("CorID response: %s\n", data.CorrelationId)
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
