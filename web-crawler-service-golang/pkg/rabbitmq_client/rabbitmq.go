package rabbitmqclient

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

var connections = map[string]*amqp.Connection{}

/*
  Creates a new global reference to a specific tcp connection to rabbitmq
  setting new connection requres the connection Name to index
  the specific connection.
*/

func SetNewConnection(name string, conn *amqp.Connection) *amqp.Connection {
	existingconn, ok := connections[name]
	if !ok {
		connections[name] = conn
		return conn
	}
	return existingconn
}

func GetConnection(name string) (*amqp.Connection, error) {
	if conn, ok := connections[name]; ok {
		return conn, nil
	}
	return nil, fmt.Errorf("Connection does not exist: %s\n", name)
}
