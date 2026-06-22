package bm25

import (
	"bytes"
	_ "encoding/json"
	"fmt"
	"log"
	"os"
	"search-engine/internal/rabbitmq"
	"search-engine/internal/types"
	"search-engine/utilities"
	"sync"
	"testing"
	"time"

	"gopkg.in/yaml.v2"
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

var TEST_QRY = "quote"

func TestProcessParallelism(t *testing.T) {

	client := mockConnection(t)
	timeStart := time.Now()

	webpageBytesChan := make(chan *bytes.Buffer)

	dbMsg, err := client.HighIngressChannel.Consume(
		client.Definitions.Queues.SE_DB_REQUEST_CBQ,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Println("Unable to consume from HighIngressChannel")
		t.Fatal(err.Error())
	}

	go client.DatabaseResponseHandler(webpageBytesChan, dbMsg)

	client.QueryDatabase(TEST_QRY)
	webpage := <-webpageBytesChan
	webpages, err := utilities.ParseWebpages(webpage.Bytes())
	if err != nil {
		log.Println("Unable to parse webpages")
		t.Fatal(err.Error())
	}

	t.Logf("TEST: Time elapsed parsing: %dms\n\n\n", time.Until(timeStart).Abs().Milliseconds())

	fmt.Printf("TEST: Comparing runtime\n\n")
	// results := [][]string{}
	// results = append(results, testResponsetime(TEST_QRY, webpages, CalculateBMRatings),
	// 	testResponsetime(TEST_QRY, webpages, Bm25TestConcurrency), testResponsetime(TEST_QRY, webpages, Bm25TestSequential))
	testResponsetime(TEST_QRY, webpages, CalculateBMRatings)
	// for _, result := range results {
	// fmt.Printf("results=%+v\n", result)
	// }

	for i, l := range *RankBM25Ratings(webpages) {
		if i > 10 {
			break
		}
		fmt.Printf("SAUCE=%s, RATE=%f\n", l.Url, l.Bm25rating)
	}
	t.Logf("TEST: %+v token", TEST_QRY)
	t.Log("TEST: test end")

}

func testResponsetime(term string, webpages *[]types.WebpageTFIDF, method func(term string, webpages *[]types.WebpageTFIDF) *[]types.WebpageTFIDF) []string {
	timings := []string{}
	fmt.Printf("WBP COUNT:%d\n", len(*webpages))
	fmt.Printf("TEST: current token= '%s'\n", term)
	timeStart := time.Now()
	_ = method(term, webpages)
	fmt.Printf("TEST: Time elapsed for old: %dms\n", time.Until(timeStart).Abs().Milliseconds())
	timings = append(timings, fmt.Sprintf("\n - timings=%d, ", time.Until(timeStart).Abs().Milliseconds()))
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

func mockConnection(t *testing.T) *rabbitmq.RabbitMQClient {

	defBuff, err := os.ReadFile("../../../rabbitmq.yml")
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
			ES_SE_QUERY_CBQ     string
		}{
			SE_DB_REQUEST_QUEUE: rbqDef.Queues.SearchEngineQueues.SE_DB_REQUEST_QUEUE,
			SE_DB_REQUEST_CBQ:   rbqDef.Queues.SearchEngineQueues.SE_DB_REQUEST_CBQ,
			ES_SE_QUERY_QUEUE:   rbqDef.Queues.ExpressServerQueues.ES_SE_QUERY_QUEUE,
			ES_SE_QUERY_CBQ:     rbqDef.Queues.ExpressServerQueues.ES_SE_QUERY_CBQ,
		},
		RoutingKeys: struct {
			SE_DB_REQUEST string
			ES_SE_QUERY   string
		}{
			SE_DB_REQUEST: rbqDef.RoutingKeys.SearchEngineKeys.SE_DB_REQUEST,
			ES_SE_QUERY:   rbqDef.RoutingKeys.ExpressServerKeys.ES_SE_QUERY,
		},
	}

	client := rabbitmq.NewRabbitMQClient(searchEngineDef)
	err = client.EstablishConnection(7)
	if err != nil {
		t.Fatalf("Connection does not exist")
	}
	err = client.SetDefinitions()
	if err != nil {
		t.Fatalf("Connection does not exist")
	}

	return &client
}
