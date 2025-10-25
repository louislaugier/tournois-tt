package fftt

import "tournois-tt/api/pkg/geocoding"

// Tournament represents a complete tournament
type Tournament struct {
	ID                     int               `json:"id"`
	Name                   string            `json:"name"`
	Type                   string            `json:"type"`
	StartDate              string            `json:"startDate"`
	EndDate                string            `json:"endDate"`
	Address                geocoding.Address `json:"address"`
	Club                   Club              `json:"club"`
	Rules                  *Rules            `json:"rules"`
	Tables                 []Table           `json:"tables"`
	Endowment              int               `json:"endowment"`
	Organization           *Organization     `json:"organization,omitempty"`
	Responses              []Response        `json:"responses,omitempty"`
	IsRulesPdfChecked      bool              `json:"isRulesPdfChecked,omitempty"`
	IsSiteExistenceChecked bool              `json:"isSiteExistenceChecked,omitempty"`
	SiteURL                string            `json:"siteUrl,omitempty"`
	SignupURL              string            `json:"signupUrl,omitempty"`
	Page                   string            `json:"page,omitempty"`
}

// Response represents tournament responses
type Response struct {
	PlayerID  int    `json:"playerId"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// Club represents a table tennis club
type Club struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Code       string `json:"code"`
	Department string `json:"department"`
	Region     string `json:"region"`
	Identifier string `json:"identifier"`
}

// Rules represents tournament rules
type Rules struct {
	AgeMin  int    `json:"ageMin"`
	AgeMax  int    `json:"ageMax"`
	Points  int    `json:"points"`
	Ranking int    `json:"ranking"`
	URL     string `json:"url,omitempty"`
}

// Table represents a tournament table
type Table struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Date        string `json:"date"`
	Time        string `json:"time"`
	Fee         int    `json:"fee"`
	Endowment   int    `json:"endowment"`
}

// Organization represents tournament organization details
type Organization struct {
	Name    string `json:"name"`
	Contact string `json:"contact"`
}
