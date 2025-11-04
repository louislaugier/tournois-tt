package instagram

import (
    "time"
    igimage "tournois-tt/api/pkg/image"
)

// Config holds Instagram and Threads API configuration
type Config struct {
	// AccessToken is the Instagram Graph API access token
	AccessToken string
	// PageID is the Instagram Business Account ID (or IGSID)
	PageID string
	// ThreadsAccessToken is the Threads API access token
	ThreadsAccessToken string
	// ThreadsUserID is the Threads user ID for posting
	ThreadsUserID string
	// Enabled determines if Instagram posting is enabled
	Enabled bool
	// ThreadsEnabled determines if Threads posting is enabled
	ThreadsEnabled bool
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
    Tournament igimage.TournamentImage
	SentAt     time.Time
	MessageID  string
	Success    bool
	Error      string
}
