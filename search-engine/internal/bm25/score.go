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

// TODO CONCURRENCY ISSUE NOT FROM TEST AND MAIN
// SOMETIMES TEXT IS FULLY RENDERED AND RANKED SOMETIMES IT ISNT
// takes exponential time
func CalculateBMRatings(query string, webpages *[]types.WebpageTFIDF) *[]types.WebpageTFIDF {
	// return immediately if database is currently empty
	if len(*webpages) == 0 {
		return webpages
	}
	var mux sync.Mutex
	tokenizedTerms := Tokenizer(query)

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

		// TODO FIX THE MATH HERE
		// divide to valid chunks valid chunks == chunk_size = total webpages in chunk
		// if total webapages is odd add another chunk
		// calculate end of each chunks
		// total wbpg = 15 chunk_size = 10
		// 1.5 = 2 chunks -> chunks[0] = total wbpg = 10, chunks[1] = total wbpg = 5
		// for each chunk process from 0 to size of wbpg in chunk
		// calculate end = start + math.Min(remaning webpages,chunk_size)

		var mwg sync.WaitGroup
		// Ranking webpages
		docLen := utilities.AvgDocLen(webpages)

		// creates task parallelism
		for _, term := range tokenizedTerms {
			mwg.Add(1)

			go func() {
				// First calculate term frequency of each webpage for each token
				// TF(q1,webpages) -> TF(q2,webpages)...

				// You might be asking, why im using floats here? math.Min returns a 64bit float :D

				var swg sync.WaitGroup
				start := float64(0)
				totalWebpages := float64(len(*webpages))
				// Since math.Round(totalWebpages/CHUNK_SIZE) could return a float < 0
				// the totalwebpages which is < CHUNK_SIZE would not be processed if math.Min()
				// was used, so instead it will always be assumed that there will always be 1 chunk
				// to process every webpage
				totalChunks := math.Min(totalWebpages/CHUNK_SIZE, CHUNK_SIZE) + 1
				// length of the total webpages in the database
				fmt.Printf("TEST chunk_distribution_length=%d\n", int(totalChunks))
				// INIT END INDEX
				end := math.Min(totalWebpages, CHUNK_SIZE)

				for range int(totalChunks) {
					swg.Add(1)
					go func() {
						defer swg.Done()

						// TF(q1,webpages) + TF(q2,webpages)...
						_ = TF(term, docLen, webpages, int(start), int(end)) // adds old rating if exists
						mux.Unlock()

					}()
					// fmt.Println(start - end)
					mux.Lock()
					start = end + 1 // always 0
					end = start + math.Min(math.Abs(start-totalWebpages), CHUNK_SIZE)
				}
				swg.Wait()
				mwg.Done()
			}()
		}
		mwg.Wait()
		wg.Done()
	}()
	wg.Wait()
	// for each IDF of each token, calculate BM25Rating for each webpages
	// by summing the rating from the previous rated token
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
