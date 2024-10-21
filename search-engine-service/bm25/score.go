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

func RankBM25Ratings(IDF float64, webpages *[]utilities.WebpageTFIDF) *[]utilities.WebpageTFIDF {
	for i := range *webpages {
		BM25Rating := BM25(IDF, (*webpages)[i])
		if math.IsNaN(BM25Rating) {
			BM25Rating = 0
		}
		(*webpages)[i].BM25Rating = BM25Rating
	}

	// need to filter out 0 score
	webpagesSlice := (*webpages)[:]
	sort.Slice(webpagesSlice, func(i, j int) bool {
		return webpagesSlice[i].BM25Rating > webpagesSlice[j].BM25Rating
	})
	filteredWebpages := utilities.Filter(webpagesSlice)

	fmt.Printf("Filtered: %+v\n", filteredWebpages)

	return &filteredWebpages
}

func BM25(IDF float64, webpage utilities.WebpageTFIDF) float64 {
	return webpage.TFScore * IDF
}
