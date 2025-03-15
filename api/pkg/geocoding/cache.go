package geocoding

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// GeocodeCache provides a thread-safe cache for geocoding results
type GeocodeCache struct {
	sync.RWMutex
	items map[string]GeocodeResult
}

// NewGeocodeCache creates a new geocode cache instance
func NewGeocodeCache() *GeocodeCache {
	return &GeocodeCache{
		items: make(map[string]GeocodeResult),
	}
}

// Get retrieves a geocode result from the cache
func (c *GeocodeCache) Get(key string) (GeocodeResult, bool) {
	c.RLock()
	defer c.RUnlock()
	item, exists := c.items[key]
	return item, exists
}

// Set adds a geocode result to the cache
func (c *GeocodeCache) Set(key string, value GeocodeResult) {
	c.Lock()
	defer c.Unlock()
	c.items[key] = value
}

// GetAll returns a copy of all items in the cache
func (c *GeocodeCache) GetAll() map[string]GeocodeResult {
	c.RLock()
	defer c.RUnlock()

	// Create a copy to avoid concurrent access issues
	result := make(map[string]GeocodeResult, len(c.items))
	for k, v := range c.items {
		result[k] = v
	}

	return result
}

// SetAll replaces all items in the cache
func (c *GeocodeCache) SetAll(items map[string]GeocodeResult) {
	c.Lock()
	defer c.Unlock()

	c.items = make(map[string]GeocodeResult, len(items))
	for k, v := range items {
		c.items[k] = v
	}
}

// DefaultGeocodeCache is the global instance of the geocode cache
var DefaultGeocodeCache = NewGeocodeCache()

// getCacheDirectory returns the absolute path to the cache directory
func getCacheDirectory() string {
	// Get the executable's directory
	execDir, err := os.Getwd()
	if err != nil {
		return "cache"
	}

	return filepath.Join(execDir, "cache")
}

// SaveGeocodeResultsToCache saves geocoding results to a JSON file
func SaveGeocodeResultsToCache(results []GeocodeResult) error {
	// Get cache directory
	cacheDir := getCacheDirectory()
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %v", err)
	}

	// Prepare cache file path
	cacheFilePath := filepath.Join(cacheDir, "data.json")

	// Add results to the in-memory cache
	for _, result := range results {
		key := GenerateCacheKey(result.Address)
		DefaultGeocodeCache.Set(key, result)
	}

	// Get all cache entries for persistence
	allCacheEntries := DefaultGeocodeCache.GetAll()
	allResults := make([]GeocodeResult, 0, len(allCacheEntries))
	for _, result := range allCacheEntries {
		allResults = append(allResults, result)
	}

	// Marshal results to JSON
	data, err := json.MarshalIndent(allResults, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal geocoding results: %v", err)
	}

	// Write to file
	if err := os.WriteFile(cacheFilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write geocoding cache: %v", err)
	}

	return nil
}

// LoadGeocodeResultsFromCache loads existing geocoding results from JSON file
func LoadGeocodeResultsFromCache() (map[string]GeocodeResult, error) {
	// Use mutex to ensure thread safety
	DefaultGeocodeCache.RLock()
	defer DefaultGeocodeCache.RUnlock()

	cacheFilePath := filepath.Join(getCacheDirectory(), "data.json")

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
		key := GenerateCacheKey(result.Address)
		cacheMap[key] = result
	}

	// Update the in-memory cache
	DefaultGeocodeCache.SetAll(cacheMap)

	return cacheMap, nil
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
	key := GenerateCacheKey(addr)
	return DefaultGeocodeCache.Get(key)
}

// SetCachedGeocodeResult stores a geocoding result in the cache
func SetCachedGeocodeResult(result GeocodeResult) {
	key := GenerateCacheKey(result.Address)
	DefaultGeocodeCache.Set(key, result)
}

// Legacy function for backward compatibility
func LoadGeocodeCache() (map[string]GeocodeResult, error) {
	cacheFilePath := filepath.Join("cache", "data.json")

	// Check if cache file exists
	if _, err := os.Stat(cacheFilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("geocoding cache file not found")
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

	cacheMap := make(map[string]GeocodeResult)
	for _, result := range cachedResults {
		key := GenerateCacheKey(result.Address)
		cacheMap[key] = result
	}

	return cacheMap, nil
}
