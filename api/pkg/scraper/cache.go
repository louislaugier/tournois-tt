package scraper

import (
	"fmt"
	"path/filepath"
	"time"
	"tournois-tt/api/pkg/cache"
)

// ScraperData represents the data structure to be cached for scraper results
type ScraperData struct {
	URL        string    `json:"url"`
	Source     string    `json:"source"`
	Data       any       `json:"data"`
	Timestamp  time.Time `json:"timestamp"`
	Expiration time.Time `json:"expiration,omitempty"`
}

// Cache is the main cache instance for scraper data
var Cache *cache.GenericCache[ScraperData]

// CacheFilePath is the path to the cache file
var CacheFilePath string

// InitCache initializes the scraper cache
func InitCache() error {
	// Set up the cache file path
	cacheDir := cache.GetCacheDirectory("scraper")
	CacheFilePath = filepath.Join(cacheDir, "data.json")

	// Load cache from JSON file
	var err error
	Cache, err = cache.LoadFromJSON(CacheFilePath, func(data ScraperData) string {
		return data.URL
	})

	if err != nil {
		return fmt.Errorf("failed to load scraper cache: %v", err)
	}

	return nil
}

// GetCachedData retrieves data from the cache based on URL
func GetCachedData(url, source string) (any, bool) {
	if Cache == nil {
		if err := InitCache(); err != nil {
			return nil, false
		}
	}

	data, exists := Cache.Get(url)
	if !exists {
		return nil, false
	}

	// If the source doesn't match, return not found
	if data.Source != source {
		return nil, false
	}

	// Check if the cache entry has expired
	if !data.Expiration.IsZero() && time.Now().After(data.Expiration) {
		return nil, false
	}

	return data.Data, true
}

// SetCachedData stores data in the cache with an optional expiration time
func SetCachedData(url, source string, data any, expiration time.Duration) error {
	if Cache == nil {
		if err := InitCache(); err != nil {
			return err
		}
	}

	var expirationTime time.Time
	if expiration > 0 {
		expirationTime = time.Now().Add(expiration)
	}

	cacheEntry := ScraperData{
		URL:        url,
		Source:     source,
		Data:       data,
		Timestamp:  time.Now(),
		Expiration: expirationTime,
	}

	Cache.Set(url, cacheEntry)

	// Persist the cache to disk
	return cache.SaveToJSON(Cache, CacheFilePath)
}

// SaveCache persists the cache to disk
func SaveCache() error {
	if Cache == nil {
		return fmt.Errorf("cache not initialized")
	}

	return cache.SaveToJSON(Cache, CacheFilePath)
}

// ClearCache clears all entries from the cache
func ClearCache() error {
	if Cache == nil {
		if err := InitCache(); err != nil {
			return err
		}
	}

	Cache.SetAll(make(map[string]ScraperData))
	return SaveCache()
}

// PruneExpiredEntries removes all expired entries from the cache
func PruneExpiredEntries() error {
	if Cache == nil {
		if err := InitCache(); err != nil {
			return err
		}
	}

	now := time.Now()
	allEntries := Cache.GetAll()
	changed := false

	for url, entry := range allEntries {
		if !entry.Expiration.IsZero() && now.After(entry.Expiration) {
			Cache.Delete(url)
			changed = true
		}
	}

	if changed {
		return SaveCache()
	}

	return nil
}
