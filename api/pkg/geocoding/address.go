// Package geocoding provides functionality for geocoding addresses
package geocoding

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

// Location represents a geocoded location
type Location struct {
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
	Failed bool    `json:"failed"`
}

// IsAddressValid checks if an address has enough data to be geocoded
func IsAddressValid(address Address) bool {
	if address.PostalCode == "" {
		return false
	}
	if address.AddressLocality == "" {
		return false
	}
	return true
}

// Note: The ConstructFullAddress function has been moved to geocoding.go to avoid duplication
