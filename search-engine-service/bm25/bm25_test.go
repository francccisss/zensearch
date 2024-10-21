package bm25

import (
	"fmt"
	"search-engine-service/utilities"
	"testing"
)

const query = "home"

func TestTokenizedQuery(t *testing.T) {
	tokenizedQuery := utilities.Tokenizer(query)

	// get IDF and TF for each token
	for i := range tokenizedQuery {
		// IDF is a constant throughout the current term
		IDF := CalculateIDF(tokenizedQuery[i], &utilities.Webpages)

		// First calculate term frequency of each webpage for each token
		// TF(q1,webpages) -> TF(qT2,webpages)...
		_ = TF(tokenizedQuery[i], &utilities.Webpages)

		// for each token calculate BM25Rating for each webpages
		// by summing the rating from the previous tokens
		for i := range *&utilities.Webpages {
			bm25rating := BM25(IDF, (*&utilities.Webpages)[i].TfRating)
			(*&utilities.Webpages)[i].TokenRating.Bm25rating += bm25rating
		}
	}
}

func TestRankingWebpages(t *testing.T) {

	calculatedWebpages := CalculateBMRatings(query, &utilities.Webpages)
	rankedWebpages := RankBM25Ratings(calculatedWebpages)

	for _, webpage := range *rankedWebpages {
		fmt.Printf("URL: %s\n", webpage.Url)
		fmt.Printf("TF Score: %f\n", webpage.TokenRating.TfRating)
		fmt.Printf("BM25 Score: %f\n\n", webpage.TokenRating.Bm25rating)
	}
	fmt.Printf("Search Query for composite query: %s\n\n", query)

}
