package bm25

import (
	"fmt"
	"math"
	"search-engine/internal/types"
	"search-engine/utilities"
	"sort"
	"strings"
	"sync"
)

const CHUNK_SIZE = 40

// takes exponential time
func CalculateBMRatings(query string, webpages *[]types.WebpageTFIDF) *[]types.WebpageTFIDF {
	fmt.Println("\n\nTEST: DP/TP/MS Pattern")
	tokenizedTerms := Tokenizer(query)

	// wpLen := len(*webpages)

	var wg sync.WaitGroup
	// get IDF and TF for each token
	IDFChan := make(chan float64, 10)
	// TODO do master slave, aggregate results back to go master routine

	go func() {
		var mwg sync.WaitGroup
		for i := range tokenizedTerms {
			mwg.Add(1)
			go func() {
				defer mwg.Done()
				// IDF is a constant throughout the current term
				IDF := CalculateIDF(tokenizedTerms[i], webpages)
				IDFChan <- IDF
			}()
		}
		mwg.Wait()
		close(IDFChan)
	}()

	wg.Add(1)
	go func() {

		var mwg sync.WaitGroup
		// Ranking webpages
		docLen := utilities.AvgDocLen(webpages)

		// creates task parallelism
		fmt.Println("TEST: creating task parallelism")
		for _, term := range tokenizedTerms {
			mwg.Add(1)

			go func() {
				// First calculate term frequency of each webpage for each token
				// TF(q1,webpages) -> TF(qT2,webpages)...

				// fmt.Printf("TEST: creating data chunks using data parallelism for term=%s\n", term)
				var swg sync.WaitGroup
				start := float64(0)
				end := float64(CHUNK_SIZE)
				wpLen := float64(len(*webpages))
				chunks := int(wpLen/CHUNK_SIZE + 1)
				// fmt.Printf("TEST chunk_distribution_length=%d\n", chunks)

				// creates data parallelism
				if len(tokenizedTerms) == 1 {
					fmt.Printf("TEST: aggregate chunks for term=%s\n", term)
					mwg.Done()
					return
				}
				for i := 0; i < chunks; i++ {
					// if equal to valid chunk size
					// fmt.Printf("TEST: start=%d, end=%d, diff=%d\n", int(start), int(end), int(math.Min(math.Abs(end-wpLen), CHUNK_SIZE)))
					swg.Add(1)
					go func() {
						defer swg.Done()
						_ = TF(term, docLen, webpages, int(start), int(end))
					}()
					start = float64(i) * CHUNK_SIZE
					end = start + math.Min(math.Abs(start+end-wpLen), CHUNK_SIZE)
				}
				// fmt.Println("TEST: waiting for chunks")
				swg.Wait()
				// fmt.Printf("TEST: aggregate chunks for term=%s\n", term)
				mwg.Done()
			}()
		}
		mwg.Wait()
		fmt.Println("TEST: Finished calculating webpages TF rating using task parallelism")
		wg.Done()
	}()

	fmt.Println("TEST: waiting for TF and IDF calculations")
	wg.Wait()
	// for each token calculate BM25Rating for each webpages
	// by summing the rating from the previous tokens
	fmt.Println("TEST: calculating bm25 rating")
	for IDF := range IDFChan {
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
	filteredWebpages := filter(webpagesSlice)

	// fmt.Printf("Filtered: %+v\n", filteredWebpages)

	return &filteredWebpages
}

func Tokenizer(query string) []string {
	tmpSlice := []string{}
	var charHolder = ""
	for i := 0; i < len(query); i++ {
		char := string(query[i])
		charHolder += char
		if char == " " {
			tmpSlice = append(tmpSlice, strings.Trim(charHolder, " "))
			charHolder = ""
		}
	}

	// add the remaining character after reaching null byte
	tmpSlice = append(tmpSlice, strings.Trim(charHolder, " "))
	fmt.Printf("Length of Token: %d\n", len(tmpSlice))
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
