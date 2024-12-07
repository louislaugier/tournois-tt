package models

import "time"

// Base contains common fields for all models
type Base struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Response is a generic response structure
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse is used for error responses
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
