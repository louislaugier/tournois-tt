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

// getCacheDirectory returns the absolute path to the cache directory
func getCacheDirectory() string {
	// Get the executable's directory
	execDir, err := os.Getwd()
	if err != nil {
		log.Printf("Warning: Failed to get working directory, falling back to relative path: %v", err)
		return "cache"
	}

	return filepath.Join(execDir, "cache")
}

// saveGeocodeResultsToCache saves geocoding results to a JSON file
func saveGeocodeResultsToCache(results []GeocodeResult) error {
	// Get cache directory
	cacheDir := getCacheDirectory()
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

// loadGeocodeResultsFromCache loads existing geocoding results from JSON file
func loadGeocodeResultsFromCache() (map[string]GeocodeResult, error) {
	cacheFilePath := filepath.Join(getCacheDirectory(), "geocoding_cache.json")

	// Check if cache file exists
	if _, err := os.Stat(cacheFilePath); os.IsNotExist(err) {
		return make(map[string]GeocodeResult), nil
	}

	// Read cache file
	data, err := os.ReadFile(cacheFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read geocoding cache: %v", err)
	}

	var cachedResults []GeocodeResult
	if err := json.Unmarshal(data, &cachedResults); err != nil {
		return nil, fmt.Errorf("failed to parse geocoding cache: %v", err)
	}

	// Convert to map for faster lookup
	cacheMap := make(map[string]GeocodeResult)
	for _, result := range cachedResults {
		key := generateCacheKey(result.Address)
		cacheMap[key] = result
	}

	log.Printf("Loaded %d geocoding results from cache at %s", len(cacheMap), cacheFilePath)
	return cacheMap, nil
}
