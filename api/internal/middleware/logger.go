package middleware

import (
	"log"
	"time"
	"tournois-tt/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

// Logger returns a middleware that logs request details including IPs and User-Agent
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Get IPs before request is processed
		ips := utils.GetIPsFromRequest(c)

		// Process request
		c.Next()

		// Log request details after completion
		latency := time.Since(start)
		log.Printf("[%s] %s %s | Status: %d | Latency: %v | IPs: %s | User-Agent: %s",
			c.Request.Method,
			c.Request.URL.Path,
			c.Request.URL.RawQuery,
			c.Writer.Status(),
			latency,
			ips,
			c.Request.UserAgent(),
		)
	}
}
