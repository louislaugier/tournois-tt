package instagram

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// PostedRecord represents a tournament that has been posted
type PostedRecord struct {
	TournamentID    int       `json:"tournament_id"`
	TournamentName  string    `json:"tournament_name"`
	PostedAt        time.Time `json:"posted_at"`
	InstagramFeed   bool      `json:"instagram_feed"`
	InstagramStory  bool      `json:"instagram_story"`
	Threads         bool      `json:"threads"`
	InstagramPostID string    `json:"instagram_post_id,omitempty"`
	InstagramStoryID string   `json:"instagram_story_id,omitempty"`
	ThreadsPostID   string    `json:"threads_post_id,omitempty"`
}

// PostedCache manages the cache of posted tournaments
type PostedCache struct {
	mu      sync.RWMutex
	records map[int]*PostedRecord // tournament_id -> record
	file    string
}

var (
	globalCache *PostedCache
	cacheOnce   sync.Once
)

// GetPostedCache returns the global posted cache singleton
func GetPostedCache() *PostedCache {
	cacheOnce.Do(func() {
		cache, err := NewPostedCache()
		if err != nil {
			log.Printf("⚠️  Warning: Failed to initialize posted cache: %v", err)
			// Create empty cache as fallback
			cache = &PostedCache{
				records: make(map[int]*PostedRecord),
				file:    "./cache/posted_instagram.json",
			}
		}
		globalCache = cache
	})
	return globalCache
}

// NewPostedCache creates a new posted cache
func NewPostedCache() (*PostedCache, error) {
	cacheFile := "./cache/posted_instagram.json"
	
	// Ensure cache directory exists
	cacheDir := filepath.Dir(cacheFile)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	cache := &PostedCache{
		records: make(map[int]*PostedRecord),
		file:    cacheFile,
	}

	// Load existing cache
	if err := cache.load(); err != nil {
		log.Printf("⚠️  Warning: Failed to load cache (will start fresh): %v", err)
	}

	return cache, nil
}

// IsPosted checks if a tournament has been posted
func (c *PostedCache) IsPosted(tournamentID int) (bool, *PostedRecord) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	record, exists := c.records[tournamentID]
	return exists, record
}

// MarkPosted records that a tournament has been posted
func (c *PostedCache) MarkPosted(record *PostedRecord) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.records[record.TournamentID] = record

	// Save to disk immediately
	if err := c.save(); err != nil {
		return fmt.Errorf("failed to save cache: %w", err)
	}

	log.Printf("✅ Cached tournament %d as posted", record.TournamentID)
	return nil
}

// Remove deletes a tournament from the cache
func (c *PostedCache) Remove(tournamentID int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.records, tournamentID)

	// Save to disk immediately
	if err := c.save(); err != nil {
		return fmt.Errorf("failed to save cache: %w", err)
	}

	log.Printf("🗑️  Removed tournament %d from cache", tournamentID)
	return nil
}

// GetAllRecords returns all posted records
func (c *PostedCache) GetAllRecords() []*PostedRecord {
	c.mu.RLock()
	defer c.mu.RUnlock()

	records := make([]*PostedRecord, 0, len(c.records))
	for _, record := range c.records {
		records = append(records, record)
	}
	return records
}

// load reads the cache from disk
func (c *PostedCache) load() error {
	data, err := os.ReadFile(c.file)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet, start with empty cache
			return nil
		}
		return fmt.Errorf("failed to read cache file: %w", err)
	}

	var records []*PostedRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return fmt.Errorf("failed to parse cache: %w", err)
	}

	// Build index
	for _, record := range records {
		c.records[record.TournamentID] = record
	}

	log.Printf("📦 Loaded %d posted tournaments from cache", len(c.records))
	return nil
}

// save writes the cache to disk
func (c *PostedCache) save() error {
	// Convert map to slice
	records := make([]*PostedRecord, 0, len(c.records))
	for _, record := range c.records {
		records = append(records, record)
	}

	// Marshal to JSON with nice formatting
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	// Write to file
	if err := os.WriteFile(c.file, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// ValidateWithAPI checks if cached posts still exist on Instagram/Threads
// Removes from cache if post has been deleted
func (c *PostedCache) ValidateWithAPI(checkInstagram, checkThreads func(int, string) (bool, error)) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	toRemove := []int{}
	
	for id, record := range c.records {
		// Check Instagram posts
		if record.InstagramFeed && record.InstagramPostID != "" {
			exists, err := checkInstagram(id, record.InstagramPostID)
			if err != nil {
				log.Printf("⚠️  Could not verify post %d on Instagram: %v", id, err)
			} else if !exists {
				log.Printf("🗑️  Post %d deleted from Instagram - removing from cache", id)
				toRemove = append(toRemove, id)
				continue
			}
		}
		
		// Check Threads posts
		if record.Threads && record.ThreadsPostID != "" {
			exists, err := checkThreads(id, record.ThreadsPostID)
			if err != nil {
				log.Printf("⚠️  Could not verify post %d on Threads: %v", id, err)
			} else if !exists {
				log.Printf("🗑️  Post %d deleted from Threads - removing from cache", id)
				toRemove = append(toRemove, id)
				continue
			}
		}
	}
	
	// Remove deleted posts
	for _, id := range toRemove {
		delete(c.records, id)
	}
	
	if len(toRemove) > 0 {
		log.Printf("🔄 Removed %d deleted posts from cache", len(toRemove))
		if err := c.save(); err != nil {
			return fmt.Errorf("failed to save cache after cleanup: %w", err)
		}
	}
	
	return nil
}

// Stats returns cache statistics
func (c *PostedCache) Stats() map[string]int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	feedCount := 0
	storyCount := 0
	threadsCount := 0

	for _, record := range c.records {
		if record.InstagramFeed {
			feedCount++
		}
		if record.InstagramStory {
			storyCount++
		}
		if record.Threads {
			threadsCount++
		}
	}

	return map[string]int{
		"total_tournaments": len(c.records),
		"instagram_feed":    feedCount,
		"instagram_story":   storyCount,
		"threads":           threadsCount,
	}
}

