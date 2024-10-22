package bm25

import (
	"search-engine-service/utilities"
	"strings"
)

const (
	k1 = 1
	b  = .75 // controlling document normalization
)

func TF(searchQuery string, webpages *[]utilities.WebpageTFIDF) error {

	for i := range *webpages {

		currentDocument := (*webpages)[i].Contents
		currentDocLength := float64(utilities.DocLength(currentDocument))
		rawTermCount := float64(strings.Count(strings.ToLower(currentDocument), strings.ToLower(searchQuery)))

		numerator := rawTermCount * (k1 + 1.0)
		denominator := (rawTermCount + k1) * ((1.0 - b + b) * (currentDocLength / avgDocLen(webpages)))
		(*webpages)[i].TokenRating.TfRating = numerator / denominator
	}
	return nil
}

func avgDocLen(webpages *[]utilities.WebpageTFIDF) float64 {
	totalTermCount := 0
	for i := range *webpages {
		docLength := utilities.DocLength((*webpages)[i].Contents)
		totalTermCount += docLength
	}
	return float64(totalTermCount) / float64(len(*webpages))
}
