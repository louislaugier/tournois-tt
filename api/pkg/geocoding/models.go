package geocoding

import (
	"time"
)

// GeocodeResult represents a cached geocoding result
type GeocodeResult struct {
	Address   Address   `json:"address"`
	Latitude  float64   `json:"latitude,omitempty"`
	Longitude float64   `json:"longitude,omitempty"`
	Failed    bool      `json:"failed"`
	Timestamp time.Time `json:"timestamp"`
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
