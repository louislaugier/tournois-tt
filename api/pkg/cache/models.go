package cache

import (
	"time"

	"tournois-tt/api/pkg/models"
)

// Address is an alias for models.Address for backward compatibility
type Address = models.Address

// Location is an alias for models.Location for backward compatibility
type Location = models.Location

// TournamentCache represents a cached tournament with all its data
type TournamentCache struct {
	ID                     int            `json:"id"`
	Name                   string         `json:"name"`
	Type                   string         `json:"type"`
	StartDate              string         `json:"startDate"`
	EndDate                string         `json:"endDate"`
	Address                models.Address `json:"address"`
	Club                   Club           `json:"club"`
	Rules                  *Rules         `json:"rules,omitempty"`
	Endowment              int            `json:"endowment"`
	IsRulesPdfChecked      bool           `json:"isRulesPdfChecked,omitempty"`
	IsSiteExistenceChecked bool           `json:"isSiteExistenceChecked,omitempty"`
	SiteUrl                string         `json:"siteUrl,omitempty"`
	SignupUrl              string         `json:"signupUrl,omitempty"`
	Timestamp              time.Time      `json:"timestamp"`
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
