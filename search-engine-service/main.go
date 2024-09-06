package main

import (
	"encoding/json"
	"fmt"
	"log"
	"search-engine-service/database"
	"search-engine-service/utilities"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to create a new TCP Connection")
	fmt.Printf("Established TCP Connection with RabbitMQ\n")

	channel, err := conn.Channel()
	failOnError(err, "Failed to create a new Channel")
	defer channel.Close()

	queryChannel, err := conn.Channel()
	failOnError(err, "Failed to create a new Channel")

	const searchQueue = "search_queue"
	sQueue, err := channel.QueueDeclare(searchQueue,
		false, false, false, false, nil,
	)
	failOnError(err, "Failed to create search queue")

	const queryQueue = "database_query_queue"
	queryChannel.QueueDeclare(
		queryQueue, // name
		false,      // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	failOnError(err, "Failed to create query queue")
	msgs, err := channel.Consume(
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
		go processSearchQuery(string(d.Body), queryChannel)
		// log.Printf("Consumed %s", d.Body)
	}
	defer queryChannel.Close()
}

func processSearchQuery(searchQuery string, ch *amqp.Channel) {
	const rpcQueue = "rpc_database_queue"
	const queryQueue = "database_query_queue"
	data := <-database.QueryDatabase(ch)
	parseWebpageQuery(data.Body)
}

func parseWebpageQuery(data []byte) []utilities.WebpageTFIDF {
	fmt.Printf("%s", data)
	var webpages []utilities.WebpageTFIDF // I dont know why it doesnt work
	err := json.Unmarshal(data, &webpages)
	failOnError(err, "Unable to Decode json data from database.")
	fmt.Printf("Decoded Data: %+v\n", webpages)
	return webpages
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err.Error())
	}
}
