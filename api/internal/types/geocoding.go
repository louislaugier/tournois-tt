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
)

// Normalize standardizes an address string for consistent comparison
func Normalize(address string) string {
	address = strings.ToLower(address)
	address = strings.Join(strings.Fields(strings.ReplaceAll(address, "-", " - ")), " ")
	address = strings.ReplaceAll(address, " - ", "-")
	address = strings.TrimRight(address, ",")
	address = strings.TrimSuffix(address, ", france")
	if len(address) > 5 && strings.ContainsAny(address[:5], "0123456789") {
		postalCode := address[:5]
		rest := strings.TrimSpace(address[5:])
		address = postalCode + " " + rest
	}
	return address
}

// Location represents a geocoded location with metadata
type Location struct {
	Lat         float64   `json:"lat"`
	Lon         float64   `json:"lon"`
	Failed      bool      `json:"failed"`
	Approximate bool      `json:"approximate"`
	LastUpdated time.Time `json:"lastUpdated"`
}

// GeocodingCache represents the persistent cache data structure
type GeocodingCache struct {
	Locations   map[string]Location `json:"Locations"`
	AliasGroups [][]string          `json:"AliasGroups"`
	LastSave    time.Time           `json:"LastSave"`
}

// Cache defines the interface for geocoding cache operations
type Cache interface {
	Get(address string) (Location, bool)
	Set(address string, lat, lon float64, failed, approximate bool)
	AddAlias(alias, canonical string)
	SaveToFile() error
	LastSaveTime() time.Time
}

// RuntimeCache represents the runtime cache with thread-safe operations
type RuntimeCache struct {
	sync.RWMutex
	*GeocodingCache
	AliasMap map[string]string // runtime-only map for quick lookups
}

// Get retrieves a location from the cache
func (c *RuntimeCache) Get(addr string) (Location, bool) {
	c.RLock()
	defer c.RUnlock()

	// Normalize the address
	addr = Normalize(addr)

	// Check if this address is an alias
	if canonical, exists := c.AliasMap[addr]; exists {
		addr = canonical
	}

	// Look up the location
	loc, exists := c.Locations[addr]
	return loc, exists
}

// Set stores a location in the cache
func (c *RuntimeCache) Set(addr string, lat, lon float64, failed, approximate bool) {
	c.Lock()
	defer c.Unlock()

	// Normalize the address
	addr = Normalize(addr)

	// Store the location
	c.Locations[addr] = Location{
		Lat:         lat,
		Lon:         lon,
		Failed:      failed,
		Approximate: approximate,
		LastUpdated: time.Now(),
	}
}

// AddAlias adds an alias for an address
func (c *RuntimeCache) AddAlias(alias, canonical string) {
	c.Lock()
	defer c.Unlock()

	// Normalize both addresses
	alias = Normalize(alias)
	canonical = Normalize(canonical)

	// Don't create self-referential aliases
	if alias == canonical {
		return
	}

	// Find existing group or create new one
	found := false
	for i := range c.AliasGroups {
		for _, addr := range c.AliasGroups[i] {
			if addr == alias || addr == canonical {
				// Add both addresses to this group if not already present
				hasAlias := false
				hasCanonical := false
				for _, a := range c.AliasGroups[i] {
					if a == alias {
						hasAlias = true
					}
					if a == canonical {
						hasCanonical = true
					}
				}
				if !hasAlias {
					c.AliasGroups[i] = append(c.AliasGroups[i], alias)
				}
				if !hasCanonical {
					c.AliasGroups[i] = append(c.AliasGroups[i], canonical)
				}
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		// Create new group
		c.AliasGroups = append(c.AliasGroups, []string{alias, canonical})
	}

	// Update runtime alias map
	c.RebuildAliasMap()
}

// RebuildAliasMap rebuilds the runtime alias map
func (c *RuntimeCache) RebuildAliasMap() {
	c.AliasMap = make(map[string]string)
	for _, group := range c.AliasGroups {
		if len(group) > 0 {
			canonical := group[0] // Use first address as canonical
			for _, alias := range group[1:] {
				c.AliasMap[alias] = canonical
			}
		}
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
	c.Lock()
	defer c.Unlock()

	cacheDir := "api/cache"
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

// Tournament represents a table tennis tournament
type Tournament struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	StartDate string  `json:"startDate"`
	EndDate   string  `json:"endDate"`
	Address   Address `json:"address"`
	Club      Club    `json:"club"`
}

// Club represents a table tennis club
type Club struct {
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
}
