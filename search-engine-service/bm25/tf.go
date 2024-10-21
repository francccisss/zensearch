package tfidf

import (
	"search-engine-service/utilities"
	"strings"
)

const (
	k1 = 0.5
	b  = 1 // controlling document normalization
)

func TF(searchQuery string, webpage *utilities.WebpageTFIDF, webpages *[]utilities.WebpageTFIDF) *utilities.WebpageTFIDF {
	currentDocument := &webpage.Contents
	currentDocLength := float64(utilities.DocLength(*currentDocument))
	rawTermCount := float64(strings.Count(strings.ToLower(*currentDocument), strings.ToLower(searchQuery)))

	numerator := rawTermCount * (k1 + 1.0)
	denominator := (rawTermCount + k1) * (1.0 - b + b*currentDocLength/AvgDocLen(webpages))
	webpage.TFScore = numerator / denominator
	return webpage
}

func AvgDocLen(webpages *[]utilities.WebpageTFIDF) float64 {
	totalTermCount := 0
	for i := range *webpages {
		docLength := utilities.DocLength((*webpages)[i].Contents)
		totalTermCount += docLength
	}
	return float64(totalTermCount) / float64(len(*webpages))
}
