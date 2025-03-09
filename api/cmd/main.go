package main

import (
	"log"
	"tournois-tt/api/crons"
	"tournois-tt/api/internal/router"
)

func main() {
	// test.LogClubEmailAddresses()
	// test.LogCommitteeAndLeagueEmailAddresses()

	crons.Schedule()

	////////////////////////////////////////////////////////

	go func() {
		log.Printf("Refreshing tournament geocoding data...")
		crons.RefreshTournaments()
	}()

	r := router.NewRouter()

	log.Printf("Server starting...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
