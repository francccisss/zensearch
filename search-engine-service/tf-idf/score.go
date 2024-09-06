package tfidf

import "search-engine-service/utilities"

func RankTFIDFRatings(IDF float64, webpages *[]utilities.WebpageTFIDF) *[]utilities.WebpageTFIDF {
	for _, webpage := range *webpages {
		webpage.TFIDFRating = calculateTFIDF(IDF, webpage)
	}
	return webpages
}

func calculateTFIDF(IDF float64, webpage utilities.WebpageTFIDF) float64 {
	return webpage.TFScore * IDF
}
