package tfidf

import (
	"math"
	"search-engine-service/utilities"
	"strings"
)

func CalculateIDF(searchQuery string, webpages *[]utilities.WebpageTFIDF) float64 {

	totalWords := float64(wordCountInCorpa(webpages))
	totalTermsInDocument := float64(termCountInCorpa(searchQuery, webpages))

	return math.Log2(totalWords / totalTermsInDocument)
}

func termCountInCorpa(term string, corpa *[]utilities.WebpageTFIDF) int {
	termCount := 0
	for i := range *corpa {
		termCount += strings.Count((*corpa)[i].Contents, term)
	}
	return termCount
}

func wordCountInCorpa(corpa *[]utilities.WebpageTFIDF) int {
	wordCount := 0
	for i := range *corpa {
		webpageWords := utilities.DocumentWordCount((*corpa)[i].Contents)
		wordCount += webpageWords
	}
	return wordCount

}
