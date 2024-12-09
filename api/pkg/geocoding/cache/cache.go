package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
			Locations:   make(map[string]types.Location),
			AliasGroups: make([][]string, 0),
		},
		AliasMap: make(map[string]string),
	}
}

// LoadFromFile loads the cache from disk
func LoadFromFile() (types.Cache, error) {
	cache := NewCache().(*types.RuntimeCache)

	cacheFile := filepath.Join("cache", "geocoding_cache.json")
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return cache, nil
		}
		return nil, fmt.Errorf("failed to read cache file: %v", err)
	}

	if err := json.Unmarshal(data, cache.GeocodingCache); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache: %v", err)
	}

	// Rebuild alias map
	cache.RebuildAliasMap()

	return cache, nil
}
