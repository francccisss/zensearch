package rabbitmq

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

func QueryDatabase(message string) {

	ch, err := GetChannel("dbChannel")
	if err != nil {
		log.Panicf("dbChannel does not exist\n")
	}
	err = ch.Publish(
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
			PUBLISH_QUEUE,
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