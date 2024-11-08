package main

import (
	"bytes"
	"encoding/binary"
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
