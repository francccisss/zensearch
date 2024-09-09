package tfidf

import (
	"math"
	"search-engine-service/utilities"
	"strings"
)

func CalculateIDF(term string, webpages *[]utilities.WebpageTFIDF) float64 {

	// totalWords := float64(wordCountInCorpa(webpages))
	numberOfDocumentsInCorpa := float64(len(*webpages))
	totalTermsInDocument := float64(termCountInCorpa(term, webpages))
	return math.Log2(numberOfDocumentsInCorpa / totalTermsInDocument)
}

func termCountInCorpa(term string, corpa *[]utilities.WebpageTFIDF) int {
	documentCount := 0
	// corpa is just all of the webpages from different websites
	for i := range *corpa {
		// for every document in the corpa, count the documents
		// containing the term within the document.
		contains := strings.Contains((*corpa)[i].Contents, term)
		if contains {
			documentCount++
		}
	}
	return documentCount
}

func wordCountInCorpa(corpa *[]utilities.WebpageTFIDF) int {
	wordCount := 0
	for i := range *corpa {
		webpageWords := utilities.DocumentWordCount((*corpa)[i].Contents)
		wordCount += webpageWords
	}
	return wordCount

}
