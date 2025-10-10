package cache

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
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

// DefaultGeocodeCache is the global instance of the geocode cache
var DefaultGeocodeCache *GenericCache[GeocodeResult]

// CacheFilePath is the path to the cache file
var CacheFilePath string

// GeocodeCache filepath
var GeocodeCacheFilePath string

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
	GeocodeCacheFilePath = filepath.Join(cacheDir, "geocode.json")

	// Load cache from JSON file
	DefaultTournamentCache, err = LoadFromJSON(CacheFilePath, func(tournament TournamentCache) string {
		return GenerateTournamentCacheKey(tournament)
	})

	if err != nil {
		return fmt.Errorf("failed to load tournament cache: %v", err)
	}

	// Initialize geocode cache
	DefaultGeocodeCache = NewGenericCache[GeocodeResult]()

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
	if err := SaveToJSON(DefaultTournamentCache, CacheFilePath); err != nil {
		return err
	}

	// Update sitemap automatically after saving tournaments
	go updateSitemap()

	return nil
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

// GenerateGeocodeCacheKey creates a unique key for an address
func GenerateGeocodeCacheKey(addr Address) string {
	return fmt.Sprintf("%s|%s|%s",
		strings.TrimSpace(addr.StreetAddress),
		strings.TrimSpace(addr.PostalCode),
		strings.TrimSpace(addr.AddressLocality))
}

// SaveGeocodeResultsToCache saves geocoding results to a JSON file
func SaveGeocodeResultsToCache(results []GeocodeResult) error {
	// Ensure cache is initialized
	if err := EnsureCacheInitialized(); err != nil {
		return err
	}

	// Add results to the in-memory cache
	for _, result := range results {
		key := GenerateGeocodeCacheKey(result.Address)
		DefaultGeocodeCache.Set(key, result)
	}

	// Save to JSON file
	return SaveToJSON(DefaultGeocodeCache, GeocodeCacheFilePath)
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

// GetCachedGeocodeResult retrieves a geocoding result from the cache
func GetCachedGeocodeResult(addr Address) (GeocodeResult, bool) {
	// Get cached tournaments
	tournaments, err := LoadTournaments()
	if err != nil {
		return GeocodeResult{}, false
	}

	// Search for matching address in cached tournaments
	key := GenerateAddressCacheKey(addr)

	// Look for a tournament with matching address
	for _, tourney := range tournaments {
		if GenerateAddressCacheKey(tourney.Address) == key {
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

// updateSitemap updates the sitemap by calling npm scripts directly
func updateSitemap() {
	// Get the project root directory
	execDir, err := os.Getwd()
	if err != nil {
		log.Printf("Warning: Could not get working directory for sitemap update: %v", err)
		return
	}

	// Find the project root (look for api directory)
	var projectRoot string
	if idx := strings.LastIndex(execDir, "api"); idx != -1 {
		projectRoot = execDir[:idx-1] // Remove "api" and the trailing slash
	} else {
		projectRoot = execDir
	}
	

	// Path to the frontend directory (different paths for dev vs prod)
	var frontendDir string
	var isDockerEnv bool

	// Check if we're in Docker environment (agnostic dev/prod)
	// Try different possible Docker frontend paths
	dockerPaths := []string{
		"/app/frontend",           // Production Docker
		"/tournois-tt/frontend",   // Development Docker
		filepath.Join(projectRoot, "frontend"), // Try project root frontend too
	}
	
	var foundDockerPath string
	for _, dockerPath := range dockerPaths {
		if _, err := os.Stat(dockerPath); err == nil {
			foundDockerPath = dockerPath
			break
		}
	}
	
	if foundDockerPath != "" {
		// We're in Docker
		frontendDir = foundDockerPath
		isDockerEnv = true
		log.Printf("Detected Docker environment, using frontend directory: %s", frontendDir)
		
		// Only generate sitemap/RSS in production Docker (single container)
		// In development Docker, containers are separate so skip generation
		if frontendDir == "/tournois-tt/frontend" {
			log.Printf("Development Docker detected - skipping sitemap/RSS generation (separate containers)")
			return // Exit early, no generation in dev Docker
		}
	} else {
		// We're in local development, frontend files are in frontend directory
		frontendDir = filepath.Join(projectRoot, "frontend")
		isDockerEnv = false
		log.Printf("Detected local development environment, using frontend directory: %s", frontendDir)
	}

	// Check if frontend directory exists
	if _, err := os.Stat(frontendDir); os.IsNotExist(err) {
		log.Printf("Warning: Frontend directory not found: %s", frontendDir)
		return
	}

	if isDockerEnv {
		// In Docker, determine output directory based on environment
		var outputDir string
		if frontendDir == "/app/frontend" {
			// Production Docker - output to nginx html directory
			outputDir = "/usr/share/nginx/html"
		} else {
			// Development Docker - output to frontend public directory
			outputDir = filepath.Join(frontendDir, "public")
		}
		
		// Execute npm scripts from Docker frontend directory
		commands := []string{
			"generate-sitemap",
			"generate-rss",
		}
		
		for _, cmdName := range commands {
			cmd := exec.Command("npm", "run", cmdName)
			cmd.Dir = frontendDir
			cmd.Env = append(os.Environ(), "OUTPUT_DIR="+outputDir)
			
			if err := cmd.Run(); err != nil {
				log.Printf("Warning: Failed to run npm run %s: %v", cmdName, err)
			} else {
				log.Printf("Successfully ran npm run %s", cmdName)
			}
		}
	} else {
		// In development, run scripts normally
		commands := []string{
			"generate-sitemap",
			"generate-rss",
		}

		for _, cmdName := range commands {
			cmd := exec.Command("npm", "run", cmdName)
			cmd.Dir = frontendDir

			if err := cmd.Run(); err != nil {
				log.Printf("Warning: Failed to run npm run %s: %v", cmdName, err)
			} else {
				log.Printf("Successfully ran npm run %s", cmdName)
			}
		}
	}
}
