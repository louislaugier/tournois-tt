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

// cleanStreetAddress removes common prefixes and standardizes the street address
func cleanStreetAddress(addr string) string {
	// Common venue prefixes to remove
	prefixes := []string{
		"gymnase", "salle", "complexe", "espace", "cosec",
		"stade", "centre", "palais", "maison", "parc",
	}

	addr = strings.ToLower(addr)
	for _, prefix := range prefixes {
		if strings.HasPrefix(addr, prefix+" ") {
			addr = strings.TrimPrefix(addr, prefix+" ")
		}
	}

	return strings.TrimSpace(addr)
}

// GenerateVariants creates a comprehensive set of address variants in priority order
func GenerateVariants(addr *AddressInput) []string {
	if addr == nil {
		return nil
	}

	// Clean inputs first
	streetAddress := strings.TrimSpace(addr.StreetAddress)
	postalCode := strings.TrimSpace(addr.PostalCode)
	locality := strings.TrimSpace(addr.AddressLocality)
	disambiguatingDesc := strings.TrimSpace(addr.DisambiguatingDescription)

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

	// Build variants in order of reliability (most reliable first)

	// 1. Street address with postal code and locality (without venue name)
	if streetAddress != "" {
		addVariant(streetAddress + ", " + baseVariant + ", france")

		// Remove venue names in parentheses if present
		reVenue := regexp.MustCompile(`\s*\([^)]+\)`)
		streetWithoutVenue := strings.TrimSpace(reVenue.ReplaceAllString(streetAddress, ""))
		if streetWithoutVenue != streetAddress {
			addVariant(streetWithoutVenue + ", " + baseVariant + ", france")
		}

		// Remove street numbers if present
		reNumber := regexp.MustCompile(`^\d+\s*`)
		streetWithoutNumber := strings.TrimSpace(reNumber.ReplaceAllString(streetAddress, ""))
		if streetWithoutNumber != streetAddress {
			addVariant(streetWithoutNumber + ", " + baseVariant + ", france")
		}

		// Try with just the street name (no number, no venue)
		streetNameOnly := strings.TrimSpace(reNumber.ReplaceAllString(streetWithoutVenue, ""))
		if streetNameOnly != "" && streetNameOnly != streetAddress && streetNameOnly != streetWithoutNumber {
			addVariant(streetNameOnly + ", " + baseVariant + ", france")
		}
	}

	// 2. Full address with venue name (if available)
	if disambiguatingDesc != "" && streetAddress != "" {
		addVariant(disambiguatingDesc + " " + streetAddress + ", " + baseVariant + ", france")
	}

	// 3. Venue name with postal code and locality
	if disambiguatingDesc != "" {
		addVariant(disambiguatingDesc + ", " + baseVariant + ", france")
	}

	// 4. Base variants (postal code and locality only)
	addVariant(baseVariant + ", france")
	addVariant(locality + ", " + postalCode + ", france")

	return variants
}

// IsValid checks if an address has enough information for geocoding
func IsValid(addr AddressInput) bool {
	return strings.TrimSpace(addr.PostalCode) != "" && strings.TrimSpace(addr.AddressLocality) != ""
}
