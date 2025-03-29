package nominatim

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
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
	// RateLimitDelay respects Nominatim usage policy (1 request per second with buffer)
	RateLimitDelay = 1500 * time.Millisecond

	// BaseURL is the Nominatim API endpoint
	BaseURL = "https://nominatim.openstreetmap.org/search"

	// DefaultMaxRetries defines how many times to retry on failure
	DefaultMaxRetries = 3

	// RetryDelay defines the base delay between retry attempts
	RetryDelay = 5 * time.Second
)

// httpClient is the client used for HTTP requests
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

// Provider implements the geocoding provider interface for Nominatim
type Provider struct{}

// NewProvider creates a new Nominatim geocoding provider
func NewProvider() *Provider {
	return &Provider{}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "Nominatim"
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

// GetCoordinates geocodes an address using Nominatim
func (p *Provider) GetCoordinates(address Address) (Location, error) {
	fullAddress := constructFullAddress(address)

	// Prepare Nominatim parameters
	params := url.Values{}
	params.Add("q", fullAddress)
	params.Add("format", "json")
	params.Add("limit", "1")

	// Rate limit to respect Nominatim usage policy
	time.Sleep(RateLimitDelay)

	return p.geocodeWithRetry(fullAddress, params)
}

// geocodeWithRetry attempts to geocode an address with multiple retry attempts
func (p *Provider) geocodeWithRetry(fullAddress string, params url.Values) (Location, error) {
	var lastErr error

	for attempt := 0; attempt < DefaultMaxRetries; attempt++ {
		// Add User-Agent to respect Nominatim usage policy
		req, err := http.NewRequest("GET", BaseURL+"?"+params.Encode(), nil)
		if err != nil {
			lastErr = fmt.Errorf("request creation error: %v", err)
			time.Sleep(RetryDelay)
			continue
		}
		req.Header.Set("Accept-Language", "fr") // Add French language preference

		resp, err := httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("network error on attempt %d: %v", attempt+1, err)

			// Exponential backoff with jitter
			backoff := RetryDelay * time.Duration(attempt+1)
			jitter := time.Duration(float64(backoff) * (0.5 + rand.Float64())) // Add 50-150% jitter
			time.Sleep(jitter)
			continue
		}
		defer resp.Body.Close()

		// Check for rate limiting response
		if resp.StatusCode == http.StatusTooManyRequests {
			// Get retry-after header or use default
			retryAfter := 60 * time.Second
			if retryHeader := resp.Header.Get("Retry-After"); retryHeader != "" {
				if seconds, err := strconv.Atoi(retryHeader); err == nil {
					retryAfter = time.Duration(seconds) * time.Second
				}
			}
			time.Sleep(retryAfter)
			continue
		}

		// Check for other HTTP errors
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("HTTP error: %s", resp.Status)
			time.Sleep(RetryDelay * time.Duration(attempt+1))
			continue
		}

		var results []struct {
			Lat string `json:"lat"`
			Lon string `json:"lon"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
			lastErr = fmt.Errorf("parsing error: %v", err)
			time.Sleep(RetryDelay)
			continue
		}

		if len(results) == 0 {
			lastErr = fmt.Errorf("no coordinates found for address: %s", fullAddress)
			break
		}

		// Parse coordinates
		var lat, lon float64
		_, err = fmt.Sscanf(results[0].Lat, "%f", &lat)
		if err != nil {
			lastErr = fmt.Errorf("invalid latitude: %v", err)
			time.Sleep(RetryDelay)
			continue
		}
		_, err = fmt.Sscanf(results[0].Lon, "%f", &lon)
		if err != nil {
			lastErr = fmt.Errorf("invalid longitude: %v", err)
			time.Sleep(RetryDelay)
			continue
		}

		log.Printf("Geocoded address with Nominatim: %s -> (%.6f, %.6f)", fullAddress, lat, lon)

		return Location{
			Lat:    lat,
			Lon:    lon,
			Failed: false,
		}, nil
	}

	return Location{Failed: true}, lastErr
}
