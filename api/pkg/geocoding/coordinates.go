package geocoding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	nominatimBaseURL     = "https://nominatim.openstreetmap.org/search"
	googleGeocodeBaseURL = "https://maps.googleapis.com/maps/api/geocode/json"
	defaultMaxRetries    = 3
	retryDelay           = 5 * time.Second
	// Nominatim usage policy requires 1 request per second
	rateLimitDelay       = 1500 * time.Millisecond // Add buffer to be safe
	maxConsecutiveErrors = 5
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

// constructFullAddress creates a standardized address string
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

// geocodeWithRetry attempts to geocode an address with multiple retry attempts
func geocodeWithRetry(fullAddress string, params url.Values) (Location, error) {
	var lastErr error

	for attempt := 0; attempt < defaultMaxRetries; attempt++ {
		// Add User-Agent to respect Nominatim usage policy
		req, err := http.NewRequest("GET", nominatimBaseURL+"?"+params.Encode(), nil)
		if err != nil {
			lastErr = fmt.Errorf("request creation error: %v", err)
			time.Sleep(retryDelay)
			continue
		}
		req.Header.Set("User-Agent", "TournoisTT/1.0")
		req.Header.Set("Accept-Language", "fr") // Add French language preference

		resp, err := httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("network error on attempt %d: %v", attempt+1, err)

			// Additional network error diagnostics
			if netErr, ok := err.(net.Error); ok {
				log.Printf("Network error details:")
				log.Printf("Timeout: %v", netErr.Timeout())
				log.Printf("Temporary: %v", netErr.Temporary())
			}

			// Exponential backoff with jitter
			backoff := retryDelay * time.Duration(attempt+1)
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
			log.Printf("Rate limited. Waiting %v before retry", retryAfter)
			time.Sleep(retryAfter)
			continue
		}

		// Check for other HTTP errors
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("HTTP error: %s", resp.Status)
			log.Printf("HTTP error for address %s: %s", fullAddress, resp.Status)
			time.Sleep(retryDelay * time.Duration(attempt+1))
			continue
		}

		var results []struct {
			Lat string `json:"lat"`
			Lon string `json:"lon"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
			lastErr = fmt.Errorf("parsing error: %v", err)
			log.Printf("Parsing attempt %d failed: %v", attempt+1, lastErr)
			time.Sleep(retryDelay)
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
			time.Sleep(retryDelay)
			continue
		}
		_, err = fmt.Sscanf(results[0].Lon, "%f", &lon)
		if err != nil {
			lastErr = fmt.Errorf("invalid longitude: %v", err)
			time.Sleep(retryDelay)
			continue
		}

		log.Printf("Geocoded address: %s -> (%.6f, %.6f)", fullAddress, lat, lon)

		return Location{
			Lat:    lat,
			Lon:    lon,
			Failed: false,
		}, nil
	}

	log.Printf("Geocoding completely failed for address: %s after %d attempts", fullAddress, defaultMaxRetries)
	return Location{Failed: true}, lastErr
}

// geocodeWithGoogle attempts to geocode an address using Google Geocoding API
func geocodeWithGoogle(fullAddress string) (Location, error) {
	// Get API key from environment
	apiKey := os.Getenv("GOOGLE_GEOCODING_API_KEY")
	if apiKey == "" {
		log.Printf("DEBUG: Google Geocoding API key is empty")
		return Location{Failed: true}, fmt.Errorf("Google Geocoding API key not set")
	}

	// Prepare URL
	params := url.Values{}
	params.Add("address", fullAddress)
	params.Add("key", apiKey)

	log.Printf("DEBUG: Attempting to geocode address via Google API: %s", fullAddress)
	log.Printf("DEBUG: Full request URL: %s", googleGeocodeBaseURL+"?"+params.Encode())

	// Create request
	req, err := http.NewRequest("GET", googleGeocodeBaseURL+"?"+params.Encode(), nil)
	if err != nil {
		log.Printf("DEBUG: Request creation error: %v", err)
		return Location{Failed: true}, fmt.Errorf("request creation error: %v", err)
	}

	// Send request
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("DEBUG: Network error: %v", err)
		return Location{Failed: true}, fmt.Errorf("network error: %v", err)
	}
	defer resp.Body.Close()

	// Log response status
	log.Printf("DEBUG: Response status code: %d", resp.StatusCode)

	// Read raw response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("DEBUG: Error reading response body: %v", err)
		return Location{Failed: true}, fmt.Errorf("error reading response body: %v", err)
	}
	bodyString := string(bodyBytes)
	log.Printf("DEBUG: Raw response body: %s", bodyString)

	// Recreate response body reader as it was consumed
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Check response status
	if resp.StatusCode != http.StatusOK {
		log.Printf("DEBUG: HTTP error status: %s", resp.Status)
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

	log.Printf("DEBUG: Attempting to decode JSON response")
	if err := json.NewDecoder(resp.Body).Decode(&googleResp); err != nil {
		log.Printf("DEBUG: JSON parsing error: %v", err)
		return Location{Failed: true}, fmt.Errorf("parsing error: %v", err)
	}

	// Log detailed decoding results
	log.Printf("DEBUG: Decoded response - Status: %s, Results count: %d",
		googleResp.Status, len(googleResp.Results))
	if googleResp.Error != "" {
		log.Printf("DEBUG: Google API Error Message: %s", googleResp.Error)
	}

	// Check response status
	if googleResp.Status != "OK" || len(googleResp.Results) == 0 {
		log.Printf("DEBUG: No coordinates found for address: %s", fullAddress)
		return Location{Failed: true}, fmt.Errorf("no coordinates found for address: %s", fullAddress)
	}

	log.Printf("DEBUG: Geocoded address via Google API: %s -> (%.6f, %.6f)",
		fullAddress,
		googleResp.Results[0].Geometry.Location.Lat,
		googleResp.Results[0].Geometry.Location.Lng)

	return Location{
		Lat:    googleResp.Results[0].Geometry.Location.Lat,
		Lon:    googleResp.Results[0].Geometry.Location.Lng,
		Failed: false,
	}, nil
}

// GetCoordinates attempts to geocode a single address
func GetCoordinates(address Address) (Location, error) {
	fullAddress := constructFullAddress(address)

	// First try Nominatim
	params := url.Values{}
	params.Add("q", fullAddress)
	params.Add("format", "json")
	params.Add("limit", "1")

	location, err := geocodeWithRetry(fullAddress, params)

	// If Nominatim fails, try Google Geocoding API
	if err != nil {
		log.Printf("Nominatim geocoding failed, falling back to Google Geocoding API: %v", err)
		location, err = geocodeWithGoogle(fullAddress)
	}

	return location, err
}
