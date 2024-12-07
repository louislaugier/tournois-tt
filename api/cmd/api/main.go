package main

import (
	"log"

	"api/internal/handlers"
	"api/internal/middleware"
	"api/pkg/config"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize configuration
	cfg := config.LoadConfig()

	// Set Gin mode
	gin.SetMode(cfg.Environment)

	// Initialize router
	router := initRouter()

	// Initialize services
	initServices(router)

	// Start server
	log.Printf("Server starting on %s", cfg.ServerAddress)
	if err := router.Run(cfg.ServerAddress); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func initRouter() *gin.Engine {
	router := gin.New()

	// Use logger and recovery middleware
	router.Use(middleware.Logger())
	router.Use(gin.Recovery())

	// Enable CORS
	router.Use(corsMiddleware())

	// Setup routes
	setupRoutes(router)

	return router
}

func initServices(router *gin.Engine) {
	// TODO: Initialize services (database, cache, etc.)
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func setupRoutes(router *gin.Engine) {
	// Health check endpoint
	router.GET("/api/health", handlers.HealthCheck)

	// API version group
	v1 := router.Group("/api/v1")
	{
		// TODO: Add your API endpoints here
		_ = v1
	}
}
