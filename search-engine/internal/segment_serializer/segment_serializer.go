package segments

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"search-engine/constants"
	"search-engine/internal/types"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Segment struct {
	Header  SegmentHeader
	Payload []byte
}

type SegmentHeader struct {
	SequenceNum   uint32
	TotalSegments uint32
}

// waits for all of the incoming segments and decodes then appends bytes into the `webpageBytesChan`
// incoming segment is the input while the webpageBytesChan is the output
func HandleIncomingSegments(dbChannel *amqp.Channel, incomingSegmentsChan <-chan amqp.Delivery, webpageBytesChan chan bytes.Buffer) {

	var (
		segmentCounter      uint32 = 0
		expectedSequenceNum uint32 = 0
	)

	timeStart := time.Now()
	var webpageBytes bytes.Buffer
	for newSegment := range incomingSegmentsChan {

		segment, err := DecodeSegments(newSegment.Body)
		if err != nil {
			log.Panicf("Unable to decode segments")
			return
		}

		if segment.Header.SequenceNum != expectedSequenceNum {
			dbChannel.Nack(newSegment.DeliveryTag, true, true)
			fmt.Printf("Expected Sequence number %d, got %d\n",
				expectedSequenceNum, segment.Header.SequenceNum)

			// TODO change this for retransmission dont crash
			log.Panicf("Unexpected sequence number\n")
			// continue
		}

		segmentCounter++
		expectedSequenceNum++

		if segmentCounter%constants.CMLTV_ACK == 0 {
			fmt.Println("Ack all prior messages from")
			dbChannel.Ack(newSegment.DeliveryTag, true)
		}
		webpageBytes.Write(segment.Payload)

		if segmentCounter == segment.Header.TotalSegments {
			fmt.Println("Ack all prior messages")
			dbChannel.Ack(newSegment.DeliveryTag, true)
			fmt.Printf("Received all of the segments from Database %d\n", segmentCounter)
			// reset everything
			expectedSequenceNum = 0
			segmentCounter = 0
			break
		}
	}
	webpageBytesChan <- webpageBytes
	fmt.Printf("Time elapsed Listening to segments: %dms\n", time.Until(timeStart).Abs().Milliseconds())
}

func DecodeSegments(newSegment []byte) (Segment, error) {

	segmentHeader, err := GetSegmentHeader(newSegment[:8])
	if err != nil {
		fmt.Println("Unable to extract segment header")
		return Segment{}, err
	}

	segmentPayload, err := GetSegmentPayload(newSegment)
	if err != nil {
		fmt.Println("Unable to extract segment payload")
		return Segment{}, err
	}

	return Segment{Header: *segmentHeader, Payload: segmentPayload}, nil
}

func GetSegmentHeader(buf []byte) (*SegmentHeader, error) {
	var newSegmentHeader SegmentHeader
	newSegmentHeader.SequenceNum = binary.LittleEndian.Uint32(buf[:4])
	newSegmentHeader.TotalSegments = binary.LittleEndian.Uint32(buf[4:])
	return &newSegmentHeader, nil
}

func GetSegmentPayload(buf []byte) ([]byte, error) {
	headerOffset := 8
	byteReader := bytes.NewBuffer(buf[headerOffset:])
	return byteReader.Bytes(), nil
}

// MSS is the maximum segment size of the bytes to be transported to the express server
func CreateSegments(webpages *[]types.WebpageTFIDF, MSS int) ([][]byte, error) {
	serializeWebpages, err := json.Marshal(webpages)
	if err != nil {
		fmt.Println("Unable to Marshal webpages")
		return nil, err
	}

	serializedSegments := [][]byte{}
	serializedWebpagesLen := len(serializeWebpages)
	segmentCount := int(serializedWebpagesLen/MSS) + 1 // for the remainder
	fmt.Printf("Total segment to be created: %d\n", segmentCount)

	var (
		currentIndex = 0

		// set the position before starting the loop to determine so that if ever the bytes
		// are less than the MSS then we can adjust it before hand
		pointerPosition = math.Min(float64(MSS), float64(serializedWebpagesLen-currentIndex))
	)

	for i := 0; i < segmentCount; i++ {

		segmentSlice := serializeWebpages[currentIndex:int(pointerPosition)]

		currentIndex = int(pointerPosition)

		pointerPosition += math.Min(float64(MSS), float64(serializedWebpagesLen-currentIndex))
		serializedSegments = append(serializedSegments, NewSegment(uint32(i), uint32(segmentCount), segmentSlice))
	}
	fmt.Printf("Total segments created: %d\n", len(serializedSegments))

	return serializedSegments, nil
}

func NewSegment(sequenceNum uint32, segmentCount uint32, payload []byte) []byte {

	headerBuf := make([]byte, 4)

	binary.LittleEndian.PutUint32(headerBuf, sequenceNum)

	header := binary.LittleEndian.AppendUint32(headerBuf, segmentCount)

	// Im gonna do what's called a pro gamer move
	// DO THIS JUST FOR NOW
	// TODO CHANGE THIS OR ELSE
	var segmentBuff bytes.Buffer
	segmentBuff.Write(header)
	segmentBuff.Write(payload)

	return segmentBuff.Bytes()
}
