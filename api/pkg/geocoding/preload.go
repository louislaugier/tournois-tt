package geocoding

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"tournois-tt/api/pkg/fftt"
)

// GeocodeConfig allows configuring geocoding behavior
type GeocodeConfig struct {
	Enabled             bool
	MaxGeocodeAttempts  int
	SkipFailedAddresses bool
}

// DefaultGeocodeConfig provides default geocoding configuration
var DefaultGeocodeConfig = GeocodeConfig{
	Enabled:             true,
	MaxGeocodeAttempts:  3,
	SkipFailedAddresses: true,
}

// PreloadTournaments fetches and processes tournament addresses
func PreloadTournaments() error {
	log.Printf("Starting tournament preloading...")

	// Load existing cache
	existingCache, err := loadGeocodeResultsFromCache()
	if err != nil {
		log.Printf("Warning: Failed to load existing cache: %v", err)
		existingCache = make(map[string]GeocodeResult)
	}

	// Create query params for future tournaments
	queryParams := url.Values{}
	queryParams.Set("startDate[after]", time.Now().Format("2006-01-02T15:04:05"))
	queryParams.Set("itemsPerPage", "999999")
	queryParams.Set("order[startDate]", "asc")

	// Call FFTT API
	resp, err := fftt.GetClient().GetTournaments(queryParams)
	if err != nil {
		return fmt.Errorf("failed to fetch tournaments: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("FFTT API returned status %d", resp.StatusCode)
	}

	var tournaments []struct {
		Address Address `json:"address"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tournaments); err != nil {
		return fmt.Errorf("failed to decode tournaments: %v", err)
	}

	log.Printf("Found %d tournaments to process", len(tournaments))

	// Prepare addresses for geocoding
	addressesToGeocode := make([]Address, 0)
	geocodeResults := make([]GeocodeResult, 0, len(tournaments))

	for _, t := range tournaments {
		if !t.Address.IsValid() {
			log.Printf("Skipping invalid address: %+v", t.Address)
			geocodeResults = append(geocodeResults, GeocodeResult{
				Address:   t.Address,
				Failed:    true,
				Timestamp: time.Now(),
			})
			continue
		}

		// Check if address is already in cache
		cacheKey := generateCacheKey(t.Address)
		if cachedResult, exists := existingCache[cacheKey]; exists {
			log.Printf("Using cached address: %s", cacheKey)
			geocodeResults = append(geocodeResults, cachedResult)
			continue
		}

		addressesToGeocode = append(addressesToGeocode, t.Address)
	}

	// Perform bulk geocoding
	bulkLocations := BulkGeocodeAddresses(addressesToGeocode)

	// Process geocoding results
	successCount := 0
	failureCount := 0

	for i, addr := range addressesToGeocode {
		result := GeocodeResult{
			Address:   addr,
			Timestamp: time.Now(),
		}

		location := bulkLocations[i]
		if location.Failed {
			log.Printf("Failed to geocode address: %s, %s %s",
				strings.TrimSpace(addr.StreetAddress),
				strings.TrimSpace(addr.PostalCode),
				strings.TrimSpace(addr.AddressLocality))

			result.Failed = true
			failureCount++
		} else {
			log.Printf("Processed address: %s, %s %s (%.6f, %.6f)",
				strings.TrimSpace(addr.StreetAddress),
				strings.TrimSpace(addr.PostalCode),
				strings.TrimSpace(addr.AddressLocality),
				location.Lat,
				location.Lon)

			result.Latitude = location.Lat
			result.Longitude = location.Lon
			result.Failed = false
			successCount++
		}

		geocodeResults = append(geocodeResults, result)
	}

	if len(addressesToGeocode) > 0 {
		// Save geocoding results to cache
		if err := saveGeocodeResultsToCache(geocodeResults); err != nil {
			log.Printf("Warning: Failed to save geocoding cache: %v", err)
		}
	}

	log.Printf("Preloading completed: %d successful, %d failed", successCount, failureCount)
	return nil
}

// generateCacheKey creates a unique key for an address
func generateCacheKey(addr Address) string {
	return fmt.Sprintf("%s|%s|%s",
		strings.TrimSpace(addr.StreetAddress),
		strings.TrimSpace(addr.PostalCode),
		strings.TrimSpace(addr.AddressLocality))
}
