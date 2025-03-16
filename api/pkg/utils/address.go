package utils

import (
	"regexp"
)

// ExtractPostalCode extracts a French postal code from a location string
func ExtractPostalCode(location string) string {
	// French postal codes are 5 digits
	// Try to find a 5-digit sequence in the location string

	// Regular expression to match French postal codes (5 digits)
	re := regexp.MustCompile(`\b\d{5}\b`)
	matches := re.FindAllString(location, -1)

	if len(matches) > 0 {
		return matches[0]
	}

	return ""
}
