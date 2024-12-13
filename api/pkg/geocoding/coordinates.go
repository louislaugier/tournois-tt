package geocoding

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	nominatimBaseURL  = "https://nominatim.openstreetmap.org/search"
	defaultMaxRetries = 3
	retryDelay        = 5 * time.Second
	rateLimitDelay    = 1 * time.Second
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

// BulkGeocodeAddresses performs geocoding for multiple addresses with rate limiting
func BulkGeocodeAddresses(addresses []Address) []Location {
	results := make([]Location, len(addresses))

	for i, addr := range addresses {
		// Rate limit between requests
		time.Sleep(rateLimitDelay)

		// Construct full address string
		fullAddress := constructFullAddress(addr)

		// Prepare Nominatim request
		params := url.Values{}
		params.Add("q", fullAddress)
		params.Add("format", "json")
		params.Add("limit", "1")

		// Attempt geocoding with retries
		location, err := geocodeWithRetry(fullAddress, params)

		if err != nil {
			log.Printf("Failed to geocode address: %s - %v", fullAddress, err)
			results[i] = Location{Failed: true}
		} else {
			results[i] = location
		}
	}

	return results
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

		resp, err := httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("network error on attempt %d: %v", attempt+1, err)

			// Additional network error diagnostics
			if netErr, ok := err.(net.Error); ok {
				log.Printf("Network error details:")
				log.Printf("Timeout: %v", netErr.Timeout())
				log.Printf("Temporary: %v", netErr.Temporary())
			}

			// Log system network information
			ifaces, _ := net.Interfaces()
			for _, iface := range ifaces {
				addrs, _ := iface.Addrs()
				log.Printf("Interface %s addresses: %v", iface.Name, addrs)
			}

			time.Sleep(retryDelay * time.Duration(attempt+1)) // Exponential backoff
			continue
		}
		defer resp.Body.Close()

		// Check for HTTP status code
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

// GetCoordinates attempts to geocode a single address
func GetCoordinates(address Address) (Location, error) {
	fullAddress := constructFullAddress(address)

	params := url.Values{}
	params.Add("q", fullAddress)
	params.Add("format", "json")
	params.Add("limit", "1")

	location, err := geocodeWithRetry(fullAddress, params)
	return location, err
}
