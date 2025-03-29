package utils

import (
	"strings"
)

// ExtractSignificantWords extracts significant words from text
// by removing common stop words and short words
func ExtractSignificantWords(text string) []string {
	// Clean up the text
	text = strings.ToLower(text)
	words := strings.Fields(text)

	// Filter out common words and keep significant ones
	var significantWords []string
	stopWords := map[string]bool{
		"de": true, "du": true, "la": true, "le": true, "les": true,
		"des": true, "un": true, "une": true, "et": true, "à": true,
		"au": true, "aux": true, "en": true, "dans": true, "par": true,
		"pour": true, "sur": true, "avec": true, "sans": true,
	}

	minWordLength := 3
	for _, word := range words {
		if len(word) >= minWordLength && !stopWords[word] {
			significantWords = append(significantWords, word)
		}
	}

	return significantWords
}
