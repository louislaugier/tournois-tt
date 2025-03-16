package geocoding

import (
	"fmt"
	"strings"
	"tournois-tt/api/pkg/cache"
)

// DefaultGeocodeCache is the global instance of the geocode cache
var DefaultGeocodeCache *cache.GenericCache[GeocodeResult]

// CacheFilePath is the path to the geocoding cache file
var CacheFilePath string

// InitCache initializes the geocoding cache
func InitCache() error {
	return nil
}

// EnsureCacheInitialized makes sure cache is initialized
func EnsureCacheInitialized() error {
	return nil
}

// SaveGeocodeResultsToCache saves geocoding results to a JSON file
func SaveGeocodeResultsToCache(results []GeocodeResult) error {
	// Ensure cache is initialized
	if err := EnsureCacheInitialized(); err != nil {
		return err
	}

	// Add results to the in-memory cache
	for _, result := range results {
		key := GenerateCacheKey(result.Address)
		DefaultGeocodeCache.Set(key, result)
	}

	// Save to JSON file
	return cache.SaveToJSON(DefaultGeocodeCache, CacheFilePath)
}

// LoadGeocodeResultsFromCache loads existing geocoding results from JSON file
func LoadGeocodeResultsFromCache() (map[string]GeocodeResult, error) {
	// Ensure cache is initialized
	if err := EnsureCacheInitialized(); err != nil {
		return nil, err
	}

	// Return a copy of all items in the cache
	return DefaultGeocodeCache.GetAll(), nil
}

// GenerateCacheKey creates a unique key for an address
func GenerateCacheKey(addr Address) string {
	return fmt.Sprintf("%s|%s|%s",
		strings.TrimSpace(addr.StreetAddress),
		strings.TrimSpace(addr.PostalCode),
		strings.TrimSpace(addr.AddressLocality))
}

// GetCachedGeocodeResult retrieves a geocoding result from the cache
func GetCachedGeocodeResult(addr Address) (GeocodeResult, bool) {
	// Convert Address to cache.Address
	cacheAddr := cache.Address{
		StreetAddress:             addr.StreetAddress,
		PostalCode:                addr.PostalCode,
		AddressLocality:           addr.AddressLocality,
		DisambiguatingDescription: addr.DisambiguatingDescription,
		Latitude:                  addr.Latitude,
		Longitude:                 addr.Longitude,
		Failed:                    addr.Failed,
	}

	// Get cached tournaments
	tournaments, err := cache.LoadTournaments()
	if err != nil {
		return GeocodeResult{}, false
	}

	// Search for matching address in cached tournaments
	key := cache.GenerateAddressCacheKey(cacheAddr)

	// Look for a tournament with matching address
	for _, tourney := range tournaments {
		if cache.GenerateAddressCacheKey(tourney.Address) == key {
			// Found a match - extract geocode data
			result := GeocodeResult{
				Address: Address{
					StreetAddress:             tourney.Address.StreetAddress,
					PostalCode:                tourney.Address.PostalCode,
					AddressLocality:           tourney.Address.AddressLocality,
					DisambiguatingDescription: tourney.Address.DisambiguatingDescription,
					Latitude:                  tourney.Address.Latitude,
					Longitude:                 tourney.Address.Longitude,
					Failed:                    tourney.Address.Failed,
				},
				Latitude:  tourney.Address.Latitude,
				Longitude: tourney.Address.Longitude,
				Failed:    tourney.Address.Failed,
				Timestamp: tourney.Timestamp,
			}
			return result, true
		}
	}

	return GeocodeResult{}, false
}

// SetCachedGeocodeResult stores a geocoding result in the cache
func SetCachedGeocodeResult(result GeocodeResult) {
	// We won't implement this separately anymore.
	// Geocode results will be stored as part of tournament data when saving tournaments.
	// This is now just a stub for backward compatibility.
}

// Legacy function for backward compatibility
func LoadGeocodeCache() (map[string]GeocodeResult, error) {
	return LoadGeocodeResultsFromCache()
}
