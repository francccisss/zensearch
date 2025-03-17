package bm25

import (
	"bytes"
	_ "encoding/json"
	"fmt"
	"log"
	"search-engine/internal/rabbitmq"
	segments "search-engine/internal/segment_serializer"
	"search-engine/internal/types"
	"search-engine/utilities"
	"sync"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

/*
 This test is used for comparing my previous iterations of calculating
 bm25 score rating, the current one uses, parallelism, previous ones
 uses basic concurrency and barrier implementation, last was sequential
 when running test make sure database and rabbitmq servers are connected

 as the terms of the user query grows, where each term is reffered to as token
 it runs through each token, concatenating the previous ones and passing it as
 an argument to each BM25ranking methods
*/

var TEST_QRY = []string{"Using", "golang", "makes", "it", "so", "simple", "to", "implement", "concurrency"}

func TestProcessParallelism(t *testing.T) {

	webpageBuffer := mockConnection(t)
	timeStart := time.Now()
	webpages, err := utilities.ParseWebpages(webpageBuffer.Bytes())
	if err != nil {
		log.Println("Unable to parse webpages")
		t.Fatal(err.Error())
	}
	t.Logf("TEST: Time elapsed parsing: %dms\n\n\n", time.Until(timeStart).Abs().Milliseconds())

	fmt.Printf("TEST: Comparing runtime\n\n")
	results := [][]string{}
	results = append(results, testResponsetime(TEST_QRY, webpages, CalculateBMRatings),
		testResponsetime(TEST_QRY, webpages, Bm25TestConcurrency), testResponsetime(TEST_QRY, webpages, Bm25TestSequential))

	for _, result := range results {
		fmt.Printf("results=%+v\n", result)
	}

	t.Logf("TEST: %+v token", TEST_QRY)
	t.Log("TEST: test end")

}

func testResponsetime(termTokens []string,
	webpages *[]types.WebpageTFIDF,
	method func(term string, webpages *[]types.WebpageTFIDF) *[]types.WebpageTFIDF) []string {
	timings := []string{}
	terms := ""
	for _, tt := range termTokens {
		terms += tt + " "
		fmt.Printf("TEST: current token= '%s'\n", terms)
		timeStart := time.Now()
		_ = method(terms, webpages)
		fmt.Printf("TEST: Time elapsed for old: %dms\n", time.Until(timeStart).Abs().Milliseconds())
		timings = append(timings, fmt.Sprintf("\n - timings=%d, ", time.Until(timeStart).Abs().Milliseconds()))
	}
	return timings
}

var m *sync.Mutex

func Bm25TestSequential(query string, webpages *[]types.WebpageTFIDF) *[]types.WebpageTFIDF {
	fmt.Println("TEST: sequential")
	tokenizedTerms := Tokenizer(query)
	docLen := utilities.AvgDocLen(webpages)
	for _, term := range tokenizedTerms {
		// IDF is a constant throughout the current term
		IDF := CalculateIDF(term, webpages)
		_ = TF(term, docLen, webpages, 0, len(*webpages))
		for j := range *webpages {
			bm25rating := BM25(IDF, (*webpages)[j].TfRating)
			(*webpages)[j].TokenRating.Bm25rating += bm25rating
		}
	}
	return webpages
}

func Bm25TestConcurrency(query string, webpages *[]types.WebpageTFIDF) *[]types.WebpageTFIDF {
	fmt.Println("TEST: MS/TP Pattern")
	tokenizedTerms := Tokenizer(query)

	var wg sync.WaitGroup
	// get IDF and TF for each token
	IDFChan := make(chan float64, 10)
	// TODO do master slave, aggregate results back to go master routine

	wg.Add(1)
	wg.Add(1)
	go func() {
		for i := range tokenizedTerms {
			// IDF is a constant throughout the current term
			IDF := CalculateIDF(tokenizedTerms[i], webpages)
			IDFChan <- IDF
		}
		close(IDFChan)
		wg.Done()
	}()

	go func() {
		docLen := utilities.AvgDocLen(webpages)
		for _, term := range tokenizedTerms {
			// IDF is a constant throughout the current term
			// Dont need to return, it uses the address of the webpages
			// First calculate term frequency of each webpage for each token
			// TF(q1,webpages) -> TF(qT2,webpages)...
			_ = TF(term, docLen, webpages, 0, len(*webpages))
		}
		wg.Done()
	}()

	wg.Wait()
	// for each token calculate BM25Rating for each webpages
	// by summing the rating from the previous tokens
	for IDF := range IDFChan {
		for j := range *webpages {
			bm25rating := BM25(IDF, (*webpages)[j].TfRating)
			(*webpages)[j].TokenRating.Bm25rating += bm25rating
		}
	}
	return webpages
}

func mockConnection(t *testing.T) bytes.Buffer {
	incomingSegmentsChan := make(chan amqp.Delivery)
	webpageBytesChan := make(chan bytes.Buffer)
	err := rabbitmq.EstablishConnection(7)

	conn, err := rabbitmq.GetConnection("conn")
	if err != nil {
		t.Fatalf("Connection does not exist")
	}
	dbQueryChannel, err := conn.Channel()
	if err != nil {
		t.Fatalf("Unable to create a database channel")
	}
	_, err = dbQueryChannel.QueueDeclare(rabbitmq.DB_QUERY_QUEUE, false, false, false, false, nil)
	if err != nil {
		t.Fatalf("Unable to declare DB_QUERY_QUEUE")
	}
	_, err = dbQueryChannel.QueueDeclare(rabbitmq.DB_RESPONSE_QUEUE, false, false, false, false, nil)
	if err != nil {
		t.Fatalf("Unable to declare DB_RESPONSE_QUEUE")
	}

	rabbitmq.SetNewChannel("dbChannel", dbQueryChannel)
	// spanw segment handler
	go segments.HandleIncomingSegments(dbQueryChannel, incomingSegmentsChan, webpageBytesChan)

	// spawn listener
	go func(chann *amqp.Channel) {

		dbMsg, err := chann.Consume(
			rabbitmq.DB_RESPONSE_QUEUE,
			"",
			false,
			false,
			false,
			false,
			nil,
		)

		if err != nil {
			log.Panicf("Unable to listen to %s", rabbitmq.SEARCH_QUEUE)
		}

		// Consume and send segment to segment channel
		for incomingSegment := range dbMsg {
			incomingSegmentsChan <- incomingSegment
		}

	}(dbQueryChannel)

	const TEST_QRY = "semaphore is really good"
	// send database query
	rabbitmq.QueryDatabase("nothing burger")

	t.Log("TEST: Waiting for webpage handler to finish")
	webpageBuffer := <-webpageBytesChan

	t.Log("TEST: Parsing and rating calculation starting...")
	return webpageBuffer
}
