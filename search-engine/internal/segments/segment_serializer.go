package Segments

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"search-engine/internal/bm25"
	"search-engine/internal/rabbitmq"
)

type Segment struct {
	Header SegmentHeader
	Data   []byte
}

type SegmentHeader struct {
	SequenceNum   uint32
	TotalSegments uint32
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

			// TODO change this for retransmission dont crash
			return nil, fmt.Errorf("Unexpected sequence number\n")
			// continue
		}

		segmentCounter++
		expectedSequenceNum++

		dbChannel.Ack(segment.DeliveryTag, false)
		webpageBytes = append(webpageBytes, segmentPayload...)

		if segmentCounter == segmentHeader.TotalSegments {
			fmt.Printf("Received all of the segments from Database %d\n", segmentCounter)

			// reset everything
			expectedSequenceNum = 0
			segmentCounter = 0

			break
		}
	}
	return webpageBytes, nil
}

// MSS is the maximum segment size of the bytes to be transported to the express server
func CreateSegments(webpages *[]bm25.WebpageTFIDF, MSS int) ([][]byte, error) {
	// GOB APPENDS METADATA ABOUT THE TYPES THAT ARE ENCODED FOR
	// THE DECODER TO INTERPRET

	serializeWebpages, err := json.Marshal(webpages)
	if err != nil {
		fmt.Println("Unable to Marshal webpages")
		return nil, err
	}

	serializedSegments := [][]byte{}
	serializedWebpagesLen := len(serializeWebpages)
	segmentCount := int(serializedWebpagesLen/MSS) + 1 // for the remainder
	fmt.Printf("Segment Count: %d\n", segmentCount)

	var (
		currentIndex    = 0
		pointerPosition = float64(MSS)
	)
	for i := 0; i < segmentCount; i++ {

		segmentSlice := serializeWebpages[currentIndex:int(pointerPosition)]
		serializedSegments = append(serializedSegments, NewSegment(uint32(i), uint32(segmentCount), segmentSlice))

		currentIndex = int(pointerPosition)

		pointerPosition += math.Min(float64(MSS), float64(serializedWebpagesLen-currentIndex))
	}

	fmt.Printf("Total segments created: %d\n", len(serializedSegments))
	defer fmt.Println("Successfully exited")

	return serializedSegments, nil
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

func NewSegment(sequenceNum uint32, segmentCount uint32, payload []byte) []byte {

	// seqNumBuff := make([]byte, binary.MaxVarintLen32)
	// binary.LittleEndian.PutUint32(seqNumBuff, uint32(sequenceNum))

	// for some reason it appends another byte of 0 before the segmenCount
	// eg: [231,0,0,0,0,233,0,0]
	//                ^ What is that??!!?!
	// header := binary.LittleEndian.AppendUint32(seqNumBuff, segmentCount)

	// Im gonna do what's called a pro gamer move
	// DO THIS JUST FOR NOW
	// TODO CHANGE THIS OR ELSE
	segment := append([]byte{byte(sequenceNum), 0, 0, 0, byte(segmentCount), 0, 0, 0}, payload...)

	return segment
}

func SendSegments() {

}

func GetSegmentHeader(buf []byte) (*SegmentHeader, error) {
	byteReader := bytes.NewBuffer(buf)
	headerOffsets := []int{0, 4}
	var newSegmentHeader SegmentHeader

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
