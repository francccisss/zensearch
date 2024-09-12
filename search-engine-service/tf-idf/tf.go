package tfidf

import (
	"search-engine-service/utilities"
	"strings"
)

func CalculateTF(searchQuery string, webpages *[]utilities.WebpageTFIDF) *[]utilities.WebpageTFIDF {

	for i := range *webpages {

		currentDocument := (*webpages)[i].Contents
		totalWords := utilities.DocumentWordCount(currentDocument)
		termCount := strings.Count(strings.ToLower(currentDocument), strings.ToLower(searchQuery))
		(*webpages)[i].TFScore = float64(termCount) / float64(totalWords)
	}
	return webpages
}
