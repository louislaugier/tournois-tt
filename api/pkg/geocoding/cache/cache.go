package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"tournois-tt/api/internal/types"
)

var DefaultCache types.Cache

func init() {
	var err error
	DefaultCache, err = LoadFromFile()
	if err != nil {
		fmt.Printf("Error loading geocoding cache: %v\n", err)
		DefaultCache = NewCache()
	}
}

// NewCache creates a new empty cache
func NewCache() types.Cache {
	return &types.RuntimeCache{
		GeocodingCache: &types.GeocodingCache{
			Locations: make(map[string]types.Location),
			LastSave:  time.Time{},
		},
	}
}

// LoadFromFile loads the cache from disk
func LoadFromFile() (types.Cache, error) {
	cache := NewCache().(*types.RuntimeCache)

	// Ensure cache directory exists
	cacheDir := "cache"
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %v", err)
	}

	cacheFile := filepath.Join(cacheDir, "geocoding_cache.json")
	absPath, err := filepath.Abs(cacheFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %v", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cache, nil
		}
		return nil, fmt.Errorf("failed to read cache file: %v", err)
	}

	if err := json.Unmarshal(data, cache.GeocodingCache); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache: %v", err)
	}

	return cache, nil
}
