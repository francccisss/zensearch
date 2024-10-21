package bm25

import (
	// "fmt"
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
		(*webpages)[i].BM25Rating = BM25Rating
	}

	// need to filter out 0 score
	webpagesSlice := (*webpages)[:]
	sort.Slice(webpagesSlice, func(i, j int) bool {
		return webpagesSlice[i].BM25Rating > webpagesSlice[j].BM25Rating
	})
	// filteredWebpages := utilities.Filter(webpagesSlice)

	// fmt.Printf("Filtered: %+v\n", filteredWebpages)

	return &webpagesSlice
}

func BM25(IDF float64, webpage utilities.WebpageTFIDF) float64 {
	return IDF * webpage.TFScore
}

// BM25 combines term frequency, inverse document frequency, and document length normalization to provide a balanced relevance score.
//
//     TF reflects how often the term appears in the document.
//     IDF reflects how important the term is based on its rarity.
//     Document length normalization adjusts for the term’s density and prevents long documents from dominating simply because of their length.
//
// Practical Impact on Ranking:
//
//     High TF and High IDF: If a rare term appears frequently in a document, that document is considered highly relevant.
//     High TF but Low IDF: A term that appears often but is common across documents will result in a lower score.
//     High IDF but Low TF: A rare term that appears only a few times can still result in a good score, especially in shorter documents.
//     Low TF and Low IDF: A common term that appears infrequently results in a low score.
