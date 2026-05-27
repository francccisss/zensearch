package rabbitmq

import (
	"bytes"
	"fmt"
	"log"
	segments "search-engine/internal/segment_serializer"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQClient struct {
	Connection            *amqp.Connection
	PublishChannel        *amqp.Channel
	HighThroughputChannel *amqp.Channel // for returning search results
	EventsChannel         *amqp.Channel
	Definitions           SearchEngineDefinitions
}

const CMLTV_ACK = 1000

func NewRabbitMQClient(def SearchEngineDefinitions) RabbitMQClient {
	rb := RabbitMQClient{
		Connection:            nil,
		PublishChannel:        nil,
		EventsChannel:         nil,
		HighThroughputChannel: nil,
		Definitions:           def,
	}
	return rb
}

func (rb *RabbitMQClient) SetDefinitions() error {

	pubCh, err := rb.Connection.Channel()
	if err != nil {
		return err
	}
	rb.PublishChannel = pubCh

	highCh, err := rb.Connection.Channel()
	if err != nil {
		return err
	}
	rb.HighThroughputChannel = highCh
	rb.HighThroughputChannel.Qos(CMLTV_ACK, 0, false)

	eventsCh, err := rb.Connection.Channel()
	if err != nil {
		return err
	}
	rb.EventsChannel = eventsCh
	_, err = pubCh.QueueDeclare(rb.Definitions.Queues.SE_DB_REQUEST_QUEUE, true, false, true, false, nil)

	_, err = pubCh.QueueDeclare(rb.Definitions.Queues.SE_DB_REQUEST_CBQ, true, false, true, false, nil)

	if err != nil {
		return err
	}

	err = pubCh.QueueBind(rb.Definitions.Queues.SE_DB_REQUEST_QUEUE, rb.Definitions.RoutingKeys.SE_DB_REQUEST, rb.Definitions.Exchange.General, false, nil)

	if err != nil {
		return err
	}

	return nil

}

func (rb *RabbitMQClient) EstablishConnection(retries int) error {

	if retries > 0 {
		conn, err := amqp.Dial("amqp://localhost:5672/")
		if err != nil {
			retries--
			fmt.Println("Retrying Search engine service connection")
			time.Sleep(2000 * time.Millisecond)
			return rb.EstablishConnection(retries)
		}
		rb.Connection = conn
		return nil
	}

	return fmt.Errorf("Shutting down search engine after serveral retries")
}

func (rb *RabbitMQClient) QueryDatabase(message string) {

	err := rb.PublishChannel.Publish(
		"",
		rb.Definitions.Queues.SE_DB_REQUEST_QUEUE,
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

func (rb *RabbitMQClient) DatabaseResponseHandler(webpageBytesChan chan bytes.Buffer, searchQuery string) {
	for {
		dbMsg, err := rb.HighThroughputChannel.Consume(
			rb.Definitions.Queues.SE_DB_REQUEST_CBQ,
			"",
			false,
			true,
			false,
			false,
			nil,
		)

		if err != nil {
			log.Panicf("Unable to listen to %s", rb.Definitions.Queues.SE_DB_REQUEST_CBQ)
		}

		fmt.Print("Query database\n")
		// Queries database to send segments to search engine

		fmt.Print("Spawn segment listener\n")

		// Listens for incoming segments from the database Query channel consumer
		done, webpageBytes, err := segments.HandleIncomingSegments(rb.HighThroughputChannel, dbMsg)
		select {
		case <-done:
			if err != nil {
				log.Fatalf("Error from Handling Segments: %s", err)
			}
			webpageBytesChan <- webpageBytes
			fmt.Printf("Clean up Handler for %s search query\n", searchQuery)
			return
		default:
			continue
		}
	}

}

func (rb *RabbitMQClient) PublishScoreRanking(segments [][]byte) {

	fmt.Printf("Sending %d ranked webpage segments\n", len(segments))
	defer fmt.Printf("Successfully sent all %d segments\n", len(segments))
	for i := range len(segments) {
		err := rb.HighThroughputChannel.Publish(
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
