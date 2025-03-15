package google

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// We need to reference the main geocoding package types, so we'll define them here
// These should match exactly with the main package definitions

type Address struct {
	StreetAddress             string
	PostalCode                string
	AddressLocality           string
	DisambiguatingDescription string
	Latitude                  float64
	Longitude                 float64
	Failed                    bool
}

type Location struct {
	Lat    float64
	Lon    float64
	Failed bool
}

const (
	// BaseURL is the Google Geocoding API endpoint
	BaseURL = "https://maps.googleapis.com/maps/api/geocode/json"
)

// httpClient is the client used for HTTP requests
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

// Provider implements the geocoding provider interface for Google
type Provider struct{}

// NewProvider creates a new Google geocoding provider
func NewProvider() *Provider {
	return &Provider{}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "Google"
}

// constructFullAddress creates a standardized address string for geocoding
func constructFullAddress(addr Address) string {
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

// GetCoordinates geocodes an address using Google Geocoding API
func (p *Provider) GetCoordinates(address Address) (Location, error) {
	fullAddress := constructFullAddress(address)
	return p.geocodeWithGoogle(fullAddress)
}

// geocodeWithGoogle attempts to geocode an address using Google Geocoding API
func (p *Provider) geocodeWithGoogle(fullAddress string) (Location, error) {
	// Get API key from environment
	apiKey := os.Getenv("GOOGLE_GEOCODING_API_KEY")
	if apiKey == "" {
		return Location{Failed: true}, fmt.Errorf("Google Geocoding API key not set")
	}

	// Prepare URL
	params := url.Values{}
	params.Add("address", fullAddress)
	params.Add("key", apiKey)

	// Create request
	req, err := http.NewRequest("GET", BaseURL+"?"+params.Encode(), nil)
	if err != nil {
		return Location{Failed: true}, fmt.Errorf("request creation error: %v", err)
	}

	// Send request
	resp, err := httpClient.Do(req)
	if err != nil {
		return Location{Failed: true}, fmt.Errorf("network error: %v", err)
	}
	defer resp.Body.Close()

	// Read raw response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return Location{Failed: true}, fmt.Errorf("error reading response body: %v", err)
	}

	// Recreate response body reader as it was consumed
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return Location{Failed: true}, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	// Parse response
	var googleResp struct {
		Results []struct {
			Geometry struct {
				Location struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"location"`
			} `json:"geometry"`
			Status string `json:"status"`
		} `json:"results"`
		Status string `json:"status"`
		Error  string `json:"error_message,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&googleResp); err != nil {
		return Location{Failed: true}, fmt.Errorf("parsing error: %v", err)
	}

	// Check response status
	if googleResp.Status != "OK" || len(googleResp.Results) == 0 {
		return Location{Failed: true}, fmt.Errorf("no coordinates found for address: %s", fullAddress)
	}

	log.Printf("Geocoded address with Google: %s -> (%.6f, %.6f)",
		fullAddress,
		googleResp.Results[0].Geometry.Location.Lat,
		googleResp.Results[0].Geometry.Location.Lng)

	return Location{
		Lat:    googleResp.Results[0].Geometry.Location.Lat,
		Lon:    googleResp.Results[0].Geometry.Location.Lng,
		Failed: false,
	}, nil
}
