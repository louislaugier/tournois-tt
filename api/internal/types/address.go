package types

// Address represents a physical address with optional geocoding information
type Address struct {
	// Basic address fields
	StreetAddress             string `json:"streetAddress"`
	PostalCode                string `json:"postalCode"`
	AddressLocality           string `json:"addressLocality"`
	AddressCountry            string `json:"addressCountry"`
	DisambiguatingDescription string `json:"disambiguatingDescription"`

	// Geocoding fields
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Approximate bool    `json:"approximate"`
}

// AddressInput represents the input format for address operations
type AddressInput struct {
	StreetAddress   string `json:"streetAddress"`
	PostalCode      string `json:"postalCode"`
	AddressLocality string `json:"addressLocality"`
	AddressCountry  string `json:"addressCountry"`
}
