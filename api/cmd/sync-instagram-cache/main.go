package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/instagram"
)

func main() {
	removeFlag := flag.String("remove", "", "Remove tournament IDs from cache (comma-separated)")
	flag.Parse()

	if *removeFlag != "" {
		// Remove specific tournaments from cache
		removeTournaments(*removeFlag)
		return
	}

	// Show cache stats
	showCacheStats()
}

func removeTournaments(idsStr string) {
	parts := strings.Split(idsStr, ",")
	if len(parts) == 0 {
		log.Fatal("No IDs provided")
	}

	cache := instagram.GetPostedCache()
	
	for _, idStr := range parts {
		var id int
		if _, err := fmt.Sscanf(strings.TrimSpace(idStr), "%d", &id); err != nil {
			log.Printf("⚠️  Invalid ID: %s", idStr)
			continue
		}

		if posted, record := cache.IsPosted(id); !posted {
			log.Printf("⚠️  Tournament %d not in cache", id)
		} else {
			log.Printf("🗑️  Removing tournament %d (%s) from cache", id, record.TournamentName)
			cache.Remove(id)
		}
	}

	log.Println()
	log.Println("✅ Cache updated successfully!")
	showCacheStats()
}

func showCacheStats() {
	cache := instagram.GetPostedCache()
	stats := cache.Stats()
	records := cache.GetAllRecords()

	log.Println("📊 Instagram Posted Cache Statistics")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("Total tournaments posted: %d", stats["total_tournaments"])
	log.Printf("  • Instagram Feed:       %d", stats["instagram_feed"])
	log.Printf("  • Instagram Story:      %d", stats["instagram_story"])
	log.Printf("  • Threads:              %d", stats["threads"])
	log.Println()

	if len(records) > 0 {
		log.Println("Recent posts:")
		count := 10
		if len(records) < count {
			count = len(records)
		}
		for i := 0; i < count; i++ {
			r := records[len(records)-1-i] // Show most recent first
			platforms := []string{}
			if r.InstagramFeed {
				platforms = append(platforms, "Feed")
			}
			if r.InstagramStory {
				platforms = append(platforms, "Story")
			}
			if r.Threads {
				platforms = append(platforms, "Threads")
			}
			log.Printf("  • ID %d - %s (%s) - %s",
				r.TournamentID,
				r.TournamentName,
				strings.Join(platforms, "+"),
				r.PostedAt.Format("2006-01-02 15:04"))
		}
		
		if len(records) > count {
			log.Printf("  ... and %d more", len(records)-count)
		}
	}
}

func loadTournaments() ([]cache.TournamentCache, error) {
	possiblePaths := []string{
		"./cache/data.json",
		"../cache/data.json",
		"../../cache/data.json",
	}

	var dataPath string
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			dataPath = path
			break
		}
	}

	if dataPath == "" {
		wd, _ := os.Getwd()
		current := wd
		for current != "/" && current != "." {
			candidate := filepath.Join(current, "cache", "data.json")
			if _, err := os.Stat(candidate); err == nil {
				dataPath = candidate
				break
			}
			current = filepath.Dir(current)
		}
	}

	if dataPath == "" {
		return nil, fmt.Errorf("data.json not found")
	}

	payload, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, err
	}

	var tournaments []cache.TournamentCache
	if err := json.Unmarshal(payload, &tournaments); err != nil {
		return nil, err
	}

	return tournaments, nil
}

