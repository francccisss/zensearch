package rabbitmq

import (
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var connections = map[string]*amqp.Connection{}

func EstablishConnection(retries int) error {

	if retries > 0 {
		conn, err := amqp.Dial("amqp://localhost:5672/")
		if err != nil {
			retries--
			fmt.Println("Retrying Crawler service connection")
			time.Sleep(2000 * time.Millisecond)
			return EstablishConnection(retries)
		}
		SetNewConnection("conn", conn)
		return nil
	}

	return fmt.Errorf("Shutting down crawler service after serveral retries")
}

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
