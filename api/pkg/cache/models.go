package cache

import "time"

// Location represents a geocoded location
type Location struct {
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
	Failed bool    `json:"failed"`
}

// GeocodeResult represents a cached geocoding result
type GeocodeResult struct {
	Address   Address   `json:"address"`
	Latitude  float64   `json:"latitude,omitempty"`
	Longitude float64   `json:"longitude,omitempty"`
	Failed    bool      `json:"failed"`
	Timestamp time.Time `json:"timestamp"`
}

// Address represents a physical address
type Address struct {
	StreetAddress             string  `json:"streetAddress"`
	PostalCode                string  `json:"postalCode"`
	AddressLocality           string  `json:"addressLocality"`
	DisambiguatingDescription string  `json:"disambiguatingDescription,omitempty"`
	Latitude                  float64 `json:"latitude,omitempty"`
	Longitude                 float64 `json:"longitude,omitempty"`
	Failed                    bool    `json:"failed,omitempty"`
}

// IsValid checks if an address has enough information for geocoding
func (a Address) IsValid() bool {
	return a.PostalCode != "" && a.AddressLocality != ""
}
