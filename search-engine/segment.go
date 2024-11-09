package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"search-engine/rabbitmq"
	"search-engine/utilities"
)

type Segment struct {
	Header SegmentHeader
	Data   []byte
}

type SegmentHeader struct {
	SequenceNum   uint32
	TotalSegments uint32
}

type TokenRating struct {
	Bm25rating float64
	TfRating   float64
	IdfRating  float64
}

func ListenIncomingSegments(searchQuery string) ([]byte, error) {

	dbChannel, err := rabbitmq.GetChannel("dbChannel")
	if err != nil {
		log.Panicf("dbChannel does not exist please restart the application\n")
	}

	dbMsg, err := dbChannel.Consume(
		rabbitmq.DB_RESPONSE_QUEUE,
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	var (
		segmentCounter      uint32 = 0
		expectedSequenceNum uint32 = 0
	)

	webpageBytes := []byte{}
	defer func(webwebpageBytes *[]byte) {
		*webwebpageBytes = nil
	}(&webpageBytes)

	for {
		segment := <-dbMsg

		segmentHeader, err := GetSegmentHeader(segment.Body)
		if err != nil {
			fmt.Println("Unable to extract segment header")
			return nil, err
		}

		segmentPayload, err := GetSegmentPayload(segment.Body)
		if err != nil {
			fmt.Println("Unable to extract segment payload")
			return nil, err
		}

		// for retransmission/requeuing
		if segmentHeader.SequenceNum != expectedSequenceNum {
			dbChannel.Nack(segment.DeliveryTag, true, true)
			fmt.Printf("Expected Sequence number %d, got %d\n",
				expectedSequenceNum, segmentHeader.SequenceNum)
			continue
		}

		segmentCounter++
		expectedSequenceNum++

		dbChannel.Ack(segment.DeliveryTag, false)
		webpageBytes = append(webpageBytes, segmentPayload...)

		if segmentCounter == segmentHeader.TotalSegments {
			log.Printf("Received all of the segments from Database %d", segmentCounter)

			// reset everything
			expectedSequenceNum = 0
			segmentCounter = 0

			break
		}
	}
	return webpageBytes, nil
}

// MSS is the maximum segment size of the bytes to be transported to the express server
func CreateSegments(webpages *[]utilities.WebpageTFIDF, MSS int) ([][]byte, error) {
	// GOB APPENDS METADATA ABOUT THE TYPES THAT ARE ENCODED FOR
	// THE DECODER TO INTERPRET

	serializeWebpages, err := json.Marshal(webpages)
	if err != nil {
		fmt.Println("Unable to Marshal webpages")
		log.Panicf(err.Error())
	}
	if err != nil {
		fmt.Printf("Unable to serialize webpage arrays for segmentation\n")
		return nil, err
	}

	serializedWebpagesLen := len(serializeWebpages)
	segmentCount := int(serializedWebpagesLen/MSS) + 1 // for the remainder
	fmt.Printf("\nSegment Count: %d\n", segmentCount)

	segments := [][]byte{}

	var (
		currentIndex    = 0
		pointerPosition = float64(MSS)
	)
	for range segmentCount {

		segmentSlice := serializeWebpages[currentIndex:int(pointerPosition)]
		segments = append(segments, NewSegment(segmentSlice))

		currentIndex = int(pointerPosition)

		pointerPosition += math.Min(float64(MSS), float64(serializedWebpagesLen-currentIndex))

		fmt.Printf("\nRemaining Data Length: %d\n", serializedWebpagesLen-currentIndex)
		fmt.Printf("Slice from %d up to %d\n", currentIndex, serializedWebpagesLen-currentIndex)
	}

	fmt.Printf("Segments Length: %d\n", len(segments))
	return segments, nil
}

func readBufferToSlice(buff bytes.Buffer) ([]byte, error) {
	newSlice := make([]byte, buff.Len())
	_, err := buff.Read(newSlice)
	if err != nil {
		fmt.Println("Unable to read buffer to slice")
		return nil, err
	}
	return newSlice, nil
}

func NewSegment(payload []byte) []byte {
	// add headers here
	return payload
}

func SendSegments() {

}

func GetSegmentHeader(buf []byte) (*SegmentHeader, error) {
	byteReader := bytes.NewBuffer(buf)
	headerOffsets := []int{0, 4}
	newSegmentHeader := SegmentHeader{}

	for i := range headerOffsets {
		buffer := make([]byte, 4)
		_, err := byteReader.Read(buffer)
		if err != nil {
			return &SegmentHeader{}, err
		}
		value := binary.LittleEndian.Uint32(buffer)

		// this feels disgusting but i dont feel like bothering with this
		if i == 0 {
			newSegmentHeader.SequenceNum = value
			continue
		}
		newSegmentHeader.TotalSegments = value
	}
	return &newSegmentHeader, nil
}

func GetSegmentPayload(buf []byte) ([]byte, error) {
	headerOffset := 8
	byteReader := bytes.NewBuffer(buf[headerOffset:])
	return byteReader.Bytes(), nil
}
