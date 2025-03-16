package geocoding

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"tournois-tt/api/pkg/cache"
)

// DefaultGeocodeCache is the global instance of the geocode cache
var DefaultGeocodeCache *cache.GenericCache[GeocodeResult]

// CacheFilePath is the path to the geocoding cache file
var CacheFilePath string

// InitCache initializes the geocoding cache
func InitCache() error {
	// Set up the cache file path - match the existing structure where data.json is in api/cache
	execDir, err := os.Getwd()
	var cacheDir string
	if err != nil {
		// Fallback to relative path if we can't get working directory
		cacheDir = filepath.Join("api", "cache")
	} else {
		// Find the api directory in the path
		if idx := strings.LastIndex(execDir, "api"); idx != -1 {
			// If we're already in the api directory or a subdirectory
			cacheDir = filepath.Join(execDir[:idx+3], "cache") // 3 = len("api")
		} else {
			// If we're in the root directory
			cacheDir = filepath.Join(execDir, "api", "cache")
		}
	}

	CacheFilePath = filepath.Join(cacheDir, "data.json")

	// Load cache from JSON file
	DefaultGeocodeCache, err = cache.LoadFromJSON(CacheFilePath, func(result GeocodeResult) string {
		return GenerateCacheKey(result.Address)
	})

	if err != nil {
		return fmt.Errorf("failed to load geocoding cache: %v", err)
	}

	return nil
}

// EnsureCacheInitialized makes sure cache is initialized
func EnsureCacheInitialized() error {
	if DefaultGeocodeCache == nil {
		return InitCache()
	}
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
	// Ensure cache is initialized
	if err := EnsureCacheInitialized(); err != nil {
		return GeocodeResult{}, false
	}

	key := GenerateCacheKey(addr)
	return DefaultGeocodeCache.Get(key)
}

// SetCachedGeocodeResult stores a geocoding result in the cache
func SetCachedGeocodeResult(result GeocodeResult) {
	// Ensure cache is initialized - ignore error as we're just setting
	if DefaultGeocodeCache == nil {
		if err := InitCache(); err != nil {
			return
		}
	}

	key := GenerateCacheKey(result.Address)
	DefaultGeocodeCache.Set(key, result)
}

// Legacy function for backward compatibility
func LoadGeocodeCache() (map[string]GeocodeResult, error) {
	// This is a legacy function, but we'll use the new implementation internally
	return LoadGeocodeResultsFromCache()
}
