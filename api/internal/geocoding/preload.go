package geocoding

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"tournois-tt/api/pkg/fftt"
)

func PreloadTournaments() error {
	log.Printf("Starting tournament preloading for geocoding...")

	// Create query params for future tournaments
	queryParams := url.Values{}
	queryParams.Set("startDate[after]", time.Now().Format("2006-01-02T15:04:05"))
	queryParams.Set("itemsPerPage", "999999") // Get all tournaments
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

	log.Printf("Found %d tournaments to geocode", len(tournaments))

	// Process each unique address
	uniqueAddresses := make(map[string]struct{})
	for _, t := range tournaments {
		variants := generateAddressVariants(t.Address)
		for _, variant := range variants {
			uniqueAddresses[variant] = struct{}{}
		}
	}

	log.Printf("Found %d unique address variants to geocode", len(uniqueAddresses))

	// Check which addresses need geocoding
	var toGeocode []string
	for addr := range uniqueAddresses {
		if _, exists := cache.Get(addr); !exists {
			toGeocode = append(toGeocode, addr)
		}
	}

	log.Printf("%d addresses need geocoding", len(toGeocode))

	// Geocode missing addresses
	for i, addr := range toGeocode {
		if i > 0 && i%10 == 0 {
			log.Printf("Progress: geocoded %d/%d addresses", i, len(toGeocode))
		}

		coords, err := geocodeAddressWithRetry(addr)
		if err != nil {
			log.Printf("Warning: Failed to geocode address: %s", addr)
			cache.Set(addr, 0, 0, true, false)
			continue
		}

		cache.Set(addr, coords.Lat, coords.Lon, coords.Failed, coords.Approximate)

		// Save cache periodically
		if time.Since(cache.LastSave) > 5*time.Minute {
			if err := cache.SaveToFile(); err != nil {
				log.Printf("Error saving cache: %v", err)
			}
		}

		// Rate limiting
		time.Sleep(time.Second)
	}

	// Final cache save
	if err := cache.SaveToFile(); err != nil {
		return fmt.Errorf("failed to save cache after preloading: %v", err)
	}

	log.Printf("Successfully completed tournament preloading")
	return nil
}
