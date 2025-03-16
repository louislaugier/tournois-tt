package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// GenericCache is a thread-safe in-memory cache with JSON persistence capabilities
type GenericCache[T any] struct {
	sync.RWMutex
	items map[string]T
}

// NewGenericCache creates a new cache instance
func NewGenericCache[T any]() *GenericCache[T] {
	return &GenericCache[T]{
		items: make(map[string]T),
	}
}

// Get retrieves an item from the cache
func (c *GenericCache[T]) Get(key string) (T, bool) {
	c.RLock()
	defer c.RUnlock()
	item, exists := c.items[key]
	return item, exists
}

// Set adds an item to the cache
func (c *GenericCache[T]) Set(key string, value T) {
	c.Lock()
	defer c.Unlock()
	c.items[key] = value
}

// GetAll returns a copy of all items in the cache
func (c *GenericCache[T]) GetAll() map[string]T {
	c.RLock()
	defer c.RUnlock()

	// Create a copy to avoid concurrent access issues
	result := make(map[string]T, len(c.items))
	for k, v := range c.items {
		result[k] = v
	}

	return result
}

// SetAll replaces all items in the cache
func (c *GenericCache[T]) SetAll(items map[string]T) {
	c.Lock()
	defer c.Unlock()

	c.items = make(map[string]T, len(items))
	for k, v := range items {
		c.items[k] = v
	}
}

// Delete removes an item from the cache
func (c *GenericCache[T]) Delete(key string) {
	c.Lock()
	defer c.Unlock()
	delete(c.items, key)
}

// Size returns the number of items in the cache
func (c *GenericCache[T]) Size() int {
	c.RLock()
	defer c.RUnlock()
	return len(c.items)
}

// getCacheDirectory returns the absolute path to the cache directory
func GetCacheDirectory(subDir string) string {
	// Get the executable's directory
	execDir, err := os.Getwd()
	if err != nil {
		return filepath.Join("cache", subDir)
	}

	return filepath.Join(execDir, "cache", subDir)
}

// LoadFromJSON loads cache entries from a JSON file
func LoadFromJSON[T any](filePath string, keyFn func(T) string) (*GenericCache[T], error) {
	cache := NewGenericCache[T]()

	// Check if cache file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return cache, nil
	}

	// Read cache file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache file: %v", err)
	}

	var items []T
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("failed to parse cache file: %v", err)
	}

	// Convert to map using provided key function
	cacheMap := make(map[string]T)
	for _, item := range items {
		key := keyFn(item)
		cacheMap[key] = item
	}

	// Update the cache
	cache.SetAll(cacheMap)

	return cache, nil
}

// SaveToJSON saves cache entries to a JSON file
func SaveToJSON[T any](cache *GenericCache[T], filePath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %v", err)
	}

	// Get all cache entries
	allItems := cache.GetAll()
	itemsList := make([]T, 0, len(allItems))
	for _, item := range allItems {
		itemsList = append(itemsList, item)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(itemsList, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache items: %v", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %v", err)
	}

	return nil
}
