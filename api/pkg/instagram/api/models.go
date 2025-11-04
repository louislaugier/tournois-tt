package api

import (
    "time"
    igimage "tournois-tt/api/pkg/image"
)

type Config struct {
    AccessToken        string
    PageID             string
    ThreadsAccessToken string
    ThreadsUserID      string
    Enabled            bool
    ThreadsEnabled     bool
}

type ErrorResponse struct {
    Error struct {
        Message   string `json:"message"`
        Type      string `json:"type"`
        Code      int    `json:"code"`
        FBTraceID string `json:"fbtrace_id"`
    } `json:"error"`
}

type TournamentNotification struct {
    Tournament igimage.TournamentImage
    SentAt     time.Time
    MessageID  string
    Success    bool
    Error      string
}


