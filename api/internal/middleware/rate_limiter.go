package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

const (
	requestsPerMinute = 45
	burstSize         = 5
)

var limiterMap = make(map[string]*rate.Limiter)

// RateLimiter returns a middleware that limits requests per host
func RateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the real IP, considering X-Forwarded-For and X-Real-IP headers
		clientIP := c.ClientIP()

		// Get or create limiter for this IP
		limiter := getLimiter(clientIP)

		// Check if request can proceed
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func getLimiter(clientIP string) *rate.Limiter {
	// Create a new limiter if it doesn't exist
	if limiter, exists := limiterMap[clientIP]; exists {
		return limiter
	}

	// Create a new rate limiter: requestsPerMinute requests per minute
	limiter := rate.NewLimiter(rate.Every(time.Minute/requestsPerMinute), burstSize)
	limiterMap[clientIP] = limiter
	return limiter
}
