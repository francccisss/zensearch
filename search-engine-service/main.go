package main

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	// "search-engine-service/database"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	fail_on_error(err, "Failed to create a new TCP Connection")
	fmt.Printf("Established TCP Connection with RabbitMQ\n")

	channel, err := conn.Channel()
	fail_on_error(err, "Failed to create a new Channel")
	defer channel.Close()

	queryChannel, err := conn.Channel()
	fail_on_error(err, "Failed to create a new Channel")
	defer queryChannel.Close()

	const searchQueue = "search_queue"
	sQueue, err := channel.QueueDeclare(searchQueue,
		false, false, false, false, nil,
	)
	fail_on_error(err, "Failed to create search queue")

	const queryQueue = "database_query_queue"
	queryChannel.QueueDeclare(
		queryQueue, // name
		false,      // durable
		false,      // delete when unused
		true,       // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	fail_on_error(err, "Failed to create query queue")

	msgs, err := channel.Consume(
		sQueue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	fail_on_error(err, "Failed to register a consumer")

	for d := range msgs {
		// processSearchQuery(string(d.Body), query_ch)
		// const rpcQueue = "rpc_database_queue"
		// queryChannel.Publish(
		// 	rpcQueue,
		// 	qQueue.Name,
		// 	false, false, amqp.Publishing{
		// 		ContentType: "text/plain",
		// 		Body:        []byte("queryWebpages"),
		// 	},
		// )
		defer queryChannel.Close()
		log.Printf("Consumed %p", d.Body)
	}
}

// func processSearchQuery(msg string, ch *amqp.Channel) {
// 	const rpcQueue = "rpc_database_queue"
// 	log.Printf("Consumed %s", msg)
// 	// database.QueryDatabase(ch)
//
// 	ch.Publish(
// 		rpcQueue,
// 		reposponse.Name,
// 		false, false, amqp.Publishing{
// 			ContentType: "text/plain",
// 			Body:        []byte("queryWebpages"),
// 		},
// 	)
// 	defer ch.Close()
//
// 	fmt.Printf("Close Thread")
// }

func fail_on_error(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
