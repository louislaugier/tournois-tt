package geocoding

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const (
	cacheDir      = "cache"
	cacheFileName = "geocoding_cache.json"
)

func removeAccents(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, s)
	return result
}

func canonicalAddress(address string) string {
	// Convert to lowercase and trim spaces
	addr := strings.ToLower(strings.TrimSpace(address))

	// Remove accents
	addr = removeAccents(addr)

	// Remove postal code pattern (5 digits)
	addr = regexp.MustCompile(`\b\d{5}\b`).ReplaceAllString(addr, "")

	// Remove street numbers
	addr = regexp.MustCompile(`\b\d+(?:[-/]\d+)?\b`).ReplaceAllString(addr, "")

	// Remove common prefixes
	prefixes := []string{
		"gymnase", "salle", "complexe", "espace", "complexe sportif",
		"salle de tennis", "salle des sports", "gymnase du", "gymnase de",
		"salle de", "salle du", "complexe du", "espace du",
	}
	for _, prefix := range prefixes {
		addr = regexp.MustCompile(`(?i)^`+prefix+`\s+`).ReplaceAllString(addr, "")
	}

	// Remove extra spaces
	addr = regexp.MustCompile(`\s+`).ReplaceAllString(addr, " ")

	// Remove trailing commas, dashes, and spaces
	addr = strings.Trim(addr, ", -")

	return addr
}

func (c *GeocodingCache) SaveToFile() error {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}

	c.LastSave = time.Now()
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(cacheDir, cacheFileName), data, 0644)
}

func LoadCacheFromFile() (*GeocodingCache, error) {
	cache := &GeocodingCache{
		Locations: make(map[string]Location),
		Aliases:   make(map[string]string),
	}

	data, err := os.ReadFile(filepath.Join(cacheDir, cacheFileName))
	if err != nil {
		if os.IsNotExist(err) {
			return cache, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cache); err != nil {
		return nil, err
	}

	return cache, nil
}

func (c *GeocodingCache) Get(address string) (Location, bool) {
	canonical := canonicalAddress(address)

	// Try direct lookup by canonical form
	if loc, exists := c.Locations[canonical]; exists {
		return loc, true
	}

	// Try lookup through aliases
	if canon, exists := c.Aliases[canonical]; exists {
		if loc, exists := c.Locations[canon]; exists {
			return loc, true
		}
	}

	return Location{}, false
}

func (c *GeocodingCache) Set(address string, lat, lon float64, failed, approximate bool) {
	canonical := canonicalAddress(address)
	now := time.Now()

	// Store the location
	c.Locations[canonical] = Location{
		Lat:         lat,
		Lon:         lon,
		Failed:      failed,
		Approximate: approximate,
		LastUpdated: now,
	}

	// Store alias only if different from canonical and not already a canonical key
	if address != canonical {
		nonCanon := canonicalAddress(address)
		if _, exists := c.Locations[nonCanon]; !exists {
			c.Aliases[nonCanon] = canonical
		}
	}
}
