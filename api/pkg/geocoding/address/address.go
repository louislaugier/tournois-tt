package address

import (
	"regexp"
	"strings"
)

// Normalize standardizes an address string for consistent comparison and cache keys
func Normalize(address string) string {
	// Convert to lowercase and trim spaces
	address = strings.TrimSpace(strings.ToLower(address))

	// Replace multiple spaces with single space
	address = strings.Join(strings.Fields(address), " ")

	// Remove trailing commas and france suffix (for cache key consistency)
	address = strings.TrimRight(address, ", ")
	address = strings.TrimSuffix(address, ", france")

	return address
}

// GenerateVariants creates a minimal set of address variants in priority order:
// 1. Full address (everything)
// 2. Without venue name in parentheses
// 3. Without street number (if exists)
// 4. Just postal code + locality
func GenerateVariants(addr *AddressInput) []string {
	if addr == nil {
		return nil
	}

	// Clean inputs first
	streetAddress := strings.TrimSpace(addr.StreetAddress)
	postalCode := strings.TrimSpace(addr.PostalCode)
	locality := strings.TrimSpace(addr.AddressLocality)

	// Skip if we don't have both postal code and locality
	if postalCode == "" || locality == "" {
		return nil
	}

	// Base variant (postal code + locality) that will be used in all variants
	baseVariant := postalCode + " " + locality

	// Track unique normalized variants
	seen := make(map[string]bool)
	var variants []string

	// Helper to add variant if unique
	addVariant := func(variant string) {
		normalized := Normalize(variant)
		if !seen[normalized] {
			seen[normalized] = true
			variants = append(variants, normalized)
		}
	}

	// Always include base variant first
	addVariant(baseVariant + ", france")

	if streetAddress != "" {
		// Full address
		addVariant(streetAddress + ", " + baseVariant + ", france")

		// Try without venue name if exists
		re := regexp.MustCompile(`\s*\([^)]+\)`)
		streetWithoutVenue := strings.TrimSpace(re.ReplaceAllString(streetAddress, ""))
		if streetWithoutVenue != streetAddress {
			addVariant(streetWithoutVenue + ", " + baseVariant + ", france")
		}

		// Try without street number if it exists
		re = regexp.MustCompile(`^[0-9]+[A-Za-z]?(?:[-/][0-9]+)?[\s,]+(.+)$`)
		if matches := re.FindStringSubmatch(streetWithoutVenue); matches != nil {
			streetWithoutNumber := strings.TrimSpace(matches[1])
			addVariant(streetWithoutNumber + ", " + baseVariant + ", france")
		}
	}

	return variants
}

// IsValid checks if an address has enough information for geocoding
func IsValid(addr AddressInput) bool {
	return strings.TrimSpace(addr.PostalCode) != "" && strings.TrimSpace(addr.AddressLocality) != ""
}
