package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"
	"search-engine-service/utilities"

	amqp "github.com/rabbitmq/amqp091-go"
)

const QUERYQUEUE = "database_query_queue"
const PUBLISH_QUEUE = "publish_ranking_queue"

func QueryDatabase(message string, ch *amqp.Channel) {
	err := ch.Publish(

		"",
		QUERYQUEUE,
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

func PublishScoreRanking(rankedWebpages *[]utilities.WebpageTFIDF, ch *amqp.Channel) {
	ch.QueueDeclare(PUBLISH_QUEUE, false, false, false, false, nil)
	encodedWebpages, err := json.Marshal(rankedWebpages)
	if err != nil {
		log.Panicf(err.Error())
	}
	err = ch.Publish(
		"",
		QUERYQUEUE,
		false, false, amqp.Publishing{
			ContentType: "text/plain",
			Body:        encodedWebpages,
		})
	if err != nil {
		log.Panicf(err.Error())
	}

}
