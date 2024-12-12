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
	Set(address string, lat, lon float64, failed, approximate bool, variants []string, tournamentName ...string)
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
func (c *RuntimeCache) Set(addr string, lat, lon float64, failed, approximate bool, variants []string, tournamentName ...string) {
	c.Lock()
	defer c.Unlock()

	// Normalize all addresses
	normalizedAddr := Normalize(addr)
	normalizedVariants := make([]string, 0)
	for _, v := range variants {
		normalizedVariants = append(normalizedVariants, Normalize(v))
	}

	allAddresses := append([]string{normalizedAddr}, normalizedVariants...)

	// Find the most complete address to use as the base key
	baseAddr := normalizedAddr
	for _, a := range allAddresses {
		// Look for addresses that contain the current baseAddr
		// This means they are more complete versions of the same address
		if strings.Contains(a, baseAddr) && len(a) > len(baseAddr) {
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

	// Check if the address is approximate (missing both street number and venue name)
	hasStreetNumber := regexp.MustCompile(`\d+`).MatchString(baseAddr)
	hasVenueName := strings.Contains(baseAddr, "gymnase") ||
		strings.Contains(baseAddr, "salle") ||
		strings.Contains(baseAddr, "complexe") ||
		strings.Contains(baseAddr, "espace") ||
		strings.Contains(baseAddr, "stade")

	// Location is approximate only if both street number and venue name are missing
	isApproximate := !hasStreetNumber && !hasVenueName

	// Store the location with the base address as key
	c.Locations[baseAddr] = Location{
		Lat:         lat,
		Lon:         lon,
		Failed:      failed,
		Approximate: isApproximate,
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

// Geocode performs geocoding for an address with more robust error handling
func (c *NominatimClient) Geocode(address string) (Location, error) {
	params := url.Values{}
	params.Add("q", address)
	params.Add("format", "json")
	params.Add("limit", "1")
	params.Add("addressdetails", "1") // Get more detailed address information

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
		// Log the raw response for debugging
		log.Printf("Failed to parse Nominatim response for address %s: %v\nRaw response: %s",
			address, err, string(body))
		return Location{}, fmt.Errorf("error parsing response: %v", err)
	}

	if len(result) == 0 {
		log.Printf("No geocoding result found for address: %s", address)
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

// generateStreetKeywords dynamically generates street keywords from the address
func generateStreetKeywords(address string) []string {
	// Common French street type prefixes and keywords
	baseKeywords := []string{
		"rue", "avenue", "boulevard", "place",
		"chemin", "impasse", "route", "allee",
		"square", "passage", "quai", "cours",
		"allée", "voie", "impasse", "esplanade",
		"mail", "promenade", "rond-point", "carrefour",
	}

	// Extract potential keywords from the address
	addressKeywords := strings.Fields(strings.ToLower(address))

	// Combine base keywords with address-specific keywords
	keywords := make(map[string]bool)
	for _, kw := range baseKeywords {
		keywords[kw] = true
	}

	// Add unique keywords from the address that might indicate a street
	for _, word := range addressKeywords {
		// Ignore very short words
		if len(word) > 2 {
			keywords[word] = true
		}
	}

	// Convert map keys to slice
	var result []string
	for kw := range keywords {
		result = append(result, kw)
	}

	return result
}

// isApproximateResult determines if a geocoding result is approximate
func (c *NominatimClient) isApproximateResult(result struct {
	Lat      string `json:"lat"`
	Lon      string `json:"lon"`
	Type     string `json:"type"`
	Class    string `json:"class"`
	Category string `json:"category"`
}, address string) bool {
	// Be more lenient with location types
	approximateTypes := []string{
		"city", "administrative", "boundary",
		"village", "town", "municipality",
		"county", "region", "district",
	}

	for _, t := range approximateTypes {
		if result.Type == t || result.Class == t {
			return true
		}
	}

	// Dynamically generate street keywords
	streetKeywords := generateStreetKeywords(address)

	// Check for street information
	hasStreetInfo := false
	for _, keyword := range streetKeywords {
		if strings.Contains(strings.ToLower(address), keyword) {
			hasStreetInfo = true
			break
		}
	}

	// If no street info, consider it approximate
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
