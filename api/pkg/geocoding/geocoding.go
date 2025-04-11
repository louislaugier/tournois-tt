package geocoding

import (
	"log"
	"strings"
	"time"

	"tournois-tt/api/pkg/geocoding/google"
	"tournois-tt/api/pkg/geocoding/nominatim"
)

// Debug enables verbose logging
var Debug = false

// debugLog logs a message if Debug is enabled
func debugLog(format string, args ...interface{}) {
	if Debug {
		log.Printf(format, args...)
	}
}

// RateLimitDelay is used for respecting Nominatim usage policy (1 request per second)
const RateLimitDelay = 1500 * time.Millisecond

// Provider defines an interface for geocoding providers
type Provider interface {
	GetCoordinates(address Address) (Location, error)
	Name() string
}

// nominatimProvider is the cached Nominatim provider instance
var nominatimProvider Provider

// googleProvider is the cached Google provider instance
var googleProvider Provider

// GetCoordinatesFunc is a function type for geocoding operations
type GetCoordinatesFunc func(address Address) (Location, error)

// GetCoordinatesFn is the current implementation of GetCoordinates
// This can be replaced in tests to mock geocoding
var GetCoordinatesFn GetCoordinatesFunc = getCoordinatesImpl

// init initializes the geocoding providers
func init() {
	nominatimProvider = &nominatimAdapter{provider: nominatim.NewProvider()}
	googleProvider = &googleAdapter{provider: google.NewProvider()}
}

// ConstructFullAddress formats an address into a single string
func ConstructFullAddress(addr Address) string {
	parts := []string{}

	if addr.StreetAddress != "" {
		parts = append(parts, strings.TrimSpace(addr.StreetAddress))
	}

	if addr.PostalCode != "" {
		parts = append(parts, strings.TrimSpace(addr.PostalCode))
	}

	if addr.AddressLocality != "" {
		parts = append(parts, strings.TrimSpace(addr.AddressLocality))
	}

	if addr.DisambiguatingDescription != "" {
		parts = append(parts, strings.TrimSpace(addr.DisambiguatingDescription))
	}

	return strings.Join(parts, ", ")
}

// CreateGeocodeResult creates a geocode result from an address and location
func CreateGeocodeResult(address Address, location Location, err error) GeocodeResult {
	result := GeocodeResult{
		Address:   address,
		Timestamp: time.Now(),
	}

	if err != nil {
		result.Failed = true
		return result
	}

	result.Latitude = location.Lat
	result.Longitude = location.Lon
	result.Failed = location.Failed

	// Update the address with the location data
	result.Address.Latitude = location.Lat
	result.Address.Longitude = location.Lon
	result.Address.Failed = location.Failed

	return result
}

// GetCoordinates gets coordinates for an address
func GetCoordinates(address Address) (Location, error) {
	return getCoordinatesImpl(address)
}

// getCoordinatesImpl is the implementation of GetCoordinates
func getCoordinatesImpl(address Address) (Location, error) {
	// Try with Nominatim first
	nominatimResult, err := GetCoordinatesNominatim(address)
	if err == nil && !nominatimResult.Failed {
		return Location{Lat: nominatimResult.Address.Latitude, Lon: nominatimResult.Address.Longitude}, nil
	}

	// Fall back to Google as a backup
	googleResult, err := GetCoordinatesGoogle(address)
	if err != nil {
		return Location{Failed: true}, err
	}

	return Location{Lat: googleResult.Address.Latitude, Lon: googleResult.Address.Longitude, Failed: googleResult.Failed}, nil
}

// GetCoordinatesNominatim gets coordinates using Nominatim
func GetCoordinatesNominatim(address Address) (GeocodeResult, error) {
	// Use the existing nominatimProvider
	location, err := nominatimProvider.GetCoordinates(address)
	if err != nil {
		return GeocodeResult{}, err
	}

	// Create and return a GeocodeResult
	return CreateGeocodeResult(address, location, nil), nil
}

// GetCoordinatesGoogle gets coordinates using Google Maps API
func GetCoordinatesGoogle(address Address) (GeocodeResult, error) {
	// Use the existing googleProvider
	location, err := googleProvider.GetCoordinates(address)
	if err != nil {
		return GeocodeResult{}, err
	}

	// Create and return a GeocodeResult
	return CreateGeocodeResult(address, location, nil), nil
}

// nominatimAdapter adapts the Nominatim provider to the Provider interface
type nominatimAdapter struct {
	provider *nominatim.Provider
}

// GetCoordinates implements the Provider interface for Nominatim
func (a *nominatimAdapter) GetCoordinates(address Address) (Location, error) {
	// Convert to the provider's address format
	providerAddress := nominatim.Address{
		StreetAddress:   address.StreetAddress,
		PostalCode:      address.PostalCode,
		AddressLocality: address.AddressLocality,
	}

	// Get coordinates from provider
	result, err := a.provider.GetCoordinates(providerAddress)
	if err != nil {
		debugLog("Nominatim geocoding error: %v", err)
		return Location{Failed: true}, err
	}

	// Convert back to our Location type
	location := Location{
		Lat:    result.Lat,
		Lon:    result.Lon,
		Failed: false,
	}

	debugLog("Nominatim geocoded [%s] to lat:%f, lon:%f",
		ConstructFullAddress(address),
		location.Lat,
		location.Lon)

	return location, nil
}

// Name returns the provider name
func (a *nominatimAdapter) Name() string {
	return "Nominatim"
}

// googleAdapter adapts the Google provider to the Provider interface
type googleAdapter struct {
	provider *google.Provider
}

// GetCoordinates implements the Provider interface for Google
func (a *googleAdapter) GetCoordinates(address Address) (Location, error) {
	// Convert to the provider's address format
	providerAddress := google.Address{
		StreetAddress:   address.StreetAddress,
		PostalCode:      address.PostalCode,
		AddressLocality: address.AddressLocality,
	}

	// Get coordinates from provider
	result, err := a.provider.GetCoordinates(providerAddress)
	if err != nil {
		debugLog("Google geocoding error: %v", err)
		return Location{Failed: true}, err
	}

	// Convert back to our Location type
	location := Location{
		Lat:    result.Lat,
		Lon:    result.Lon,
		Failed: false,
	}

	debugLog("Google geocoded [%s] to lat:%f, lon:%f",
		ConstructFullAddress(address),
		location.Lat,
		location.Lon)

	return location, nil
}

// Name returns the provider name
func (a *googleAdapter) Name() string {
	return "Google"
}
