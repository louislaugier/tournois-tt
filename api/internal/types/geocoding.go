package types

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
	"tournois-tt/api/pkg/geocoding/address"
)

// Normalize standardizes an address string for consistent comparison
func Normalize(addr string) string {
	return address.Normalize(addr)
}

// Location represents a geocoded location with metadata
type Location struct {
	Lat         float64   `json:"lat"`
	Lon         float64   `json:"lon"`
	Failed      bool      `json:"failed"`
	Approximate bool      `json:"approximate"`
	LastUpdated time.Time `json:"lastUpdated"`
	Variants    []string  `json:"variants,omitempty"`
}

// GeocodingCache represents the persistent cache data structure
type GeocodingCache struct {
	Locations map[string]Location `json:"Locations"`
	LastSave  time.Time           `json:"LastSave"`
}

// Cache defines the interface for geocoding cache operations
type Cache interface {
	Get(address string) (Location, bool)
	Set(address string, lat, lon float64, failed, approximate bool, variants []string)
	SaveToFile() error
	LastSaveTime() time.Time
}

// RuntimeCache represents the runtime cache with thread-safe operations
type RuntimeCache struct {
	sync.RWMutex
	*GeocodingCache
}

// Get retrieves a location from the cache
func (c *RuntimeCache) Get(addr string) (Location, bool) {
	c.RLock()
	defer c.RUnlock()

	addr = Normalize(addr)
	if loc, exists := c.Locations[addr]; exists {
		return loc, true
	}

	// Check variants
	for _, loc := range c.Locations {
		if !loc.Failed {
			for _, variant := range loc.Variants {
				if variant == addr {
					return loc, true
				}
			}
		}
	}
	return Location{}, false
}

// Set stores a location in the cache
func (c *RuntimeCache) Set(addr string, lat, lon float64, failed, approximate bool, variants []string) {
	c.Lock()
	defer c.Unlock()

	// Normalize all addresses
	normalizedAddr := Normalize(addr)
	normalizedVariants := make([]string, 0)
	for _, v := range variants {
		normalizedVariants = append(normalizedVariants, Normalize(v))
	}

	// Find the most specific (longest) address to use as the base key
	baseAddr := normalizedAddr
	allAddresses := append([]string{normalizedAddr}, normalizedVariants...)
	for _, a := range allAddresses {
		if len(a) > len(baseAddr) {
			baseAddr = a
		}
	}

	// Create variants list excluding the base address
	uniqueVariants := make([]string, 0)
	seen := make(map[string]bool)
	seen[baseAddr] = true

	// Add all other addresses as variants
	for _, a := range allAddresses {
		if !seen[a] && a != baseAddr {
			seen[a] = true
			uniqueVariants = append(uniqueVariants, a)
		}
	}

	// Remove any existing entries that are now variants
	for variant := range seen {
		if variant != baseAddr {
			delete(c.Locations, variant)
		}
	}

	// Store the location with the base address as key
	c.Locations[baseAddr] = Location{
		Lat:         lat,
		Lon:         lon,
		Failed:      failed,
		Approximate: approximate,
		LastUpdated: time.Now(),
		Variants:    uniqueVariants,
	}
}

// LastSaveTime returns the time of the last cache save
func (c *RuntimeCache) LastSaveTime() time.Time {
	c.RLock()
	defer c.RUnlock()
	return c.LastSave
}

// SaveToFile saves the cache to disk
func (c *RuntimeCache) SaveToFile() error {
	cacheDir := "cache"
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %v", err)
	}

	c.LastSave = time.Now()
	data, err := json.MarshalIndent(c.GeocodingCache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %v", err)
	}

	cacheFile := filepath.Join(cacheDir, "geocoding_cache.json")
	if err := os.WriteFile(cacheFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %v", err)
	}

	return nil
}

// NominatimClient represents a client for the Nominatim geocoding service
type NominatimClient struct {
	HTTPClient *http.Client
	BaseURL    string
}

// Geocode performs geocoding for an address
func (c *NominatimClient) Geocode(address string) (Location, error) {
	params := url.Values{}
	params.Add("q", address)
	params.Add("format", "json")
	params.Add("limit", "1")

	req, err := http.NewRequest("GET", c.BaseURL+"?"+params.Encode(), nil)
	if err != nil {
		return Location{}, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("User-Agent", "TournoisTT/1.0")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return Location{}, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Location{}, fmt.Errorf("error reading response: %v", err)
	}

	var result NominatimResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return Location{}, fmt.Errorf("error parsing response: %v", err)
	}

	if len(result) == 0 {
		return Location{Failed: true, LastUpdated: time.Now()}, nil
	}

	var lat, lon float64
	fmt.Sscanf(result[0].Lat, "%f", &lat)
	fmt.Sscanf(result[0].Lon, "%f", &lon)

	isApproximate := c.isApproximateResult(result[0], address)

	log.Printf("Found coordinates for address %s: (%.6f, %.6f) type: %s class: %s category: %s approximate: %v",
		address, lat, lon, result[0].Type, result[0].Class, result[0].Category, isApproximate)

	return Location{
		Lat:         lat,
		Lon:         lon,
		Failed:      false,
		Approximate: isApproximate,
		LastUpdated: time.Now(),
	}, nil
}

// isApproximateResult determines if a geocoding result is approximate
func (c *NominatimClient) isApproximateResult(result struct {
	Lat      string `json:"lat"`
	Lon      string `json:"lon"`
	Type     string `json:"type"`
	Class    string `json:"class"`
	Category string `json:"category"`
}, address string) bool {
	// Check if it's a city-level or administrative result
	if result.Type == "city" || result.Type == "administrative" || result.Class == "boundary" {
		return true
	}

	// Check if we're using a simplified variant without street info
	hasStreetInfo := strings.Contains(address, "rue") ||
		strings.Contains(address, "avenue") ||
		strings.Contains(address, "boulevard") ||
		strings.Contains(address, "place") ||
		strings.Contains(address, "chemin") ||
		strings.Contains(address, "impasse")

	if !hasStreetInfo {
		return true
	}

	// Check if we're using a variant without a street number
	hasNumber := regexp.MustCompile(`^\d+`).MatchString(address)
	if strings.Contains(address, " ") && !hasNumber {
		return true
	}

	return false
}

// NominatimResponse represents the response from Nominatim geocoding service
type NominatimResponse []struct {
	Lat      string `json:"lat"`
	Lon      string `json:"lon"`
	Type     string `json:"type"`
	Class    string `json:"class"`
	Category string `json:"category"`
}
