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

// buildFullAddress creates a complete address string including all components
func buildFullAddress(addr types.Address) string {
	var parts []string

	// Always start with disambiguating description (venue name) if available
	if addr.DisambiguatingDescription != "" {
		parts = append(parts, strings.TrimSpace(addr.DisambiguatingDescription))
	}

	// Then add street address
	if addr.StreetAddress != "" {
		parts = append(parts, strings.TrimSpace(addr.StreetAddress))
	}

	// Always include postal code and locality
	if addr.PostalCode != "" && addr.AddressLocality != "" {
		parts = append(parts, strings.TrimSpace(addr.PostalCode)+" "+strings.TrimSpace(addr.AddressLocality))
	}

	// Join all parts with commas and add france suffix
	fullAddr := strings.Join(parts, ", ")
	if fullAddr != "" {
		fullAddr += ", france"
	}

	return strings.ToLower(fullAddr)
}

// GetCoordinates returns the coordinates for an address, using cache if available
func GetCoordinates(addr types.Address) (*types.Location, error) {
	// Log input address details
	log.Printf("Geocoding request for address:")
	log.Printf("  Street: %q", addr.StreetAddress)
	log.Printf("  Postal: %q", addr.PostalCode)
	log.Printf("  Locality: %q", addr.AddressLocality)
	log.Printf("  Description: %q", addr.DisambiguatingDescription)

	// Build full address for cache key
	fullAddress := buildFullAddress(addr)
	log.Printf("Full address (cache key): %q", fullAddress)

	// Convert to package address type
	pkgAddr := address.AddressInput{
		StreetAddress:             addr.StreetAddress,
		PostalCode:                addr.PostalCode,
		AddressLocality:           addr.AddressLocality,
		DisambiguatingDescription: addr.DisambiguatingDescription,
	}

	if !address.IsValid(pkgAddr) {
		log.Printf("Invalid address: missing postal code or locality")
		return nil, fmt.Errorf("invalid address: missing postal code or locality")
	}

	// Generate address variants
	variants := address.GenerateVariants(&pkgAddr)
	log.Printf("Generated %d address variants:", len(variants))
	for i, variant := range variants {
		log.Printf("  %d. %q", i+1, variant)
	}

	// First, check if we have the full address in cache
	if loc, exists := cache.DefaultCache.Get(fullAddress); exists {
		if !loc.Failed {
			log.Printf("Found successful cache entry for full address: %q", fullAddress)
			return &types.Location{
				Lat:         loc.Lat,
				Lon:         loc.Lon,
				Failed:      loc.Failed,
				Approximate: loc.Approximate,
			}, nil
		}
		log.Printf("Found failed cache entry for full address: %q", fullAddress)
	}

	// Then check variants in cache
	for _, variant := range variants {
		if loc, exists := cache.DefaultCache.Get(variant); exists && !loc.Failed {
			log.Printf("Found successful cache entry for variant: %q", variant)
			// Store the result with the full address as key
			cache.DefaultCache.Set(fullAddress, loc.Lat, loc.Lon, false, loc.Approximate, variants, "")
			return &types.Location{
				Lat:         loc.Lat,
				Lon:         loc.Lon,
				Failed:      false,
				Approximate: loc.Approximate,
			}, nil
		}
	}

	// If not in cache, try geocoding each address form
	log.Printf("No successful cache entries found, attempting geocoding...")
	client := nominatim.NewClient()
	var lastError error
	var failedVariants []string

	// First try the full address
	log.Printf("Attempting to geocode full address: %q", fullAddress)
	coords, err := client.Geocode(fullAddress)
	if err == nil && !coords.Failed {
		log.Printf("Successfully geocoded full address %q: (%.6f, %.6f) approximate: %v",
			fullAddress, coords.Lat, coords.Lon, coords.Approximate)

		// Store with full address as key
		cache.DefaultCache.Set(fullAddress, coords.Lat, coords.Lon, false, coords.Approximate, variants, "")
		return &types.Location{
			Lat:         coords.Lat,
			Lon:         coords.Lon,
			Failed:      false,
			Approximate: coords.Approximate,
		}, nil
	}

	// If full address fails, try each variant
	for _, variant := range variants {
		log.Printf("Attempting to geocode variant: %q", variant)
		coords, err := client.Geocode(variant)
		if err != nil {
			lastError = err
			log.Printf("Error geocoding variant %q: %v", variant, err)
			failedVariants = append(failedVariants, variant)
			continue
		}

		if !coords.Failed {
			log.Printf("Successfully geocoded variant %q: (%.6f, %.6f) approximate: %v",
				variant, coords.Lat, coords.Lon, coords.Approximate)

			// Store with full address as key
			cache.DefaultCache.Set(fullAddress, coords.Lat, coords.Lon, false, coords.Approximate, variants, "")
			return &types.Location{
				Lat:         coords.Lat,
				Lon:         coords.Lon,
				Failed:      false,
				Approximate: coords.Approximate,
			}, nil
		}

		log.Printf("No results found for variant: %q", variant)
		failedVariants = append(failedVariants, variant)
	}

	// If all attempts fail, store the failure with the full address
	cache.DefaultCache.Set(fullAddress, 0, 0, true, false, variants, "")

	// Log detailed failure information
	log.Printf("Failed to geocode address after trying full address and %d variants", len(variants))
	log.Printf("Failed variants: %s", strings.Join(failedVariants, ", "))

	if lastError != nil {
		return nil, fmt.Errorf("failed to geocode address: %v (tried full address and %d variants)", lastError, len(variants))
	}
	return nil, fmt.Errorf("no successful geocoding result for full address or any of %d variants: %s",
		len(variants), strings.Join(append([]string{fullAddress}, variants...), " / "))
}
