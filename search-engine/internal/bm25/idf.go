package bm25

import (
	"math"
	"search-engine/internal/types"
	"strings"
)

// Takes O(n)
func CalculateIDF(term string, webpages *[]types.WebpageTFIDF) float64 {

	numberOfDocumentsInCorpa := float64(len(*webpages))
	documentCountWithTerm := float64(termCountInCorpa(term, webpages))
	if documentCountWithTerm == 0 {
		return 0.0
	}
	return math.Log(numberOfDocumentsInCorpa / documentCountWithTerm)
}

func termCountInCorpa(term string, webpages *[]types.WebpageTFIDF) int {
	documentCount := 0

	for i := range *webpages {
		contains := strings.Contains(strings.ToLower((*webpages)[i].Contents), strings.ToLower(term))
		if contains {
			documentCount++
		}
	}

	return documentCount
}
