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

	// Try each variant in cache
	for _, variant := range variants {
		if loc, exists := cache.DefaultCache.Get(variant); exists {
			if !loc.Failed {
				return &types.Location{
					Lat:         loc.Lat,
					Lon:         loc.Lon,
					Failed:      loc.Failed,
					Approximate: loc.Approximate,
				}, nil
			}
		}
	}

	// If not in cache, geocode the first non-failed variant
	client := nominatim.NewClient()
	for _, variant := range variants {
		if loc, exists := cache.DefaultCache.Get(variant); exists && loc.Failed {
			continue
		}

		coords, err := client.Geocode(variant)
		if err != nil {
			log.Printf("Failed to geocode variant %s: %v", variant, err)
			cache.DefaultCache.Set(variant, 0, 0, true, false)
			continue
		}

		if !coords.Failed {
			// Add aliases for all variants
			for _, v := range variants {
				if v != variant {
					cache.DefaultCache.AddAlias(v, variant)
				}
			}

			return &types.Location{
				Lat:         coords.Lat,
				Lon:         coords.Lon,
				Failed:      coords.Failed,
				Approximate: coords.Approximate,
			}, nil
		}

		cache.DefaultCache.Set(variant, 0, 0, true, false)
	}

	return nil, fmt.Errorf("failed to geocode address: %s", strings.Join(variants, " / "))
}
