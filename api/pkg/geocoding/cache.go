package geocoding

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// Cache represents a simple in-memory cache for geocoding
type Cache struct {
	sync.RWMutex
	items map[string]interface{}
}

// NewCache creates a new cache instance
func NewCache() *Cache {
	return &Cache{
		items: make(map[string]interface{}),
	}
}

// Get retrieves an item from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.RLock()
	defer c.RUnlock()
	item, exists := c.items[key]
	return item, exists
}

// Set adds an item to the cache
func (c *Cache) Set(key string, value interface{}) {
	c.Lock()
	defer c.Unlock()
	c.items[key] = value
}

// DefaultCache is the global cache instance
var DefaultCache = NewCache()

func init() {
	fmt.Println("Geocoding cache initialized")
}

// saveGeocodeResultsToCache saves geocoding results to a JSON file
func saveGeocodeResultsToCache(results []GeocodeResult) error {
	// Ensure cache directory exists
	cacheDir := "cache"
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %v", err)
	}

	// Prepare cache file path
	cacheFilePath := filepath.Join(cacheDir, "geocoding_cache.json")

	// Marshal results to JSON
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal geocoding results: %v", err)
	}

	// Write to file
	if err := os.WriteFile(cacheFilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write geocoding cache: %v", err)
	}

	log.Printf("Saved %d geocoding results to %s", len(results), cacheFilePath)
	return nil
}
