package utilities

import (
	"strings"
)

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
	return tmpSlice
}
