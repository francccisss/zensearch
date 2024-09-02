package connect

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func ConnectDatabase(TCPCon *amqp.Connection) {
	ch, err := TCPCon.Channel()
	if err != nil {
		log.Panicf("Unable to create a database channel.")
	}

	// create a channel for sending a message to the database
	// query_database_rpc

}
