package main

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"search-engine-service/database"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	fail_on_error(err, "Failed to create a new TCP Connection")
	fmt.Printf("Established TCP Connection with RabbitMQ\n")
	ch, err := conn.Channel()
	fail_on_error(err, "Failed to create a new Channel")
	defer ch.Close()

	const queue = "search_queue"
	q, err := ch.QueueDeclare(queue,
		false, false, false, false, nil,
	)
	fail_on_error(err, "Failed to connect a to a message queue")

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	fail_on_error(err, "Failed to register a consumer")

	// msgs IS A CHANNEL BUFFER, AND NEEDS TO WAIT UNTIL DATA IS PUSED INTO THE msgs CHANNEL BUFFER;
	for d := range msgs {
		fmt.Printf("Received a message: %+v", d.Body)
		go processSearchQuery(string(d.Body), conn)
	}
	log.Printf("Consumed %p", msgs)
}

func processSearchQuery(msg string, conn *amqp.Connection) {
	ch, err := database.CreateDatabaseChannel(conn)
	if err != nil {
		log.Panicf("Unable to create a database channel.")
	}
	database.QueryDatabase(ch)
	log.Printf("Consumed %s", msg)
	fmt.Printf("Close Thread")
}

func fail_on_error(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
