package rabbitmqclient

import (
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

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
