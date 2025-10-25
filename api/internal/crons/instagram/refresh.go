package instagram

import (
	"log"
	"time"

	"tournois-tt/api/internal/config"
)

// RefreshTokenOnStartup checks and refreshes the token when the app starts
func RefreshTokenOnStartup() {
	if !config.InstagramEnabled {
		log.Println("INFO: Instagram is disabled, skipping startup token refresh")
		return
	}

	if config.InstagramAccessToken == "" {
		log.Println("INFO: No Instagram access token configured, skipping startup token refresh")
		return
	}

	log.Println("INFO: Checking if Instagram token needs refresh on startup...")

	// Wait a bit to let the app fully start
	time.Sleep(5 * time.Second)

	CheckAndRefreshToken()
}

// RefreshThreadsTokenOnStartup checks and refreshes the Threads token when the app starts
func RefreshThreadsTokenOnStartup() {
	if !config.ThreadsEnabled {
		log.Println("INFO: Threads is disabled, skipping startup token refresh")
		return
	}

	if config.ThreadsAccessToken == "" {
		log.Println("INFO: No Threads access token configured, skipping startup token refresh")
		return
	}

	log.Println("INFO: Checking if Threads token needs refresh on startup...")

	// Wait a bit to let the app fully start
	time.Sleep(5 * time.Second)

	CheckAndRefreshThreadsToken()
}
