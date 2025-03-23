package geocoding

import (
	"fmt"
	"log"
	"strings"
	"time"

	"tournois-tt/api/pkg/geocoding/google"
	"tournois-tt/api/pkg/geocoding/nominatim"
	"tournois-tt/api/pkg/models"
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

// Provider defines the interface that all geocoding providers must implement
type Provider interface {
	GetCoordinates(address models.Address) (models.Location, error)
	Name() string
}

// nominatimProvider is the cached Nominatim provider instance
var nominatimProvider Provider

// googleProvider is the cached Google provider instance
var googleProvider Provider

// GetCoordinatesFunc is the function type for the GetCoordinates function
type GetCoordinatesFunc func(address models.Address) (models.Location, error)

// GetCoordinatesFn is the current implementation of GetCoordinates
// This can be replaced in tests to mock geocoding
var GetCoordinatesFn GetCoordinatesFunc = getCoordinatesImpl

// init initializes the geocoding providers
func init() {
	nominatimProvider = &nominatimAdapter{provider: nominatim.NewProvider()}
	googleProvider = &googleAdapter{provider: google.NewProvider()}
}

// ConstructFullAddress creates a standardized address string
func ConstructFullAddress(addr models.Address) string {
	fullAddress := addr.StreetAddress

	// Add disambiguating description if street number is not in the address
	if addr.StreetAddress == "" || !strings.Contains(addr.StreetAddress, " ") {
		if addr.DisambiguatingDescription != "" {
			fullAddress = addr.DisambiguatingDescription + " " + fullAddress
		}
	}

	// Append postal code and locality
	return fmt.Sprintf("%s, %s %s, France",
		strings.TrimSpace(fullAddress),
		strings.TrimSpace(addr.PostalCode),
		strings.TrimSpace(addr.AddressLocality))
}

// CreateGeocodeResult creates a consistent GeocodeResult
func CreateGeocodeResult(address models.Address, location models.Location, err error) GeocodeResult {
	if err != nil {
		log.Printf("Geocoding error for address %v: %v", address, err)
		return GeocodeResult{
			Address:   address,
			Failed:    true,
			Latitude:  0,
			Longitude: 0,
			Timestamp: time.Now(),
		}
	}

	return GeocodeResult{
		Address:   address,
		Failed:    location.Failed,
		Latitude:  location.Lat,
		Longitude: location.Lon,
		Timestamp: time.Now(),
	}
}

// GetCoordinates attempts to geocode a single address
// It tries Nominatim first, then falls back to Google if Nominatim fails
func GetCoordinates(address models.Address) (models.Location, error) {
	return GetCoordinatesFn(address)
}

// getCoordinatesImpl is the actual implementation of GetCoordinates
func getCoordinatesImpl(address models.Address) (models.Location, error) {
	// Try with Nominatim first
	location, err := nominatimProvider.GetCoordinates(address)

	// If Nominatim fails, try Google Geocoding API
	if err != nil {
		debugLog("Nominatim geocoding failed, trying Google: %v", err)
		location, err = googleProvider.GetCoordinates(address)
	}

	return location, err
}

// GetCoordinatesNominatim attempts to geocode an address using Nominatim
func GetCoordinatesNominatim(address models.Address) (GeocodeResult, error) {
	location, err := nominatimProvider.GetCoordinates(address)
	return CreateGeocodeResult(address, location, err), err
}

// GetCoordinatesGoogle attempts to geocode an address using Google Geocoding API
func GetCoordinatesGoogle(address models.Address) (GeocodeResult, error) {
	location, err := googleProvider.GetCoordinates(address)
	return CreateGeocodeResult(address, location, err), err
}

// nominatimAdapter adapts the Nominatim provider to the common interface
type nominatimAdapter struct {
	provider *nominatim.Provider
}

// GetCoordinates implements the Provider interface for nominatimAdapter
func (a *nominatimAdapter) GetCoordinates(address models.Address) (models.Location, error) {
	// Convert Address to nominatim.Address
	nomAddr := nominatim.Address{
		StreetAddress:             address.StreetAddress,
		PostalCode:                address.PostalCode,
		AddressLocality:           address.AddressLocality,
		DisambiguatingDescription: address.DisambiguatingDescription,
		Latitude:                  address.Latitude,
		Longitude:                 address.Longitude,
		Failed:                    address.Failed,
	}

	// Call the Nominatim provider
	nomLocation, err := a.provider.GetCoordinates(nomAddr)
	if err != nil {
		return models.Location{Failed: true}, err
	}

	// Convert nominatim.Location to Location
	return models.Location{
		Lat:    nomLocation.Lat,
		Lon:    nomLocation.Lon,
		Failed: nomLocation.Failed,
	}, nil
}

// Name returns the provider name
func (a *nominatimAdapter) Name() string {
	return a.provider.Name()
}

// googleAdapter adapts the Google provider to the common interface
type googleAdapter struct {
	provider *google.Provider
}

// GetCoordinates implements the Provider interface for googleAdapter
func (a *googleAdapter) GetCoordinates(address models.Address) (models.Location, error) {
	// Convert Address to google.Address
	googleAddr := google.Address{
		StreetAddress:             address.StreetAddress,
		PostalCode:                address.PostalCode,
		AddressLocality:           address.AddressLocality,
		DisambiguatingDescription: address.DisambiguatingDescription,
		Latitude:                  address.Latitude,
		Longitude:                 address.Longitude,
		Failed:                    address.Failed,
	}

	// Call the Google provider
	googleLocation, err := a.provider.GetCoordinates(googleAddr)
	if err != nil {
		return models.Location{Failed: true}, err
	}

	// Convert google.Location to Location
	return models.Location{
		Lat:    googleLocation.Lat,
		Lon:    googleLocation.Lon,
		Failed: googleLocation.Failed,
	}, nil
}

// Name returns the provider name
func (a *googleAdapter) Name() string {
	return a.provider.Name()
}
