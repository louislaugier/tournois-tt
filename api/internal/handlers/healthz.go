package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthzHandler handles the health check endpoint
func HealthzHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}
