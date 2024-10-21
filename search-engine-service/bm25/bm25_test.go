package bm25

import (
	"fmt"
	"search-engine-service/utilities"
	"sort"
	"testing"
)

func TestBM25Rating(t *testing.T) {

	const query = "community"
	IDF := CalculateIDF(query, &utilities.Webpages)
	err := TF(query, &utilities.Webpages)
	if err != nil {
		t.Fatalf("Error occured")
	}
	rankedWebpages := RankBM25Ratings(IDF, &utilities.Webpages)

	for _, webpage := range *&utilities.Webpages {
		fmt.Printf("URL: %s\n", webpage.Url)
		fmt.Printf("TF Score: %f\n", webpage.TFScore)
		fmt.Printf("BM25 Score: %f\n\n", webpage.BM25Rating)
	}
	fmt.Printf("Search Query for single token: %s\n\n", query)
	if len(*rankedWebpages) < 10 {
		t.Fatalf("Some webpages were not rated")
	}
}

func TestTokenizedQuery(t *testing.T) {
	query := "Learn more"
	tokenizedQuery := utilities.Tokenizer(query)

	// get IDF and TF for each token
	for i := range tokenizedQuery {
		IDF := CalculateIDF(tokenizedQuery[i], &utilities.Webpages)
		_ = TF(tokenizedQuery[i], &utilities.Webpages)
		// for each token calculate BM25Rating for each webpages
		// by summing the rating from the previous tokens
		for i := range *&utilities.Webpages {
			BM25Rating := BM25(IDF, (*&utilities.Webpages)[i])
			(*&utilities.Webpages)[i].BM25Rating += BM25Rating
		}
	}

	webpagesSlice := (*&utilities.Webpages)[:]
	sort.Slice(webpagesSlice, func(i, j int) bool {
		return webpagesSlice[i].BM25Rating > webpagesSlice[j].BM25Rating
	})
	for _, webpage := range webpagesSlice[:5] {
		fmt.Printf("URL: %s\n", webpage.Url)
		fmt.Printf("TF Score: %f\n", webpage.TFScore)
		fmt.Printf("BM25 Score: %f\n\n", webpage.BM25Rating)
	}
	fmt.Printf("Search Query for composite query: %s\n\n", query)
}
