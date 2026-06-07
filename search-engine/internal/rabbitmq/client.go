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

// Must always be called right after establishing a connection and before setting up consumer handlers
func (rb *RabbitMQClient) SetDefinitions() error {

	if rb.Connection == nil {
		return fmt.Errorf("SetDefinitions() Connection is not connected after running EstablishConnection ")
	}
	pubCh, err := rb.Connection.Channel()
	if err != nil {
		fmt.Println("From PublishChannel")
		return err
	}
	rb.PublishChannel = pubCh

	highCh, err := rb.Connection.Channel()
	if err != nil {
		fmt.Println("From HighChannel")
		return err
	}
	rb.HighThroughputChannel = highCh
	rb.HighThroughputChannel.Qos(CMLTV_ACK, 0, false)

	eventsCh, err := rb.Connection.Channel()
	if err != nil {
		fmt.Println("From EventsChannel")
		return err
	}
	rb.EventsChannel = eventsCh

	rb.PublishChannel.ExchangeDeclare(rb.Definitions.Exchange.General, "direct", true, false, false, false, nil)

	_, err = rb.PublishChannel.QueueDeclare(rb.Definitions.Queues.SE_DB_REQUEST_QUEUE, true, false, false, false, nil)
	if err != nil {
		fmt.Printf("From Declaring Queue %s\n", rb.Definitions.Queues.SE_DB_REQUEST_QUEUE)
		return err
	}

	_, err = rb.PublishChannel.QueueDeclare(rb.Definitions.Queues.SE_DB_REQUEST_CBQ, true, false, false, false, nil)

	if err != nil {
		fmt.Printf("From Declaring Queue %s\n", rb.Definitions.Queues.SE_DB_REQUEST_CBQ)
		return err
	}

	_, err = rb.PublishChannel.QueueDeclare(rb.Definitions.Queues.ES_SE_QUERY_QUEUE, true, false, false, false, nil)

	if err != nil {
		fmt.Printf("From Declaring Queue %s\n", rb.Definitions.Queues.ES_SE_QUERY_QUEUE)
		return err
	}

	err = rb.PublishChannel.QueueBind(rb.Definitions.Queues.SE_DB_REQUEST_QUEUE, rb.Definitions.RoutingKeys.SE_DB_REQUEST, rb.Definitions.Exchange.General, false, nil)

	err = rb.PublishChannel.QueueBind(rb.Definitions.Queues.ES_SE_QUERY_QUEUE, rb.Definitions.RoutingKeys.ES_SE_QUERY, rb.Definitions.Exchange.General, false, nil)

	if err != nil {

		fmt.Printf("From Binding Queue %s\n", rb.Definitions.Queues.SE_DB_REQUEST_QUEUE)
		return err
	}

	return nil

}

func (rb *RabbitMQClient) EstablishConnection(retries int) error {

	if retries > 0 {
		conn, err := amqp.Dial("amqp://localhost:5672/")
		if err != nil {
			retries--
			time.Sleep(2000 * time.Millisecond)
			return rb.EstablishConnection(retries)
		}
		fmt.Println("Successfully connected to RabbitMQ")
		rb.Connection = conn
		return nil
	}

	return fmt.Errorf("Shutting down search engine after serveral retries")
}

func (rb *RabbitMQClient) QueryDatabase(message string) {
	err := rb.PublishChannel.Publish(
		rb.Definitions.Exchange.General,
		rb.Definitions.RoutingKeys.SE_DB_REQUEST,
		false, false, amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
			ReplyTo:     rb.Definitions.Queues.SE_DB_REQUEST_CBQ,
		},
	)
	if err != nil {
		log.Panic(err.Error())
	}
}

func (rb *RabbitMQClient) DatabaseResponseHandler(webpageBytesChan chan *bytes.Buffer, searchQuery string) {
	serializer := segments.NewSegmentSerializer(rb.HighThroughputChannel)
	for {
		dbMsg, err := rb.HighThroughputChannel.Consume(
			rb.Definitions.Queues.SE_DB_REQUEST_CBQ,
			"",
			false,
			false,
			false,
			false,
			nil,
		)

		if err != nil {
			log.Panicf("Unable to listen to %s", rb.Definitions.Queues.SE_DB_REQUEST_CBQ)
		}

		fmt.Print("Spawn segment listener\n")

		// Listens for incoming segments from the database Query channel consumer
		done, webpageBytes, err := serializer.HandleIncomingSegments(dbMsg)
		select {
		case <-done:
			fmt.Println("Done Handling Segments")
			if err != nil {
				log.Fatalf("Error from Handling Segments: %s", err)
			}
			webpageBytesChan <- &webpageBytes
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
			rb.Definitions.Queues.ES_SE_QUERY_CBQ,
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
