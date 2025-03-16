package cache

import "time"

// Location represents a geocoded location
type Location struct {
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
	Failed bool    `json:"failed"`
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

// TournamentCache represents a cached tournament with all its data
type TournamentCache struct {
	ID                     int       `json:"id"`
	Name                   string    `json:"name"`
	Type                   string    `json:"type"`
	StartDate              string    `json:"startDate"`
	EndDate                string    `json:"endDate"`
	Address                Address   `json:"address"`
	Club                   Club      `json:"club"`
	Rules                  *Rules    `json:"rules,omitempty"`
	Endowment              int       `json:"endowment"`
	IsRulesPdfChecked      bool      `json:"isRulesPdfChecked,omitempty"`
	IsSiteExistenceChecked bool      `json:"isSiteExistenceChecked,omitempty"`
	SiteUrl                string    `json:"siteUrl,omitempty"`
	SignupUrl              string    `json:"signupUrl,omitempty"`
	Timestamp              time.Time `json:"timestamp"`
}

// Club represents a table tennis club
type Club struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Code       string `json:"code"`
	Department string `json:"department,omitempty"`
	Region     string `json:"region,omitempty"`
	Identifier string `json:"identifier"`
}

// Rules represents tournament rules
type Rules struct {
	AgeMin  int    `json:"ageMin,omitempty"`
	AgeMax  int    `json:"ageMax,omitempty"`
	Points  int    `json:"points,omitempty"`
	Ranking int    `json:"ranking,omitempty"`
	URL     string `json:"url,omitempty"`
}
