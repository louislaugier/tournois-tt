package main

import (
	"log"
	"tournois-tt/api/internal/crons/tournaments"
)

func main() {
	log.Println("🔄 Manually triggering tournament refresh...")
	tournaments.RefreshListWithGeocoding()
	log.Println("✅ Done!")
}
