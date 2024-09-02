package database

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

func QueryDatabase(ch *amqp.Channel) <-chan amqp.Delivery {
	const queryQueue = "database_query_queue"
	const rpcQueue = "rpc_database_queue"
	const corID = "f8123727-50ac-4655-aefc-3defcbc695d0"

	ch.QueueDeclare(queryQueue, false, false, false, false, nil)
	ch.QueueDeclare(rpcQueue, false, false, false, false, nil)

	err := ch.Publish(
		"",
		queryQueue,
		false, false, amqp.Publishing{
			CorrelationId: corID,
			ReplyTo:       rpcQueue,
			ContentType:   "text/plain",
			Body:          []byte("queryWebpages"),
		},
	)

	if err != nil {
		log.Panicf(err.Error())
	}
	queriedData, err := ch.Consume(
		rpcQueue,
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

	return queriedData
}
