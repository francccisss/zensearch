package bm25

import (
	"bytes"
	"encoding/json"
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

const TEST_QRY = "threads"

func TestProcessParallelism(t *testing.T) {

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

	// send database query
	rabbitmq.QueryDatabase(TEST_QRY)

	t.Log("TEST: Waiting for webpage handler to finish")
	webpageBuffer := <-webpageBytesChan

	t.Log("TEST: Parsing and rating calculation starting...")

	timeStart := time.Now()
	// compressor := util.NewSegmentBuffer()
	// decompressed, err := compressor.DecompressData(webpageBuffer)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	continue
	// }
	webpages, err := utilities.ParseWebpages(webpageBuffer.Bytes())
	if err != nil {
		log.Println("Unable to parse webpages")
		t.Fatal(err.Error())
	}
	t.Logf("TEST: Time elapsed parsing: %dms\n", time.Until(timeStart).Abs().Milliseconds())

	// Ranking webpages
	timeStart = time.Now()

	calculatedRatings := CalculateBMRatings(TEST_QRY, webpages)
	rankedWebpages := RankBM25Ratings(calculatedRatings)

	if len((*rankedWebpages)) > 0 {
		b, err := json.Marshal((*rankedWebpages)[0])
		if err != nil {
			t.Fatal(err.Error())
		}
		t.Logf("TEST: 1st webpage title: %+v\n", (*rankedWebpages)[0].Url)
		t.Logf("TEST: 2nd webpage title: %+v\n", (*rankedWebpages)[1].Url)
		t.Logf("TEST: webpage tf rating: %+v\n", (*rankedWebpages)[0].TfRating)
		t.Logf("TEST: webpage bm rating: %+v\n", (*rankedWebpages)[0].Bm25rating)
		t.Logf("TEST: webpage size: %dkb, %db", len(b)/1024, len(b))
	}
	t.Logf("TEST: Total ranked webpages: %d\n", len(*rankedWebpages))
	t.Logf("TEST: Time elapsed ranking: %dms\n", time.Until(timeStart).Abs().Milliseconds())
}

var m *sync.Mutex

func Bm25TestRatings(query string, webpages *[]types.WebpageTFIDF) *[]types.WebpageTFIDF {
	tokenizedTerms := Tokenizer(query)
	fmt.Println(tokenizedTerms)
	docLen := utilities.AvgDocLen(webpages)
	for _, term := range tokenizedTerms {
		// IDF is a constant throughout the current term
		IDF := CalculateIDF(term, webpages)
		_ = TF(term, webpages, docLen, m)
		for j := range *webpages {
			bm25rating := BM25(IDF, (*webpages)[j].TfRating)
			(*webpages)[j].TokenRating.Bm25rating += bm25rating
		}
	}
	return webpages
}

func Bm25TestRatingsConcurrency(query string, webpages *[]types.WebpageTFIDF) *[]types.WebpageTFIDF {
	fmt.Println("\n\nTEST: Process Pattern")
	tokenizedTerms := Tokenizer(query)
	fmt.Println(tokenizedTerms)

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
		fmt.Println("TEST: Finished getting IDF rating for each token")
		wg.Done()
	}()

	go func() {
		docLen := utilities.AvgDocLen(webpages)
		for _, term := range tokenizedTerms {
			// IDF is a constant throughout the current term
			// Dont need to return, it uses the address of the webpages
			// First calculate term frequency of each webpage for each token
			// TF(q1,webpages) -> TF(qT2,webpages)...
			_ = TF(term, webpages, docLen, m)
		}
		fmt.Println("TEST: Finished calculating and applying TF rating of each token to webpages")
		wg.Done()
	}()

	fmt.Println("TEST: waiting for TF and IDF calculations")
	wg.Wait()
	// for each token calculate BM25Rating for each webpages
	// by summing the rating from the previous tokens
	fmt.Println("TEST: calculating bm25 rating")
	for IDF := range IDFChan {
		for j := range *webpages {
			bm25rating := BM25(IDF, (*webpages)[j].TfRating)
			(*webpages)[j].TokenRating.Bm25rating += bm25rating
		}
	}
	return webpages
}
