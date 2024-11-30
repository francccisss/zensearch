package utilities

import (
	"bytes"
	// "compress/flate"
	"compress/zlib"
	"fmt"
	"io"
)

type compressor interface {
	CompressData()
	DecompressData()
}

type SegmentBuffer struct{}

func NewSegmentBuffer() *SegmentBuffer {
	return &SegmentBuffer{}
}

func (sb SegmentBuffer) CompressData() {

}

// Takes in c as Compressed data and decompresses and returns a buffer
func (sb SegmentBuffer) DecompressData(c []byte) ([]byte, error) {

	var decompressed bytes.Buffer

	r := bytes.NewReader(c)
	fr, err := zlib.NewReader(r)

	if err != nil {
		fmt.Println("Unsupported encoding for zlib decompression")
		return []byte{}, err
	}
	_, err = io.Copy(&decompressed, fr)
	if err != nil {
		fmt.Println("Unable to decompress segment buffer")
		return []byte{}, err
	}

	return decompressed.Bytes(), nil

}
