package tfidf

import (
	"fmt"
	"math"
	"search-engine-service/utilities"
	"sort"
)

type WebpageRanking struct {
	Url    string
	Rating float64
}

// WHY DOES LOG RETURN NAN ON BOTH floats
func RankTFIDFRatings(IDF float64, webpages *[]utilities.WebpageTFIDF) *[]utilities.WebpageTFIDF {
	for i := range *webpages {
		tfidfRating := calculateTFIDF(IDF, (*webpages)[i])
		if math.IsNaN(tfidfRating) {
			tfidfRating = 0
		}
		(*webpages)[i].TFIDFRating = tfidfRating
	}

	// need to filter out 0 score
	webpagesSlice := (*webpages)[:]
	sort.Slice(webpagesSlice, func(i, j int) bool {
		return webpagesSlice[i].TFIDFRating > webpagesSlice[j].TFIDFRating
	})
	filteredWebpages := utilities.Filter(webpagesSlice)

	fmt.Printf("Filtered: %+v\n", filteredWebpages)

	return &filteredWebpages
}

func calculateTFIDF(IDF float64, webpage utilities.WebpageTFIDF) float64 {
	return webpage.TFScore * IDF
}
