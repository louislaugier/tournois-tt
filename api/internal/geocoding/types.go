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

// Coordinates represents a geographical point with latitude and longitude
type Coordinates struct {
	Lat    float64
	Lon    float64
	Failed bool // Indicates if geocoding failed for this address
}

// GeocodingCache represents the cache structure for geocoded addresses
type GeocodingCache struct {
	Addresses map[string]Coordinates
	LastSave  time.Time
}

// nominatimResponse represents the response structure from Nominatim API
type nominatimResponse []struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}
