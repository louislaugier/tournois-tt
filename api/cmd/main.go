package main

import (
	"log"

	"api/internal/config"
	"api/internal/router"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	gin.SetMode(cfg.Environment)

	r := router.NewRouter()

	log.Printf("Server starting on %s", cfg.ServerAddress)
	if err := r.Run(cfg.ServerAddress); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
