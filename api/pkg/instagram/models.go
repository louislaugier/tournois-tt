package instagram

import "time"

// Config holds Instagram API configuration
type Config struct {
	// AccessToken is the Instagram Graph API access token
	AccessToken string
	// PageID is the Instagram Business Account ID (or IGSID)
	PageID string
	// Enabled determines if Instagram DM sending is enabled
	Enabled bool
}

// TournamentImage represents the data needed to generate a tournament image
type TournamentImage struct {
	Name          string
	Type          string
	Club          string
	Endowment     int
	StartDate     string
	EndDate       string
	Address       string
	RulesURL      string
	TournamentID  int
	TournamentURL string
}

// ErrorResponse represents an error from Instagram API
type ErrorResponse struct {
	Error struct {
		Message   string `json:"message"`
		Type      string `json:"type"`
		Code      int    `json:"code"`
		FBTraceID string `json:"fbtrace_id"`
	} `json:"error"`
}

// TournamentNotification represents a notification about a new tournament
type TournamentNotification struct {
	Tournament TournamentImage
	SentAt     time.Time
	MessageID  string
	Success    bool
	Error      string
}
