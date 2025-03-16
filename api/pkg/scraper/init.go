package scraper

import (
	"log"
)

// Initialize sets up the scraper package, including the cache
func Initialize() error {
	// Initialize the cache
	if err := InitCache(); err != nil {
		log.Printf("Warning: Failed to initialize scraper cache: %v", err)
		return err
	}

	// Prune expired entries
	if err := PruneExpiredEntries(); err != nil {
		log.Printf("Warning: Failed to prune expired cache entries: %v", err)
	}

	log.Printf("Scraper cache initialized with %d entries", Cache.Size())
	return nil
}
