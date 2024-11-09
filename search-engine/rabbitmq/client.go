package rabbitmq

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

var connections = map[string]*amqp.Connection{}
var channels = map[string]*amqp.Channel{}

/*
  Creates a new global reference to a specific tcp connection to rabbitmq
  setting new connection requres the connection Name to index
  the specific connection.
*/

// I know these are duplicates but they should work for now
// I should change this later

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

func SetNewChannel(name string, channel *amqp.Channel) *amqp.Channel {
	existingChan, ok := channels[name]
	if !ok {
		channels[name] = channel
		return channel
	}
	return existingChan
}

func GetChannel(name string) (*amqp.Channel, error) {
	if conn, ok := channels[name]; ok {
		return conn, nil
	}
	return nil, fmt.Errorf("Connection does not exist: %s\n", name)
}
