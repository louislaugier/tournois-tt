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

// PreloadTournaments fetches and processes tournament addresses
func PreloadTournaments() error {
	log.Printf("Starting tournament preloading...")

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

	// Process addresses
	successCount := 0
	failureCount := 0
	var geocodeResults []GeocodeResult

	for _, t := range tournaments {
		if !t.Address.IsValid() {
			log.Printf("Skipping invalid address: %+v", t.Address)

			// Add failed address to results
			geocodeResults = append(geocodeResults, GeocodeResult{
				Address:   t.Address,
				Failed:    true,
				Timestamp: time.Now(),
			})
			failureCount++
			continue
		}

		// Attempt to get coordinates
		coords, err := GetCoordinates(t.Address)
		if err != nil {
			log.Printf("Failed to geocode address: %v", err)

			// Add failed address to results
			geocodeResults = append(geocodeResults, GeocodeResult{
				Address:   t.Address,
				Failed:    true,
				Timestamp: time.Now(),
			})
			failureCount++
			continue
		}

		if coords.Failed {
			log.Printf("No coordinates found for address: %s, %s %s",
				strings.TrimSpace(t.Address.StreetAddress),
				strings.TrimSpace(t.Address.PostalCode),
				strings.TrimSpace(t.Address.AddressLocality))

			// Add failed address to results
			geocodeResults = append(geocodeResults, GeocodeResult{
				Address:   t.Address,
				Failed:    true,
				Timestamp: time.Now(),
			})
			failureCount++
		} else {
			successCount++
			log.Printf("Processed address: %s, %s %s (%.6f, %.6f)",
				strings.TrimSpace(t.Address.StreetAddress),
				strings.TrimSpace(t.Address.PostalCode),
				strings.TrimSpace(t.Address.AddressLocality),
				coords.Lat,
				coords.Lon)

			// Add successful geocode result
			geocodeResults = append(geocodeResults, GeocodeResult{
				Address:   t.Address,
				Latitude:  coords.Lat,
				Longitude: coords.Lon,
				Failed:    false,
				Timestamp: time.Now(),
			})
		}
	}

	// Save geocoding results to cache
	if err := saveGeocodeResultsToCache(geocodeResults); err != nil {
		log.Printf("Warning: Failed to save geocoding cache: %v", err)
	}

	log.Printf("Preloading completed: %d successful, %d failed", successCount, failureCount)
	return nil
}
