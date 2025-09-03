package frontier

import (
	"crawler/internal/rabbitmq"
	"crawler/internal/types"
	"encoding/binary"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Dequeues the url in the current queue, the domain corresponds to the
// crawler's current running job from a URL seed

type FrontierQueue interface {
	Enqueue(ExtractedUrls) error
	ListenDequeuedUrl()
	Dequeue(string) error
	Len(string) (uint32, error)
	GetChann() chan types.DequeuedUrl
}

type Queue struct {
	QueueChann  chan types.DequeuedUrl
	amqpChannel *amqp.Channel
}

type ExtractedUrls struct {
	Domain string
	Nodes  []string
}

func New() FrontierQueue {

	chann, err := rabbitmq.GetChannel("frontierChannel")
	if err != nil {
		panic(err)
	}
	q := Queue{
		QueueChann:  make(chan types.DequeuedUrl),
		amqpChannel: chann,
	}
	return q

}

func (q Queue) Dequeue(root string) error {

	err := q.amqpChannel.Publish("",
		"crawler_db_dequeue_url_queue",
		false, false,
		amqp.Publishing{
			Body:    []byte(root),
			ReplyTo: rabbitmq.DB_CRAWLER_DEQUEUE_URL_CBQ,
		})

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (q Queue) ListenDequeuedUrl() {
	fmt.Println("Listening to dequeued url")

	msg, err := q.amqpChannel.Consume(rabbitmq.DB_CRAWLER_DEQUEUE_URL_CBQ, "", false, false, false, false, nil)

	if err != nil {
		fmt.Println(err)
		return
	}

	for chanMsg := range msg {
		dq := &types.DequeuedUrl{}
		fmt.Println("CRAWLER TEST: received dequeued URL")
		err = json.Unmarshal(chanMsg.Body, dq)
		if err != nil {
			fmt.Println("ERROR: unable to unmarshal dequeued url")
			fmt.Println(err.Error())
			return
		}
		q.amqpChannel.Ack(chanMsg.DeliveryTag, false)
		q.QueueChann <- *dq
	}
}

func (q Queue) Enqueue(exUrls ExtractedUrls) error {

	fmt.Printf("CRAWLER TEST: ENQUEUING %+d URLS\n", len(exUrls.Nodes))

	b, err := json.Marshal(exUrls)
	if err != nil {
		return err
	}
	err = q.amqpChannel.Publish("", "crawler_db_storeurls_queue", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        b,
	})
	if err != nil {
		return err
	}
	return nil
}

func (q Queue) Len(hostname string) (uint32, error) {
	fmt.Printf("CRAWLER TEST: GETTING QUEUE LENGTH FOR %s\n", hostname)

	// TODO: change queue name

	err := q.amqpChannel.Publish("", "crawler_db_len_queue", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        []byte(hostname),
		ReplyTo:     "get_queue_len_queue",
	})
	if err != nil {
		return 0, err
	}

	lenMsg, err := q.amqpChannel.Consume("get_queue_len_queue", "", false, false, false, false, nil)

	msg := <-lenMsg

	queueLen := binary.LittleEndian.Uint32(msg.Body)

	fmt.Printf("CRAWLER TEST: BODY BUF = %v\n", msg.Body)
	fmt.Printf("CRAWLER TEST: CURRENT QUEUE LEN: %d\n", queueLen)
	q.amqpChannel.Ack(msg.DeliveryTag, false)

	return queueLen, nil
}

func (q Queue) GetChann() chan types.DequeuedUrl {
	return q.QueueChann
}
