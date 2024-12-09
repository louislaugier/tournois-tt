package geocoding

import "time"

// Address represents a physical address with its components
type Address struct {
	StreetAddress             string
	PostalCode                string
	AddressLocality           string
	DisambiguatingDescription string // For gymnasium names or additional location info
	AddressCountry            *string
}

// Location represents a geographical point with its metadata
type Location struct {
	Lat         float64
	Lon         float64
	Failed      bool // Indicates if geocoding failed for this address
	Approximate bool // Indicates if location is approximate (city-level only)
	LastUpdated time.Time
}

// GeocodingCache represents the cache structure for geocoded addresses
type GeocodingCache struct {
	// Map of canonical address to location
	Locations map[string]Location
	// Map of non-canonical forms to canonical form
	Aliases  map[string]string
	LastSave time.Time
}

// nominatimResponse represents the response structure from Nominatim API
type nominatimResponse []struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
	// Nominatim specific fields for determining precision
	Type     string `json:"type"`
	Class    string `json:"class"`
	Category string `json:"category"`
}
