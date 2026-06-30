package crawler

import (
	"crawler/internal/rabbitmq"
	"encoding/binary"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Dequeues the url in the current queue, the domain corresponds to the
// crawler's current running job from a URL seed

type DequeuedUrl struct {
	RemainingInQueue int
	Url              string
}

type FrontierQueue interface {
	Enqueue(ExtractedUrls) error
	ListenDequeuedUrl()
	Dequeue(string) error
	Len(string) (uint32, error)
	GetChann() chan DequeuedUrl
}

type Queue struct {
	QueueChann chan DequeuedUrl
	RBQClient  rabbitmq.RabbitMQClient
}

type ExtractedUrls struct {
	Root  string
	Nodes []string
}

func (crm *CrawlerManager) NewFrontierQueue() FrontierQueue {

	q := Queue{
		QueueChann: make(chan DequeuedUrl, 1),
		RBQClient:  *crm.RBQClient,
	}
	return q

}

func (q Queue) Dequeue(root string) error {
	fmt.Println("Dequeue request sent")

	err := q.RBQClient.PublishChannel.Publish(q.RBQClient.Definitions.Exchange.Crawler,
		q.RBQClient.Definitions.RoutingKeys.CR_DB_DEQUEUE,
		false, false,
		amqp.Publishing{
			Body:    []byte(root),
			ReplyTo: q.RBQClient.Definitions.Queues.CR_DB_DEQUEUE_CBQ,
		})

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (q Queue) ListenDequeuedUrl() {
	fmt.Println("Listening to dequeued url")
	msg, err := q.RBQClient.EventsChannel.Consume(q.RBQClient.Definitions.Queues.CR_DB_DEQUEUE_CBQ, "", false, false, false, false, nil)
	if err != nil {
		panic("PANIC RAISED FROM QUEUE: CR_DB_DEQUEUE_CBQ")
	}
	for {
		chanMsg := <-msg
		var dq DequeuedUrl
		fmt.Println("CRAWLER TEST: received dequeued URL")
		err = json.Unmarshal(chanMsg.Body, &dq)
		if err != nil {
			// BUG: Ignore the last message error when crawl was successful:
			// At the end of a successful crawl, this somehow receives a message that is unable to be
			// unmarshalled at the end for some reason.
			fmt.Println("ERROR: unable to unmarshal dequeued url")
			fmt.Println(err.Error())
			return
		}
		q.RBQClient.EventsChannel.Ack(chanMsg.DeliveryTag, false)
		fmt.Printf("Interesting: %+v\n", dq)
		q.QueueChann <- dq
	}
}

func (q Queue) Enqueue(exUrls ExtractedUrls) error {

	fmt.Printf("Enqueuing URL %+v\n", exUrls)
	b, err := json.Marshal(exUrls)
	if err != nil {
		return err
	}
	err = q.RBQClient.PublishChannel.Publish(q.RBQClient.Definitions.Exchange.Crawler, q.RBQClient.Definitions.RoutingKeys.CR_DB_ENQUEUE, false, false, amqp.Publishing{
		ReplyTo:     q.RBQClient.Definitions.Queues.CR_DB_DEQUEUE_CBQ,
		ContentType: "application/json",
		Body:        b,
	})
	if err != nil {
		return err
	}
	return nil
}

func (q Queue) Len(hostname string) (uint32, error) {

	err := q.RBQClient.PublishChannel.Publish("", q.RBQClient.Definitions.Queues.CR_DB_GETLEN_QUEUE, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        []byte(hostname),
		ReplyTo:     q.RBQClient.Definitions.Queues.CR_DB_GETLEN_CBQ,
	})
	if err != nil {
		return 0, err
	}

	lenMsg, err := q.RBQClient.EventsChannel.Consume(q.RBQClient.Definitions.Queues.CR_DB_GETLEN_CBQ, "", false, false, false, false, nil)

	msg := <-lenMsg

	queueLen := binary.LittleEndian.Uint32(msg.Body)

	q.RBQClient.EventsChannel.Ack(msg.DeliveryTag, false)

	return queueLen, nil
}

func (q Queue) GetChann() chan DequeuedUrl {
	return q.QueueChann
}
