package cache

import (
	"time"

	"tournois-tt/api/pkg/geocoding"
)

// Address is an alias for geocoding.Address for backward compatibility
type Address = geocoding.Address

// Location is an alias for geocoding.Location for backward compatibility
type Location = geocoding.Location

// TournamentCache represents a cached tournament with all its data
type TournamentCache struct {
	ID        int               `json:"id"`
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	StartDate string            `json:"startDate"`
	EndDate   string            `json:"endDate"`
	Address   geocoding.Address `json:"address"`
	Club      Club              `json:"club"`
	Rules     *Rules            `json:"rules,omitempty"`
	Endowment int               `json:"endowment"`
	Page      string            `json:"page,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
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

// GeocodeResult represents a cached geocoding result
type GeocodeResult struct {
	Address   geocoding.Address `json:"address"`
	Latitude  float64           `json:"latitude,omitempty"`
	Longitude float64           `json:"longitude,omitempty"`
	Failed    bool              `json:"failed"`
	Timestamp time.Time         `json:"timestamp"`
}

// GeocodeConfig allows configuring geocoding behavior
type GeocodeConfig struct {
	Enabled             bool
	MaxGeocodeAttempts  int
	SkipFailedAddresses bool
}

// DefaultGeocodeConfig provides default geocoding configuration
var DefaultGeocodeConfig = GeocodeConfig{
	Enabled:             false,
	MaxGeocodeAttempts:  3,
	SkipFailedAddresses: true,
}
