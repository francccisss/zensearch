package bm25

import (
	"search-engine/utilities"
	"strings"
)

const (
	k1 = 4  // controls the weight of term frequency, lower value saturates the term frequency quicker
	b  = .4 // controlling document normalization
)

func TF(searchQuery string, webpages *[]WebpageTFIDF, AvgDocLen float64) error {

	for i := range *webpages {

		currentDocument := (*webpages)[i].Contents
		currentDocLength := float64(utilities.DocLength(currentDocument))
		rawTermCount := float64(strings.Count(strings.ToLower(currentDocument), strings.ToLower(searchQuery)))

		numerator := rawTermCount * (k1 + 1.0)
		denominator := (rawTermCount + k1) * ((1.0 - b + b) * (currentDocLength / AvgDocLen))
		(*webpages)[i].TokenRating.TfRating = numerator / denominator
	}
	return nil
}

func AvgDocLen(webpages *[]WebpageTFIDF) float64 {
	totalTermCount := 0
	for i := range *webpages {
		docLength := utilities.DocLength((*webpages)[i].Contents)
		totalTermCount += docLength
	}
	return float64(totalTermCount) / float64(len(*webpages))
}
