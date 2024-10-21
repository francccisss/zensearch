package bm25

import (
	"fmt"
	"search-engine-service/utilities"
	"testing"
)

func TestTF(t *testing.T) {

	const query = "our"
	IDF := CalculateIDF(query, &utilities.Webpages)
	err := TF(query, &utilities.Webpages)
	if err != nil {
		t.Fatalf("Error occured")
	}
	rankedWebpages := RankBM25Ratings(IDF, &utilities.Webpages)
	for _, webpage := range *rankedWebpages {
		fmt.Printf("URL: %s\n", webpage.Url)
		fmt.Printf("TF Score: %f\n", webpage.TFScore)
		fmt.Printf("BM25 Score: %f\n", webpage.BM25Rating)
	}
	fmt.Printf("Search Query: %s\n", query)
	fmt.Printf("IDF Score: %f\n", IDF)
	if len(*rankedWebpages) < 10 {
		t.Fatalf("Some webpages were not rated")

	}

}
