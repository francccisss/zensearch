package database

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func CreateDatabaseChannel(TCPCon *amqp.Connection) (*amqp.Channel, error) {
	ch, err := TCPCon.Channel()
	if err != nil {
		log.Panicf("Unable to create a database channel.")
		return ch, nil
	}
	return ch, nil
}

func QueryDatabase(ch *amqp.Channel) {
	const queue = "database_query_queue"
	const rpcQueue = "rpc_database_queue"
	response, err := ch.QueueDeclare(
		queue, // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		log.Panicf("Unable to declare a queue.")
	}
	fmt.Printf(response.Name)
	// ch.Publish(
	// 	rpcQueue,
	// 	reposponse.Name,
	// 	false, false, amqp.Publishing{
	// 		ContentType: "text/plain",
	// 		Body:        []byte("queryWebpages"),
	// 	},
	// )

}
