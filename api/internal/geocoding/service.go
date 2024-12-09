package geocoding

import (
	"fmt"
	"time"
	"tournois-tt/api/internal/types"
	"tournois-tt/api/pkg/geocoding/address"
	"tournois-tt/api/pkg/geocoding/cache"
	"tournois-tt/api/pkg/geocoding/nominatim"
)

var client *types.NominatimClient

func init() {
	client = nominatim.NewClient()
}

// GetCoordinates returns the coordinates for a given address, using cache if available
func GetCoordinates(addr types.Address) (types.Location, error) {
	pkgAddr := address.AddressInput{
		StreetAddress:   addr.StreetAddress,
		PostalCode:      addr.PostalCode,
		AddressLocality: addr.AddressLocality,
		AddressCountry:  addr.AddressCountry,
	}

	if !address.IsValid(pkgAddr) {
		return types.Location{Failed: true, LastUpdated: time.Now()}, fmt.Errorf("invalid address")
	}

	// Generate all possible variants of the address
	variants := address.GenerateVariants(&pkgAddr)
	if len(variants) == 0 {
		return types.Location{Failed: true, LastUpdated: time.Now()}, fmt.Errorf("no valid address variants")
	}

	// Try to find any variant in cache first
	for _, variant := range variants {
		if loc, exists := cache.DefaultCache.Get(variant); exists {
			return loc, nil
		}
	}

	// If not in cache, try geocoding the most specific variant first
	loc, err := client.Geocode(variants[0])
	if err != nil || loc.Failed {
		// If the most specific variant fails, try the next one
		for _, variant := range variants[1:] {
			loc, err = client.Geocode(variant)
			if err == nil && !loc.Failed {
				break
			}
		}
	}

	// If we found a valid location, cache it and create aliases
	if err == nil && !loc.Failed {
		cache.DefaultCache.Set(variants[0], loc.Lat, loc.Lon, false, loc.Approximate)
		for _, v := range variants[1:] {
			cache.DefaultCache.AddAlias(v, variants[0])
		}

		// Save cache if needed
		if time.Since(cache.DefaultCache.LastSaveTime()) > 5*time.Minute {
			if err := cache.DefaultCache.SaveToFile(); err != nil {
				fmt.Printf("Error saving geocoding cache: %v\n", err)
			}
		}

		return loc, nil
	}

	// If all variants fail, mark as failed in cache
	failedLoc := types.Location{Failed: true, LastUpdated: time.Now()}
	cache.DefaultCache.Set(variants[0], 0, 0, true, false)
	for _, v := range variants[1:] {
		cache.DefaultCache.AddAlias(v, variants[0])
	}

	return failedLoc, fmt.Errorf("geocoding failed for all address variants")
}
