package main

import (
	"log"

	"tournois-tt/api/internal/crons"
	"tournois-tt/api/internal/crons/tournaments"
	"tournois-tt/api/internal/router"
)

func start() {
	go tournaments.RefreshListWithGeocoding()

	crons.Schedule()

	r := router.NewRouter()

	log.Printf("Server starting...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func main() {
	start()
}
