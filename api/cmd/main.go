package main

import (
	"log"

	"tournois-tt/api/internal/router"
)

func main() {
	r := router.NewRouter()

	log.Printf("Server starting...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
