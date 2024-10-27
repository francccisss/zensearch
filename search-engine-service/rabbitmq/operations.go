package rabbitmq

import (
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

func QueryDatabase(message string, ch *amqp.Channel) {
	err := ch.Publish(
		"",
		DB_QUERY_QUEUE,
		false, false, amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
	fmt.Printf("Push message to database service.\n")
	if err != nil {
		log.Panicf(err.Error())
	}
	log.Printf("End of Query\n")
}

func PublishScoreRanking(rankedWebpages any, ch *amqp.Channel) {
	ch.QueueDeclare(PUBLISH_QUEUE, false, false, false, false, nil)
	encodedWebpages, err := json.Marshal(rankedWebpages)
	if err != nil {
		log.Panicf(err.Error())
	}
	err = ch.Publish(
		"",
		PUBLISH_QUEUE,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        encodedWebpages,
		})
	// TODO Dont panic its organic
	if err != nil {
		log.Panicf(err.Error())
	}
}
