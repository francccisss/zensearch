package bm25

import (
	"fmt"
	"math"
	"search-engine/internal/types"
	"search-engine/utilities"
	"sort"
	"sync"
)

// CHUNK_SIZE is the total amount of webpages within each chunks to be proccessed in parallel
const CHUNK_SIZE = 40
const MAX_TOKEN_COUNT = 10

// TODO CONCURRENCY ISSUE NOT FROM TEST AND MAIN
// SOMETIMES TEXT IS FULLY RENDERED AND RANKED SOMETIMES IT ISNT
// takes exponential time
func CalculateBMRatings(query string, webpages *[]types.WebpageTFIDF) *[]types.WebpageTFIDF {
	// return immediately if database is currently empty
	if len(*webpages) == 0 {
		return nil
	}
	tokenizedTerms := Tokenizer(query)

	// get IDF and TF for each token
	// token count is unknown so keeping a space of MAX_TOKEN_COUNT
	// for reasonable user search tokens.
	IDFChann := make(chan float64, MAX_TOKEN_COUNT)

	var iwg sync.WaitGroup
	for i := range tokenizedTerms {
		iwg.Go(func() {
			// IDF is a constant throughout the current term
			IDF := CalculateIDF(tokenizedTerms[i], webpages)
			IDFChann <- IDF
		})
	}

	totalWebpages := float64(len(*webpages))
	totalChunks := math.Min(totalWebpages/CHUNK_SIZE, CHUNK_SIZE) + 1
	fmt.Printf("[total_webpages=%d, default_chunk_size=%d, chunk_distribution_length=%d]\n", len(*webpages), CHUNK_SIZE, int(totalChunks))
	var twg sync.WaitGroup
	// Ranking webpages
	docLen := utilities.AvgDocLen(webpages)

	var mux sync.Mutex
	// creates task parallelism
	for _, term := range tokenizedTerms {

		twg.Go(func() {
			// First calculate term frequency of each webpage for each token
			// TF(q1,webpages) -> TF(q2,webpages)...
			var swg sync.WaitGroup
			start := float64(0)
			end := math.Min(totalWebpages, CHUNK_SIZE)

			for range int(totalChunks) {
				swg.Go(func() {
					_ = TF(term, docLen, webpages, int(start), int(end)) // adds old rating if exists
					mux.Lock()
					start = end + 1 // always 0
					end = start + math.Min(math.Abs(start-totalWebpages), CHUNK_SIZE)
					mux.Unlock()
				})
				// Need to make sure that each go routine is synchornized to process their own distributed chunks.
			}
			swg.Wait()
		})
	}
	// release of both TF and IDF processing
	twg.Wait()
	iwg.Wait()
	close(IDFChann)
	// for each IDF of each token, calculate BM25Rating for each webpages
	// by summing the rating from the previous rated token
	for IDF := range IDFChann {
		for j := range *webpages {
			bm25rating := BM25(IDF, (*webpages)[j].TfRating)
			(*webpages)[j].TokenRating.Bm25rating += bm25rating
		}
	}
	return webpages
}

func RankBM25Ratings(webpages *[]types.WebpageTFIDF) *[]types.WebpageTFIDF {
	webpagesSlice := (*webpages)[:]
	// TODO replace sort.Slice with slices.SortFunc
	sort.Slice(webpagesSlice, func(i, j int) bool {
		return webpagesSlice[i].TokenRating.Bm25rating > webpagesSlice[j].TokenRating.Bm25rating
	})
	// filteredWebpages := filter(webpagesSlice)

	return &webpagesSlice
}

// TODO SKIP SPACES
func Tokenizer(query string) []string {
	tmpSlice := []string{}
	cur := 0
	read := 0

	for range query {
		if query[read] == ' ' {
			stringHolder := query[cur:read]
			tmpSlice = append(tmpSlice, stringHolder)
			cur = read
		}
		read++
	}

	stringHolder := query[cur:read]
	tmpSlice = append(tmpSlice, stringHolder)
	fmt.Printf("TOKENIZER SLICE=%+v\n", tmpSlice)
	return tmpSlice
}

func filter(webpages []types.WebpageTFIDF) []types.WebpageTFIDF {
	tmp := make([]types.WebpageTFIDF, 0)
	for _, webpage := range webpages {
		if webpage.TokenRating.Bm25rating == 0 {
			break
		}
		tmp = append(tmp, webpage)
	}
	return tmp
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
