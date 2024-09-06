package tfidf

import (
	"fmt"
	"search-engine-service/utilities"
	"strings"
)

// test this
func CalculateTF(searchQuery string, webpage *utilities.WebpageTFIDF) float32 {

	currentDocument := webpage.Contents
	totalWords := documentWordCount(currentDocument)
	termCount := strings.Count(strings.ToLower(currentDocument), strings.ToLower(searchQuery))
	webpage.TFScore = float32(termCount) / float32(totalWords)

	fmt.Printf("Term to look for: %s\n", searchQuery)
	fmt.Printf("Total terms from search query: %d\n", termCount)
	fmt.Printf("Total words from current document: %d\n", totalWords)
	fmt.Printf("TF Score: %f\n", webpage.TFScore)

	return webpage.TFScore
}

func documentWordCount(contents string) int {
	totalWords := 0
	currentWord := ""
	for char, i := range contents {
		curChar := string(char)
		nextChar := string(contents[i+1])
		if nextChar == " " || contents[i+1] == 0 {
			totalWords++
			continue
		}
		currentWord += curChar
	}
	return totalWords
}
