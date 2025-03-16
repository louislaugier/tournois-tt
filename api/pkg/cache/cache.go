package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

	// Use multiple goroutines to process items
	numWorkers := 4 // Number of worker goroutines
	if len(items) < numWorkers {
		numWorkers = len(items)
	}

	// Create a channel for the items
	itemsChan := make(chan T, len(items))
	for _, item := range items {
		itemsChan <- item
	}
	close(itemsChan)

	// Create a channel for the results
	type keyValuePair struct {
		key   string
		value T
	}
	resultsChan := make(chan keyValuePair, len(items))

	// Create a wait group to wait for all workers
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	// Start workers
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for item := range itemsChan {
				key := keyFn(item)
				resultsChan <- keyValuePair{key, item}
			}
		}()
	}

	// Wait for all workers to finish and close the results channel
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	cacheMap := make(map[string]T)
	for result := range resultsChan {
		cacheMap[result.key] = result.value
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

	// Use multiple goroutines for marshaling large datasets
	if len(itemsList) > 100 {
		type marshalResult struct {
			data []byte
			err  error
		}

		// Marshal the entire collection directly since this is more efficient
		// than parallel marshaling for individual items with subsequent combine
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

	// For small datasets, just marshal directly
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

// DefaultTournamentCache is the global instance of the tournament cache
var DefaultTournamentCache *GenericCache[TournamentCache]

// CacheFilePath is the path to the cache file
var CacheFilePath string

// InitCache initializes the unified cache
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
	DefaultTournamentCache, err = LoadFromJSON(CacheFilePath, func(tournament TournamentCache) string {
		return GenerateTournamentCacheKey(tournament)
	})

	if err != nil {
		return fmt.Errorf("failed to load tournament cache: %v", err)
	}

	return nil
}

// EnsureCacheInitialized makes sure cache is initialized
func EnsureCacheInitialized() error {
	if DefaultTournamentCache == nil {
		return InitCache()
	}
	return nil
}

// SaveTournamentsToCache saves tournament data to a JSON file
func SaveTournamentsToCache(tournaments []TournamentCache) error {
	// Ensure cache is initialized
	if err := EnsureCacheInitialized(); err != nil {
		return err
	}

	// Add tournaments to the in-memory cache
	for _, tournament := range tournaments {
		key := GenerateTournamentCacheKey(tournament)
		DefaultTournamentCache.Set(key, tournament)
	}

	// Save to JSON file
	return SaveToJSON(DefaultTournamentCache, CacheFilePath)
}

// LoadTournaments loads existing tournament data from JSON file
func LoadTournaments() (map[string]TournamentCache, error) {
	// Ensure cache is initialized
	if err := EnsureCacheInitialized(); err != nil {
		return nil, err
	}

	// Return a copy of all items in the cache
	return DefaultTournamentCache.GetAll(), nil
}

// GenerateTournamentCacheKey creates a unique key for a tournament
func GenerateTournamentCacheKey(tournament TournamentCache) string {
	return fmt.Sprintf("%d", tournament.ID)
}

// GetCachedTournament retrieves a tournament from the cache
func GetCachedTournament(id int) (TournamentCache, bool) {
	// Ensure cache is initialized
	if err := EnsureCacheInitialized(); err != nil {
		return TournamentCache{}, false
	}

	key := fmt.Sprintf("%d", id)
	return DefaultTournamentCache.Get(key)
}

// SetCachedTournament stores a tournament in the cache
func SetCachedTournament(tournament TournamentCache) {
	// Ensure cache is initialized - ignore error as we're just setting
	if DefaultTournamentCache == nil {
		if err := InitCache(); err != nil {
			return
		}
	}

	key := GenerateTournamentCacheKey(tournament)
	DefaultTournamentCache.Set(key, tournament)
}

// GenerateAddressCacheKey creates a unique key for an address
func GenerateAddressCacheKey(addr Address) string {
	return fmt.Sprintf("%s|%s|%s",
		strings.TrimSpace(addr.StreetAddress),
		strings.TrimSpace(addr.PostalCode),
		strings.TrimSpace(addr.AddressLocality))
}
