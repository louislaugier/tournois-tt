package geocoding

import (
	"strings"
)

func getCountry(country *string) string {
	if country == nil || *country == "" {
		return "France"
	}
	return *country
}

func hasValidAddress(addr Address) bool {
	street := strings.TrimSpace(addr.StreetAddress)
	desc := strings.TrimSpace(addr.DisambiguatingDescription)
	postal := strings.TrimSpace(addr.PostalCode)
	locality := strings.TrimSpace(addr.AddressLocality)

	// Must have either a street address or a gymnasium name
	hasLocation := street != "" || desc != ""
	if !hasLocation {
		return false
	}

	// Must have either postal code or city name
	hasArea := postal != "" || locality != ""
	if !hasArea {
		return false
	}

	return true
}
