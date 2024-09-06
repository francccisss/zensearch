package tfidf

import (
	"fmt"
	"search-engine-service/utilities"
	"strings"
)

func CalculateTF(searchQuery string, webpages *[]utilities.WebpageTFIDF) *[]utilities.WebpageTFIDF {

	for _, webpage := range *webpages {

		currentDocument := webpage.Contents
		totalWords := utilities.DocumentWordCount(currentDocument)
		termCount := strings.Count(strings.ToLower(currentDocument), strings.ToLower(searchQuery))
		webpage.TFScore = float64(termCount) / float64(totalWords)

		fmt.Printf("Term to look for: %s\n", searchQuery)
		fmt.Printf("Current Document URL: %s\n", webpage.Webpage_url)
		fmt.Printf("Total terms from search query: %d\n", termCount)
		fmt.Printf("Total words from current document: %d\n", totalWords)
		fmt.Printf("TF Score: %f\n", webpage.TFScore)
	}

	return webpages
}
