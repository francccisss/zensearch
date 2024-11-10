package bm25

import (
	"fmt"
	"search-engine/utilities"
	"sort"
)

type WebpageTFIDF struct {
	Contents string
	Title    string
	Url      string
	TokenRating
}
type TokenRating struct {
	Bm25rating float64
	TfRating   float64
	IdfRating  float64
}

type WebpageRanking struct {
	Url    string
	Rating float64
}

func CalculateBMRatings(query string, webpages *[]WebpageTFIDF, AvgDocLen float64) *[]WebpageTFIDF {
	tokenizedQuery := utilities.Tokenizer(query)
	fmt.Println(tokenizedQuery)

	// get IDF and TF for each token
	for i := range tokenizedQuery {
		// IDF is a constant throughout the current term
		IDF := CalculateIDF(tokenizedQuery[i], webpages)

		// First calculate term frequency of each webpage for each token
		// TF(q1,webpages) -> TF(qT2,webpages)...
		_ = TF(tokenizedQuery[i], webpages, AvgDocLen)

		// for each token calculate BM25Rating for each webpages
		// by summing the rating from the previous tokens
		for j := range *webpages {
			bm25rating := BM25(IDF, (*webpages)[j].TfRating)
			(*webpages)[j].TokenRating.Bm25rating += bm25rating
		}
	}
	return webpages
}

func RankBM25Ratings(webpages *[]WebpageTFIDF) *[]WebpageTFIDF {
	webpagesSlice := (*webpages)[:]

	// TODO replace sort.Slice with slices.SortFunc
	sort.Slice(webpagesSlice, func(i, j int) bool {
		return webpagesSlice[i].TokenRating.Bm25rating > webpagesSlice[j].TokenRating.Bm25rating
	})
	filteredWebpages := utilities.Filter(webpagesSlice)

	// fmt.Printf("Filtered: %+v\n", filteredWebpages)

	return &filteredWebpages
}

func BM25(IDF, TF float64) float64 {
	return IDF * TF
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
