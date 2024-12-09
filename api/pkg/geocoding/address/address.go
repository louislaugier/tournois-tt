package address

import (
	"fmt"
	"strings"
)

// Normalize standardizes an address string for consistent comparison
func Normalize(address string) string {
	// Convert to lowercase
	address = strings.ToLower(address)

	// Remove extra spaces, including any within words
	address = strings.Join(strings.Fields(strings.ReplaceAll(address, "-", " - ")), " ")
	address = strings.ReplaceAll(address, " - ", "-") // restore hyphens

	// Remove trailing commas
	address = strings.TrimRight(address, ",")

	// Remove ", france" suffix
	address = strings.TrimSuffix(address, ", france")

	// Normalize postal code format (ensure space after postal code)
	if len(address) > 5 && strings.ContainsAny(address[:5], "0123456789") {
		postalCode := address[:5]
		rest := strings.TrimSpace(address[5:])
		address = postalCode + " " + rest
	}

	return address
}

// GenerateVariants creates all possible variants of an address
func GenerateVariants(addr *AddressInput) []string {
	if addr == nil {
		return nil
	}

	var variants []string
	streetAddress := strings.TrimSpace(addr.StreetAddress)
	postalCode := strings.TrimSpace(addr.PostalCode)
	locality := strings.TrimSpace(addr.AddressLocality)
	country := "france"

	// Helper function to add variants with and without country
	addVariant := func(base string) {
		if base != "" {
			variants = append(variants, Normalize(base))
			variants = append(variants, Normalize(base+", "+country))
		}
	}

	// 1. Try with absolutely everything first
	if streetAddress != "" && postalCode != "" && locality != "" {
		addVariant(fmt.Sprintf("%s, %s %s", streetAddress, postalCode, locality))
		addVariant(fmt.Sprintf("%s %s %s", streetAddress, postalCode, locality))
	}

	// 2. Try with street address and locality
	if streetAddress != "" && locality != "" {
		addVariant(fmt.Sprintf("%s, %s", streetAddress, locality))
		addVariant(fmt.Sprintf("%s %s", streetAddress, locality))
	}

	// 3. If it's a gymnasium/sports facility, try variations
	if strings.Contains(strings.ToLower(streetAddress), "gymnase") ||
		strings.Contains(strings.ToLower(streetAddress), "salle") ||
		strings.Contains(strings.ToLower(streetAddress), "complexe") {

		// Extract street number if present
		var number string
		var street string
		parts := strings.Fields(streetAddress)
		if len(parts) > 0 && strings.ContainsAny(parts[0], "0123456789") {
			number = parts[0]
			street = strings.Join(parts[1:], " ")
		} else {
			street = streetAddress
		}

		// Try with number if available
		if number != "" {
			addVariant(fmt.Sprintf("%s %s, %s", number, street, locality))
			addVariant(fmt.Sprintf("%s %s %s", number, street, locality))
		}

		// Try without number
		addVariant(fmt.Sprintf("%s, %s", street, locality))
		addVariant(fmt.Sprintf("%s %s", street, locality))

		// Try just the gymnasium name with locality
		for _, prefix := range GetFacilityPrefixes() {
			if strings.HasPrefix(strings.ToLower(street), strings.ToLower(prefix)) {
				gymName := strings.TrimSpace(street[len(prefix):])
				if gymName != "" {
					addVariant(fmt.Sprintf("%s, %s", gymName, locality))
					addVariant(fmt.Sprintf("%s %s", gymName, locality))
				}
			}
		}
	}

	// 4. Try postal code with locality
	if postalCode != "" && locality != "" {
		addVariant(fmt.Sprintf("%s %s", postalCode, locality))
	}

	// 5. Try locality only as last resort (mark as approximate)
	if locality != "" {
		addVariant(locality)
	}

	// Remove duplicates while preserving order
	seen := make(map[string]bool)
	var uniqueVariants []string
	for _, v := range variants {
		if !seen[v] {
			seen[v] = true
			uniqueVariants = append(uniqueVariants, v)
		}
	}

	return uniqueVariants
}

// IsValid checks if an address has enough information for geocoding
func IsValid(addr AddressInput) bool {
	return (addr.StreetAddress != "" && addr.PostalCode != "" && addr.AddressLocality != "") ||
		addr.AddressLocality != "" ||
		addr.PostalCode != ""
}
