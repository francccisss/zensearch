package utilities

// Used for filtering out unrelated data from the query
// First the Slice will be passed into a sorting function,
// and then will be passed into this function for filtering,
// this function will go through each webpage, and ONCE
// it points to a webpage with a rating of 0, it will filter the rest out
// and return the filtered webpages.

func Filter(webpages []WebpageTFIDF) []WebpageTFIDF {
	tmp := make([]WebpageTFIDF, 0)
	for _, webpage := range webpages {
		if webpage.TokenRating.Bm25rating == 0 {
			break
		}
		tmp = append(tmp, webpage)
	}
	return tmp
}
