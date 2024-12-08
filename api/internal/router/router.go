package router

import (
	"tournois-tt/api/internal/config"
	"tournois-tt/api/internal/handlers"
	"tournois-tt/api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	router := gin.Default()
	router.ForwardedByClientIP = true

	// Trust Docker's default bridge network (172.17.0.0/16) and localhost
	router.SetTrustedProxies([]string{"172.0.0.0/8", "127.0.0.1", "::1"})

	router.Use(middleware.Logger())
	router.Use(gin.Recovery())

	router.Use(corsMiddleware())

	setupRoutes(router)

	return router
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", config.FrontendURL)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Origin")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func setupRoutes(router *gin.Engine) {
	v1 := router.Group("/v1")
	{
		v1.GET("/healthz", handlers.HealthzHandler)
		v1.GET("/tournaments", handlers.GetTournaments)
	}
}
