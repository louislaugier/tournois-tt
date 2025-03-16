package main

import (
	"log"
	"tournois-tt/api/crons"
	"tournois-tt/api/crons/tournaments"
	"tournois-tt/api/internal/router"
)

func main() {
	// activities, err := helloasso.SearchActivities(context.Background(), "tournoi tennis de table courbevoie")
	// if err != nil {
	// 	log.Println(err)
	// }
	// log.Println(activities)

	// test.LogClubEmailAddresses()
	// test.LogCommitteeAndLeagueEmailAddresses()

	////////////////////////////////////////////////////////

	crons.Schedule()

	// Run geocoding refresh in a background goroutine
	go func() {
		tournaments.RefreshGeocoding()
	}()

	r := router.NewRouter()

	log.Printf("Server starting...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
