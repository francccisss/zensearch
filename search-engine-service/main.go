package main

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

func main() {
	fmt.Printf("Hello, World")

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	fail_on_error(err, "Failed to create a new TCP Connection")
	fmt.Printf("connected")

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

	for d := range msgs {
		log.Printf("Received a message: %s", d.Body)
	}

}

func fail_on_error(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
