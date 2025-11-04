package main

import (
	"log"
	instagramCron "tournois-tt/api/internal/crons/instagram"
)

func main() {
	log.Println("🔄 Manually triggering Instagram cache sync...")
	log.Println()
	
	instagramCron.SyncPostedCache()
	
	log.Println()
	log.Println("✅ Done!")
}

