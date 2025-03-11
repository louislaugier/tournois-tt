package crons

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
	"tournois-tt/api/pkg/fftt"
	"tournois-tt/api/pkg/geocoding"
	"tournois-tt/api/pkg/utils"
)

func RefreshTournaments() {
	lastSeasonStart, _ := utils.GetLastFinishedSeason()
	if err := refresh(&lastSeasonStart, nil); err != nil {
		log.Printf("Warning: Failed to refresh tournament data: %v", err)
	}
}

// refresh fetches and processes tournament addresses
func refresh(startDateAfter, startDateBefore *time.Time) error {

	// Load existing cache
	existingCache, err := geocoding.LoadGeocodeResultsFromCache()
	if err != nil {
		existingCache = make(map[string]geocoding.GeocodeResult)
	}

	// Create query params for future tournaments
	queryParams := url.Values{}
	queryParams.Set("startDate[after]", startDateAfter.Format("2006-01-02T15:04:05"))
	if startDateBefore != nil {
		queryParams.Set("startDate[before]", startDateBefore.Format("2006-01-02T15:04:05"))
	}
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
		Address geocoding.Address `json:"address"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tournaments); err != nil {
		return fmt.Errorf("failed to decode tournaments: %v", err)
	}

	// Prepare addresses for geocoding
	addressesToGeocode := make([]geocoding.Address, 0)
	geocodeResults := make([]geocoding.GeocodeResult, 0, len(tournaments))
	successCount := 0
	failureCount := 0

	for _, t := range tournaments {
		if !t.Address.IsValid() {
			geocodeResults = append(geocodeResults, geocoding.GeocodeResult{
				Address:   t.Address,
				Failed:    true,
				Timestamp: time.Now(),
			})
			continue
		}

		// Check if address is already in cache
		cacheKey := geocoding.GenerateCacheKey(t.Address)
		if cachedResult, exists := existingCache[cacheKey]; exists {
			geocodeResults = append(geocodeResults, cachedResult)
			continue
		}

		addressesToGeocode = append(addressesToGeocode, t.Address)
	}

	// Perform individual geocoding
	for _, addr := range addressesToGeocode {
		result := geocoding.GeocodeResult{
			Address:   addr,
			Timestamp: time.Now(),
		}

		// Rate limit between requests
		time.Sleep(geocoding.RateLimitDelay)

		// Geocode individual address
		location, err := geocoding.GetCoordinates(addr)
		if err != nil {
			result.Failed = true
			failureCount++
		} else {
			result.Latitude = location.Lat
			result.Longitude = location.Lon
			result.Failed = false
			successCount++
		}

		geocodeResults = append(geocodeResults, result)
	}

	if len(addressesToGeocode) > 0 {
		// Save geocoding results to cache
		if err := geocoding.SaveGeocodeResultsToCache(geocodeResults); err != nil {
			log.Printf("Warning: Failed to save geocoding cache: %v", err)
		}
	}

	log.Printf("Refreshing completed: %d successful, %d failed", successCount, failureCount)
	return nil
}
