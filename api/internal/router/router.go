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

	// Only trust nginx reverse proxy
	router.SetTrustedProxies([]string{"nginx"})

	router.Use(gin.Recovery())
	router.Use(middleware.RateLimiter())
	router.Use(corsMiddleware())

	setupRoutes(router)

	return router
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", config.FrontendURL)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Origin")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

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
		v1.GET("/tournaments", middleware.Logger(), handlers.TournamentsHandler)
		v1.POST("/newsletter", handlers.NewsletterHandler)
	}

	// Direct redirect from root id to rules pdf: /:id -> rules url or '/'
	router.GET("/:id", handlers.RedirectRulesHandler)
}
