package bm25

import (
	"fmt"
	"math"
	"search-engine-service/utilities"
	"strings"
)

func CalculateIDF(term string, webpages *[]utilities.WebpageTFIDF) float64 {

	// totalWords := float64(wordCountInCorpa(webpages))
	numberOfDocumentsInCorpa := float64(len(*webpages))
	documentCountWithTerm := float64(termCountInCorpa(term, webpages))

	fmt.Printf("DocCount: %f\n", numberOfDocumentsInCorpa)
	fmt.Printf("DocsWithTerm: %f\n", documentCountWithTerm)
	return math.Log(numberOfDocumentsInCorpa / documentCountWithTerm)
}

func termCountInCorpa(term string, webpages *[]utilities.WebpageTFIDF) int {
	documentCount := 0
	// corpa is just all of the webpages from different websites
	for i := range *webpages {
		// for every document in the corpa, count the documents
		// containing the term within the document.
		contains := strings.Contains(strings.ToLower((*webpages)[i].Contents), strings.ToLower(term))
		if contains {
			documentCount++
		}
	}
	return documentCount
}
