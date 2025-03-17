package rabbitmq

import (
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var connections = map[string]*amqp.Connection{}
var channels = map[string]*amqp.Channel{}

func EstablishConnection(retries int) error {

	if retries > 0 {
		conn, err := amqp.Dial("amqp://localhost:5672/")
		if err != nil {
			retries--
			fmt.Println("Retrying Search engine service connection")
			time.Sleep(2000 * time.Millisecond)
			return EstablishConnection(retries)
		}
		SetNewConnection("conn", conn)
		return nil
	}

	return fmt.Errorf("Shutting down search engine after serveral retries")
}

func QueryDatabase(message string) {

	ch, err := GetChannel("dbChannel")
	if err != nil {
		log.Panicf("dbChannel does not exist\n")
	}
	err = ch.Publish(
		"",
		SENGINE_DB_REQUEST_QUEUE,
		false, false, amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
			ReplyTo:     DB_SENGINE_REQUEST_CBQ,
		},
	)
	if err != nil {
		log.Panic(err.Error())
	}
}

func PublishScoreRanking(segments [][]byte) {

	ch, err := GetChannel("mainChannel")
	if err != nil {
		log.Panicf("mainChannel does not exist\n")
	}

	fmt.Printf("Sending %d ranked webpage segments\n", len(segments))
	defer fmt.Printf("Successfully sent all %d segments\n", len(segments))
	for i := 0; i < len(segments); i++ {
		err = ch.Publish(
			"",
			SENGINE_EXPRESS_QUERY_CBQ,
			false,
			false,
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        segments[i],
			})

		// TODO Dont panic its organic
		if err != nil {
			log.Panicf(err.Error())
		}
	}

}

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
