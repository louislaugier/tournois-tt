package instagram

import (
	"log"
	"tournois-tt/api/internal/config"
	instagramapi "tournois-tt/api/pkg/instagram/api"
)

// SyncPostedCache validates the posted cache against Instagram/Threads APIs
// Removes entries for posts that have been deleted
func SyncPostedCache() {
	log.Println("🔄 Syncing Instagram posted cache with APIs...")

	if !config.InstagramEnabled {
		log.Println("⚠️  Instagram disabled - skipping cache sync")
		return
	}

	// Create Instagram client
	instagramConfig := instagramapi.Config{
		AccessToken:        config.InstagramAccessToken,
		PageID:             config.InstagramPageID,
		ThreadsAccessToken: config.ThreadsAccessToken,
		ThreadsUserID:      config.ThreadsUserID,
		Enabled:            config.InstagramEnabled,
		ThreadsEnabled:     config.ThreadsEnabled,
	}

	client := instagramapi.NewClient(instagramConfig)

	// Validate cache
	if err := client.SyncCacheWithAPI(); err != nil {
		log.Printf("⚠️  Cache sync failed: %v", err)
		return
	}

	log.Println("✅ Cache sync completed successfully")
}

