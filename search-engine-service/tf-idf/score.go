package tfidf

import (
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

	webpagesSlice := (*webpages)[:]
	sort.Slice(webpagesSlice, func(i, j int) bool {
		return webpagesSlice[i].TFIDFRating > webpagesSlice[j].TFIDFRating
	})
	return &webpagesSlice
}

func calculateTFIDF(IDF float64, webpage utilities.WebpageTFIDF) float64 {
	return webpage.TFScore * IDF
}
