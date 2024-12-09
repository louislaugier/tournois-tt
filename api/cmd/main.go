package main

import (
	"log"

	"tournois-tt/api/internal/geocoding"
	"tournois-tt/api/internal/router"
)

func main() {
	go func() {
		log.Printf("Preloading tournament geocoding data...")
		if err := geocoding.PreloadTournaments(); err != nil {
			log.Printf("Warning: Failed to preload tournament data: %v", err)
		}
	}()

	r := router.NewRouter()

	log.Printf("Server starting...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
