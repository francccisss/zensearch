package bm25

import (
	"math"
	"search-engine-service/utilities"
	"strings"
)

func CalculateIDF(term string, webpages *[]utilities.WebpageTFIDF) float64 {

	numberOfDocumentsInCorpa := float64(len(*webpages))
	documentCountWithTerm := float64(termCountInCorpa(term, webpages))

	return math.Log(numberOfDocumentsInCorpa / documentCountWithTerm)
}

func termCountInCorpa(term string, webpages *[]utilities.WebpageTFIDF) int {
	documentCount := 0
	tokenizedTerm := utilities.Tokenizer(term)

	/*
	 Using composite terms we need to check if a document
	 contains all of these terms? or just some of the terms to
	 constitute a document as having the tokenized term?
	*/

	for i := range *webpages {
		containsAllTerms := true
		for j := range len(tokenizedTerm) {
			contains := strings.Contains(strings.ToLower((*webpages)[i].Contents), strings.ToLower(tokenizedTerm[j]))
			if !contains {
				containsAllTerms = false
				break
			}
		}
		if containsAllTerms {
			documentCount++
		}
	}

	return documentCount
}
