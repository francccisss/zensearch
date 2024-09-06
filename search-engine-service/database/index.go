package database

import (
	"log"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

func QueryDatabase(ch *amqp.Channel) amqp.Delivery {
	const queryQueue = "database_query_queue"
	const rpcQueue = "rpc_database_queue"
	corID := uuid.New().String()

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
	data := <-queriedData

	log.Printf("CorID response: %s\n", data.CorrelationId)
	log.Printf("End of Query\n")
	return data
}
