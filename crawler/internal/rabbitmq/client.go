package rabbitmq

import (
	"context"
	"crawler/internal/types"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQClient struct {
	Connection     *amqp.Connection
	PublishChannel *amqp.Channel
	EventsChannel  *amqp.Channel
	Definitions    CrawlerDefinitions
}

func NewRabbitMQClient(def CrawlerDefinitions) RabbitMQClient {
	rb := RabbitMQClient{
		Connection:     nil,
		PublishChannel: nil,
		EventsChannel:  nil,
		Definitions:    def,
	}
	return rb
}

func (rb *RabbitMQClient) EstablishConnection(retries int) error {

	if retries > 0 {
		conn, err := amqp.Dial("amqp://localhost:5672/")
		if err != nil {
			retries--
			fmt.Println("Retrying Crawler service connection")
			time.Sleep(2000 * time.Millisecond)
			return rb.EstablishConnection(retries)
		}
		rb.Connection = conn
		return nil
	}

	return fmt.Errorf("Shutting down crawler service after serveral retries")
}

func (rb *RabbitMQClient) SetDefinitions() error {

	pubCh, err := rb.Connection.Channel()
	if err != nil {
		return err
	}
	rb.PublishChannel = pubCh

	eventCh, err := rb.Connection.Channel()
	if err != nil {
		return err
	}
	rb.EventsChannel = eventCh

	err = rb.PublishChannel.ExchangeDeclare(rb.Definitions.Exchange.Crawler, "direct", true, false, false, false, nil)
	if err != nil {
		return err
	}
	fmt.Println("Exchange Crawler Ok")
	err = rb.PublishChannel.ExchangeDeclare(rb.Definitions.Exchange.General, "direct", true, false, false, false, nil)
	if err != nil {
		return err
	}
	fmt.Println("Exchange General Ok")

	// CR_DB QUEUSE ARE BOUND TO CRAWLER EXCHANGE

	rb.PublishChannel.QueueDeclare(rb.Definitions.Queues.ES_CR_REQUEST_QUEUE, true, false, false, false, nil)
	rb.PublishChannel.QueueDeclare(rb.Definitions.Queues.CR_DB_INDEXING_QUEUE, true, false, false, false, nil)
	rb.PublishChannel.QueueDeclare(rb.Definitions.Queues.CR_DB_ENQUEUE_QUEUE, true, false, false, false, nil)
	rb.PublishChannel.QueueDeclare(rb.Definitions.Queues.CR_DB_DEQUEUE_QUEUE, true, false, false, false, nil)
	rb.PublishChannel.QueueDeclare(rb.Definitions.Queues.CR_DB_GETLEN_QUEUE, true, false, false, false, nil)

	rb.PublishChannel.QueueDeclare(rb.Definitions.Queues.ES_CR_REQUEST_CBQ, true, false, false, false, nil)
	rb.PublishChannel.QueueDeclare(rb.Definitions.Queues.CR_DB_INDEXING_CBQ, true, false, false, false, nil)
	rb.PublishChannel.QueueDeclare(rb.Definitions.Queues.CR_DB_ENQUEUE_CBQ, true, false, false, false, nil)
	rb.PublishChannel.QueueDeclare(rb.Definitions.Queues.CR_DB_DEQUEUE_CBQ, true, false, false, false, nil)
	rb.PublishChannel.QueueDeclare(rb.Definitions.Queues.CR_DB_GETLEN_CBQ, true, false, false, false, nil)

	rb.PublishChannel.QueueBind(rb.Definitions.Queues.ES_CR_REQUEST_QUEUE, rb.Definitions.RoutingKeys.ES_CR_REQUEST, rb.Definitions.Exchange.General, false, nil)
	rb.PublishChannel.QueueBind(rb.Definitions.Queues.CR_DB_INDEXING_QUEUE, rb.Definitions.RoutingKeys.CR_DB_INDEXING, rb.Definitions.Exchange.Crawler, false, nil)
	rb.PublishChannel.QueueBind(rb.Definitions.Queues.CR_DB_ENQUEUE_QUEUE, rb.Definitions.RoutingKeys.CR_DB_ENQUEUE, rb.Definitions.Exchange.Crawler, false, nil)
	rb.PublishChannel.QueueBind(rb.Definitions.Queues.CR_DB_DEQUEUE_QUEUE, rb.Definitions.RoutingKeys.CR_DB_DEQUEUE, rb.Definitions.Exchange.Crawler, false, nil)
	rb.PublishChannel.QueueBind(rb.Definitions.Queues.CR_DB_GETLEN_QUEUE, rb.Definitions.RoutingKeys.CR_DB_GETLEN, rb.Definitions.Exchange.Crawler, false, nil)

	return nil
}

func (rb *RabbitMQClient) HandleIncomingUrls(ctx context.Context, list types.CrawlList) {
}
