package tournaments

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
	"tournois-tt/api/pkg/fftt"
	"tournois-tt/api/pkg/geocoding"
)

// refresh fetches and processes tournament addresses
func refreshGeocoding(startDateAfter, startDateBefore *time.Time) error {
	// Load existing cache - no need to store it in a variable as we'll use the thread-safe methods
	_, err := geocoding.LoadGeocodeResultsFromCache()
	if err != nil {
		log.Printf("Warning: Failed to load geocoding cache: %v", err)
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

	var tournaments []fftt.Tournament

	if err := json.NewDecoder(resp.Body).Decode(&tournaments); err != nil {
		return fmt.Errorf("failed to decode tournaments: %v", err)
	}

	// Prepare addresses for geocoding
	addressesToGeocode := make([]geocoding.Address, 0)
	geocodeResults := make([]geocoding.GeocodeResult, 0, len(tournaments))
	successCount := 0
	failureCount := 0

	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	for _, t := range tournaments {
		wg.Add(1)

		go func(t fftt.Tournament) {
			defer wg.Done()

			if !t.Address.IsValid() {
				result := geocoding.GeocodeResult{
					Address:   t.Address,
					Failed:    true,
					Timestamp: time.Now(),
				}

				// Protect geocodeResults with a mutex for thread-safety
				mu.Lock()
				geocodeResults = append(geocodeResults, result)
				geocoding.SetCachedGeocodeResult(result)
				mu.Unlock()

				return
			}

			// Check if address is already in cache using thread-safe method
			cachedResult, exists := geocoding.GetCachedGeocodeResult(t.Address)
			if exists {
				// Protect geocodeResults with a mutex for thread-safety
				mu.Lock()
				geocodeResults = append(geocodeResults, cachedResult)
				mu.Unlock()
				return
			}

			mu.Lock()
			addressesToGeocode = append(addressesToGeocode, t.Address)
			mu.Unlock()
		}(t)
	}
	wg.Wait()

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

		// Store result in the thread-safe cache
		geocoding.SetCachedGeocodeResult(result)
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
