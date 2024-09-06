package utilities

func DocumentWordCount(contents string) int {
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
