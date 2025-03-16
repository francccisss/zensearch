package bm25

import (
	"search-engine/internal/types"
	"search-engine/utilities"
	"strings"
)

const (
	k1 = 4  // controls the weight of term frequency, lower value saturates the term frequency quicker
	b  = .4 // controlling document normalization
)

// Updating TF ranking of each webpage
func TF(searchQuery string, webpages *[]types.WebpageTFIDF, AvgDocLen float64) error {

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
