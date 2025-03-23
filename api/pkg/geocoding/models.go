package geocoding

import (
	"time"

	"tournois-tt/api/pkg/models"
)

// For backward compatibility - use models.Address and models.Location directly

// GeocodeResult represents a cached geocoding result
type GeocodeResult struct {
	Address   models.Address `json:"address"`
	Latitude  float64        `json:"latitude,omitempty"`
	Longitude float64        `json:"longitude,omitempty"`
	Failed    bool           `json:"failed"`
	Timestamp time.Time      `json:"timestamp"`
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
