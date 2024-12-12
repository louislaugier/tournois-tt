package geocoding

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
	"tournois-tt/api/pkg/fftt"
	"tournois-tt/api/pkg/geocoding/address"
	gcache "tournois-tt/api/pkg/geocoding/cache"
	"tournois-tt/api/pkg/geocoding/nominatim"
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
		Address address.AddressInput `json:"address"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tournaments); err != nil {
		return fmt.Errorf("failed to decode tournaments: %v", err)
	}

	log.Printf("Found %d tournaments to geocode", len(tournaments))

	// Log raw addresses for debugging
	for _, t := range tournaments {
		log.Printf("Raw address: street=%q, postal=%q, locality=%q, desc=%q",
			t.Address.StreetAddress,
			t.Address.PostalCode,
			t.Address.AddressLocality,
			t.Address.DisambiguatingDescription)
	}

	// Process each unique address
	uniqueAddresses := make(map[string]address.AddressInput)
	addressVariants := make(map[string][]string) // Track variants for each original address
	for _, t := range tournaments {
		if !address.IsValid(t.Address) {
			log.Printf("Skipping invalid address: %+v", t.Address)
			continue
		}

		variants := address.GenerateVariants(&t.Address)
		log.Printf("Generated %d variants for address: %s, %s %s",
			len(variants), t.Address.StreetAddress, t.Address.PostalCode, t.Address.AddressLocality)

		// Log all generated variants
		for _, variant := range variants {
			log.Printf("  Variant: %s", variant)
		}

		// Store original address as key for variants
		key := fmt.Sprintf("%s, %s %s",
			t.Address.StreetAddress,
			t.Address.PostalCode,
			t.Address.AddressLocality)
		addressVariants[key] = variants

		// Always add all variants to try
		for _, variant := range variants {
			uniqueAddresses[variant] = t.Address
		}
	}

	log.Printf("Found %d unique address variants to geocode", len(uniqueAddresses))

	// Check which addresses need geocoding
	var toGeocode []string
	for addr := range uniqueAddresses {
		if loc, exists := gcache.DefaultCache.Get(addr); exists && !loc.Failed {
			log.Printf("Skipping cached address: %s", addr)
			continue // Skip if we have a successful geocoding result
		}
		toGeocode = append(toGeocode, addr)
	}

	log.Printf("%d addresses need geocoding", len(toGeocode))
	log.Printf("Addresses to geocode:")
	for _, addr := range toGeocode {
		log.Printf("  %s", addr)
	}

	// Create Nominatim client
	client := nominatim.NewClient()

	// Geocode missing addresses
	successCount := 0
	failureCount := 0
	for i, addr := range toGeocode {
		log.Printf("Geocoding address (%d/%d): %s", i+1, len(toGeocode), addr)

		coords, err := client.Geocode(addr)
		if err != nil {
			log.Printf("Warning: Failed to geocode address: %s: %v", addr, err)
			failureCount++
			gcache.DefaultCache.Set(addr, 0, 0, true, false, nil, "")
			continue
		}

		if !coords.Failed {
			successCount++
			// Find the original address key
			originalAddr := uniqueAddresses[addr]
			key := fmt.Sprintf("%s, %s %s",
				originalAddr.StreetAddress,
				originalAddr.PostalCode,
				originalAddr.AddressLocality)

			// Store the successful result with all its variants
			variants := addressVariants[key]
			gcache.DefaultCache.Set(addr, coords.Lat, coords.Lon, false, coords.Approximate, variants, key)

			log.Printf("Successfully geocoded: %s (%.6f, %.6f) approximate: %v",
				addr, coords.Lat, coords.Lon, coords.Approximate)
		} else {
			failureCount++
			gcache.DefaultCache.Set(addr, 0, 0, true, false, nil, "")
			log.Printf("Failed to find coordinates for: %s", addr)
		}

		// Save cache periodically
		if time.Since(gcache.DefaultCache.LastSaveTime()) > 5*time.Minute {
			if err := gcache.DefaultCache.SaveToFile(); err != nil {
				log.Printf("Error saving cache: %v", err)
			}
		}

		time.Sleep(time.Second) // Rate limiting
	}

	// Final cache save
	if err := gcache.DefaultCache.SaveToFile(); err != nil {
		return fmt.Errorf("failed to save cache after preloading: %v", err)
	}

	log.Printf("Geocoding completed: %d successful, %d failed", successCount, failureCount)
	log.Printf("Successfully completed tournament preloading")
	return nil
}
