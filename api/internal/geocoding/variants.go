package geocoding

import (
	"regexp"
	"strings"
)

func normalizeAddressForCache(address string) string {
	// Convert to lowercase
	address = strings.ToLower(address)

	// Remove extra spaces
	address = strings.TrimSpace(address)
	address = regexp.MustCompile(`\s+`).ReplaceAllString(address, " ")

	return address
}

func generateAddressVariants(addr Address) []string {
	variants := make(map[string]bool) // Use map to avoid duplicates

	// Helper function to add normalized variant
	addVariant := func(variant string) {
		if variant = normalizeAddressForCache(variant); variant != "" {
			variants[variant] = true
		}
	}

	// Base components
	streetName := strings.TrimSpace(addr.StreetAddress)
	city := strings.TrimSpace(addr.AddressLocality)
	gymnasium := strings.TrimSpace(addr.DisambiguatingDescription)
	postalCode := strings.TrimSpace(addr.PostalCode)

	// Extract street number if present
	var streetNum string
	streetNumRegex := regexp.MustCompile(`^\d+\s*`)
	if match := streetNumRegex.FindString(streetName); match != "" {
		streetNum = strings.TrimSpace(match)
		streetName = strings.TrimSpace(streetNumRegex.ReplaceAllString(streetName, ""))
	}

	// 1. Street name + city combinations
	if streetName != "" && city != "" {
		// With street number
		if streetNum != "" {
			addVariant(streetNum + " " + streetName + ", " + city)
			if postalCode != "" {
				addVariant(streetNum + " " + streetName + ", " + postalCode + " " + city)
			}
		}
		// Without street number
		addVariant(streetName + ", " + city)
		if postalCode != "" {
			addVariant(streetName + ", " + postalCode + " " + city)
		}
	}

	// 2. Street name + gymnasium combinations
	if streetName != "" && gymnasium != "" {
		// With street number
		if streetNum != "" {
			addVariant(streetNum + " " + streetName + " " + gymnasium)
			addVariant(gymnasium + ", " + streetNum + " " + streetName)
		}
		// Without street number
		addVariant(streetName + " " + gymnasium)
		addVariant(gymnasium + ", " + streetName)
	}

	// 3. Gymnasium + city combinations
	if gymnasium != "" && city != "" {
		addVariant(gymnasium + ", " + city)
		if postalCode != "" {
			addVariant(gymnasium + ", " + postalCode + " " + city)
		}
	}

	// Convert map keys to slice
	result := make([]string, 0, len(variants))
	for variant := range variants {
		result = append(result, variant)
	}
	return result
}
