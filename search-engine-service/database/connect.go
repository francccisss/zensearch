package database

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
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
	const queue = "query_database"
	const rpcQueue = "rpc_database_queue"
	reposponse, err := ch.QueueDeclare(
		queue, // name
		false, // durable
		true,  // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	err = ch.Publish(
		"",
		reposponse.Name,
		false, false, amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte("queryWebpages"),
		},
	)
	if err != nil {
		log.Panicf("Unable to push message to database service.")
	}

}
