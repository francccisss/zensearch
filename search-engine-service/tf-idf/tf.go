package tfidf

import (
	"fmt"
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
	fmt.Printf("Bruh: %+v\n", (*webpages)[0])
	return webpages
}
