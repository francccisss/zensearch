package bm25

import (
	"fmt"
	"search-engine-service/utilities"
	"testing"
)

func TestBM25Rating(t *testing.T) {

	const query = "our"
	IDF := CalculateIDF(query, &utilities.Webpages)
	err := TF(query, &utilities.Webpages)
	if err != nil {
		t.Fatalf("Error occured")
	}
	rankedWebpages := RankBM25Ratings(IDF, &utilities.Webpages)
	if len(*rankedWebpages) < 10 {
		t.Fatalf("Some webpages were not rated")
	}
}

func TestTokenizedQuery(t *testing.T) {
	query := "Welcome to our"
	tokenizedQuery := utilities.Tokenizer(query)

	for i := range tokenizedQuery {
		IDF := CalculateIDF(tokenizedQuery[i], &utilities.Webpages)
		_ = TF(tokenizedQuery[i], &utilities.Webpages)
		// Apply bm25 ratings for the current token.
		for i := range *&utilities.Webpages {
			BM25Rating := BM25(IDF, (*&utilities.Webpages)[i])
			(*&utilities.Webpages)[i].BM25Rating += BM25Rating
			// summing the existing rating on the current webpage with the newly calculated one
		}
	}

	for _, webpage := range *&utilities.Webpages {
		fmt.Printf("URL: %s\n", webpage.Url)
		fmt.Printf("TF Score: %f\n", webpage.TFScore)
		fmt.Printf("BM25 Score: %f\n", webpage.BM25Rating)
	}
	fmt.Printf("Search Query: %s\n", query)
}
