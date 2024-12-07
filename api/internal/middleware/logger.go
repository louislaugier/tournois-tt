package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger middleware logs request details
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Log request details
		gin.DefaultWriter.Write([]byte(
			"[GIN] " + time.Now().Format("2006/01/02 - 15:04:05") + " | " +
				c.Request.Method + " | " +
				c.Request.URL.Path + " | " +
				c.ClientIP() + " | " +
				latency.String() + " | " +
				strconv.Itoa(c.Writer.Status()) + "\n",
		))
	}
}
