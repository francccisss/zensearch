package Segments

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"math"
	"search-engine/internal/bm25"
)

type Segment struct {
	Header  SegmentHeader
	Payload []byte
}

type SegmentHeader struct {
	SequenceNum   uint32
	TotalSegments uint32
}

func DecodeSegments(newSegment amqp.Delivery) (Segment, error) {

	segmentHeader, err := GetSegmentHeader(newSegment.Body)
	if err != nil {
		fmt.Println("Unable to extract segment header")
		return Segment{}, err
	}

	segmentPayload, err := GetSegmentPayload(newSegment.Body)
	if err != nil {
		fmt.Println("Unable to extract segment payload")
		return Segment{}, err
	}

	return Segment{Header: *segmentHeader, Payload: segmentPayload}, nil
}

// MSS is the maximum segment size of the bytes to be transported to the express server
func CreateSegments(webpages *[]bm25.WebpageTFIDF, MSS int) ([][]byte, error) {
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
