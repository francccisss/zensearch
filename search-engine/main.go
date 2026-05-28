package main

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"search-engine/constants"
	"search-engine/internal/bm25"
	"search-engine/internal/rabbitmq"
	"search-engine/internal/segment_serializer"
	"search-engine/utilities"
	"time"
)

// TODO SYSTEM ERRORS SHOULD RESTART THE SERVICE... I DONT KNOW HOW TO DO IT

// Maximum segment size in bytes

func main() {

	defBuff, err := os.ReadFile("../rabbitmq.yml")
	if err != nil {
		panic(err)
	}
	if len(defBuff) == 0 {
		panic("Empty config file")
	}
	fmt.Printf("DefBuff len: %d\n", len(defBuff))
	var searchEngineDef rabbitmq.SearchEngineDefinitions
	var rbqDef rabbitmq.RabbitMQDefinitions
	err = yaml.Unmarshal(defBuff, &rbqDef)
	if err != nil {
		panic(err)
	}

	searchEngineDef = rabbitmq.SearchEngineDefinitions{
		Exchange: rbqDef.Exchange,
		Queues: struct {
			SE_DB_REQUEST_QUEUE string
			SE_DB_REQUEST_CBQ   string
			ES_SE_QUERY_QUEUE   string
		}{
			SE_DB_REQUEST_QUEUE: rbqDef.Queues.SearchEngineQueues.SE_DB_REQUEST_QUEUE,
			SE_DB_REQUEST_CBQ:   rbqDef.Queues.SearchEngineQueues.SE_DB_REQUEST_CBQ,
			ES_SE_QUERY_QUEUE:   rbqDef.Queues.ExpressServerQueues.ES_SE_QUERY_QUEUE,
		},
		RoutingKeys: struct {
			SE_DB_REQUEST string
		}{
			SE_DB_REQUEST: rbqDef.RoutingKeys.SearchEngineKeys.SE_DB_REQUEST,
		},
	}

	fmt.Printf("definition: :%+v\n", searchEngineDef)

	client := rabbitmq.NewRabbitMQClient(searchEngineDef)

	err = client.EstablishConnection(7)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	err = client.SetDefinitions()
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("Search engine established TCP Connection with RabbitMQ")

	defer client.Connection.Close()

	for {

		msgs, err := client.EventsChannel.Consume(
			client.Definitions.Queues.ES_SE_QUERY_QUEUE,
			"", false, false, false, false, nil)
		if err != nil {
			log.Fatalf("Error from Consume %s", err)
		}
		newMsg := <-msgs

		err = client.EventsChannel.Ack(newMsg.DeliveryTag, true)
		if err != nil {
			log.Fatalf("Error from Ack %s", err)
		}

		webpageBytesChan := make(chan bytes.Buffer)
		fmt.Printf("User's Query: %s\n", newMsg.Body)

		client.QueryDatabase(string(newMsg.Body))

		go client.DatabaseResponseHandler(webpageBytesChan, string(newMsg.Body))

		// // Handling search engine logic for parsing webpage to json, ranking and data segmentation for transpotation
		go func() {

			// TODO THROW ERRORS TO FRONT END
			for webpageBuffer := range webpageBytesChan {
				// Parsing webpages

				timeStart := time.Now()
				webpages, err := utilities.ParseWebpages(webpageBuffer.Bytes())
				if err != nil {
					fmt.Println(err.Error())
					log.Println("Unable to parse webpages")
					continue
				}
				fmt.Printf("Time elapsed parsing: %dms\n", time.Until(timeStart).Abs().Milliseconds())

				// Ranking webpages
				timeStart = time.Now()

				calculatedRatings := bm25.CalculateBMRatings(string(newMsg.Body), webpages)
				rankedWebpages := bm25.RankBM25Ratings(calculatedRatings)

				fmt.Printf("Total ranked webpages: %d\n", len(*rankedWebpages))
				fmt.Printf("Time elapsed ranking: %dms\n", time.Until(timeStart).Abs().Milliseconds())

				// create segments in this section after ranking
				timeStart = time.Now()
				segments, err := segments.CreateSegments(rankedWebpages, constants.MSS)
				if err != nil {
					fmt.Println(err.Error())
					log.Println("Unable to create segments")
					continue
				}

				fmt.Printf("Time elapsed data segmentation: %dms\n", time.Until(timeStart).Abs().Milliseconds())
				go client.PublishScoreRanking(segments)

			}

		}()
	}

}

// TODO Instead of panicking, create a recursive retry and then close application
func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err.Error())
	}
}
