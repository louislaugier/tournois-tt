package address

// AddressInput represents the input format for address operations
type AddressInput struct {
	StreetAddress             string `json:"streetAddress"`
	PostalCode                string `json:"postalCode"`
	AddressLocality           string `json:"addressLocality"`
	AddressCountry            string `json:"addressCountry"`
	DisambiguatingDescription string `json:"disambiguatingDescription"`
}
