package types

// Address represents a physical address
type Address struct {
	StreetAddress   string  `json:"streetAddress"`
	PostalCode      string  `json:"postalCode"`
	AddressLocality string  `json:"addressLocality"`
	Latitude        float64 `json:"latitude,omitempty"`
	Longitude       float64 `json:"longitude,omitempty"`
	Approximate     bool    `json:"approximate,omitempty"`
}
