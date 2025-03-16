package utilities

import "search-engine/internal/types"

func DocLength(contents string) int {
	totalWords := 0
	for i := 0; i < len(contents); i++ {
		char := string(contents[i])
		if char == " " {
			totalWords++
		}
	}
	// count last character
	totalWords++
	return totalWords
}

func AvgDocLen(webpages *[]types.WebpageTFIDF) float64 {

	totalTermCount := 0
	for i := range *webpages {
		docLength := DocLength((*webpages)[i].Contents)
		totalTermCount += docLength
	}
	return float64(totalTermCount) / float64(len(*webpages))
}
