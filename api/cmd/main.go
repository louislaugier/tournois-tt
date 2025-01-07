package main

import (
	"log"
	"tournois-tt/api/crons"
	"tournois-tt/api/internal/router"
	"tournois-tt/api/pkg/geocoding"
	"tournois-tt/api/pkg/utils"
)

func main() {
	// test.LogClubEmailAddresses()
	// test.LogCommitteeAndLeagueEmailAddresses()

	crons.Schedule()

	////////////////////////////////////////////////////////

	go func() {
		log.Printf("Preloading tournament geocoding data...")

		lastSeasonStart, _ := utils.GetLatestFinishedSeason()
		if err := geocoding.PreloadTournaments(&lastSeasonStart, nil); err != nil {
			log.Printf("Warning: Failed to preload tournament data: %v", err)
		}
	}()

	r := router.NewRouter()

	log.Printf("Server starting...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
