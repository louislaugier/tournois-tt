package utils

import (
	"regexp"
	"strings"
)

// ExtractPostalCode extracts a French postal code from a location string
func ExtractPostalCode(location string) string {
	if location == "" {
		return ""
	}

	// First, normalize the location string
	location = strings.TrimSpace(location)

	// Check for pattern "City (XXXXX)" which is common in HelloAsso
	reParentheses := regexp.MustCompile(`\((\d{5})\)`)
	parenthesesMatches := reParentheses.FindStringSubmatch(location)
	if len(parenthesesMatches) > 1 {
		return parenthesesMatches[1]
	}

	// French postal codes are 5 digits
	// Regular expression to match French postal codes (5 digits)
	re := regexp.MustCompile(`\b\d{5}\b`)
	matches := re.FindAllString(location, -1)

	if len(matches) > 0 {
		return matches[0]
	}

	// If we can't find a full postal code, try to extract just the first two
	// digits (department number) for a partial match
	rePartial := regexp.MustCompile(`\b(\d{2})[^\d]`)
	partialMatches := rePartial.FindStringSubmatch(location)

	if len(partialMatches) > 1 {
		// Return with wildcards for the remaining 3 digits
		return partialMatches[1] + "000"
	}

	return ""
}
