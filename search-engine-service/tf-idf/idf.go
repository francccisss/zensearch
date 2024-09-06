package tfidf

import (
	"fmt"
	"math"
	"search-engine-service/utilities"
	"strings"
)

func CalculateIDF(searchQuery string, webpages []utilities.WebpageTFIDF) float64 {

	totalWords := float64(wordCountInCorpa(webpages))
	totalTermsInDocument := float64(termCountInDocuments(searchQuery, webpages))

	fmt.Printf("Total words in corpa: %f\n", totalWords)
	fmt.Printf("Total terms in corpa: %f\n", totalWords)
	fmt.Printf("Calculated IDF: %f\n", math.Log2(totalTermsInDocument/totalWords))

	return math.Log2(totalTermsInDocument / totalWords)
}

func termCountInDocuments(term string, corpa []utilities.WebpageTFIDF) int {
	termCount := 0
	for _, webpage := range corpa {
		termCount += strings.Count(webpage.Contents, term)
	}
	return termCount
}

func wordCountInCorpa(corpa []utilities.WebpageTFIDF) int {
	wordCount := 0
	for _, webpage := range corpa {
		webpageWords := utilities.DocumentWordCount(webpage.Contents)
		wordCount += webpageWords
	}
	return wordCount

}
