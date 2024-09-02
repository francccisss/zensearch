package main

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"search-engine-service/database"
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
	log.Printf("queried data: %s\n", rpcQueue)
	data := <-database.QueryDatabase(ch)
	log.Printf("queried data: %s\n", data.Body)

}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err.Error())
	}
}
