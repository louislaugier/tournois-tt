package geocoding

import (
	"fmt"
	"log"
	"strings"
	"tournois-tt/api/internal/types"
	"tournois-tt/api/pkg/geocoding/address"
	"tournois-tt/api/pkg/geocoding/cache"
	"tournois-tt/api/pkg/geocoding/nominatim"
)

// GetCoordinates returns the coordinates for an address, using cache if available
func GetCoordinates(addr types.Address) (*types.Location, error) {
	// Convert to package address type
	pkgAddr := address.AddressInput{
		StreetAddress:   addr.StreetAddress,
		PostalCode:      addr.PostalCode,
		AddressLocality: addr.AddressLocality,
	}

	if !address.IsValid(pkgAddr) {
		return nil, fmt.Errorf("invalid address")
	}

	// Generate address variants
	variants := address.GenerateVariants(&pkgAddr)

	// Try each variant in cache first
	for _, variant := range variants {
		if loc, exists := cache.DefaultCache.Get(variant); exists && !loc.Failed {
			return &types.Location{
				Lat:         loc.Lat,
				Lon:         loc.Lon,
				Failed:      loc.Failed,
				Approximate: loc.Approximate,
			}, nil
		}
	}

	// If not in cache, try geocoding each variant until we find a success
	client := nominatim.NewClient()
	var lastError error
	for _, variant := range variants {
		// Skip if we already know this variant failed
		if loc, exists := cache.DefaultCache.Get(variant); exists && loc.Failed {
			continue
		}

		coords, err := client.Geocode(variant)
		if err != nil {
			lastError = err
			log.Printf("Failed to geocode variant %s: %v", variant, err)
			cache.DefaultCache.Set(variant, 0, 0, true, false, nil)
			continue
		}

		if !coords.Failed {
			// Success! Store with all variants and return
			cache.DefaultCache.Set(variant, coords.Lat, coords.Lon, false, coords.Approximate, variants)
			return &types.Location{
				Lat:         coords.Lat,
				Lon:         coords.Lon,
				Failed:      false,
				Approximate: coords.Approximate,
			}, nil
		}

		// Mark this variant as failed in cache
		cache.DefaultCache.Set(variant, 0, 0, true, false, nil)
	}

	if lastError != nil {
		return nil, fmt.Errorf("failed to geocode address: %v", lastError)
	}
	return nil, fmt.Errorf("no successful geocoding result for variants: %s", strings.Join(variants, " / "))
}
