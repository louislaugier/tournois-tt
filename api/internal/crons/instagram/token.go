package instagram

import (
	"log"

	"tournois-tt/api/pkg/instagram"
)

// CheckAndRefreshToken checks if the Instagram token needs refresh and refreshes if necessary
// This runs daily to ensure the token doesn't expire
func CheckAndRefreshToken() {
	log.Println("🔑 Checking Instagram token status...")

	expiresAt, daysRemaining, err := instagram.GetTokenInfo()
	if err != nil {
		log.Printf("⚠️  Failed to get token info: %v", err)
		log.Println("   Attempting to refresh token anyway...")
		if err := instagram.RefreshToken(); err != nil {
			log.Printf("❌ Token refresh failed: %v", err)
			log.Println("   ⚠️  MANUAL ACTION REQUIRED: Generate new token in Meta dashboard")
		} else {
			log.Println("✅ Token refreshed successfully")
		}
		return
	}

	log.Printf("   Token expires at: %s", expiresAt.Format("2006-01-02 15:04:05"))
	log.Printf("   Days remaining: %.1f days", daysRemaining)

	if daysRemaining < 7 {
		log.Printf("⚠️  Token expires in %.1f days, refreshing...", daysRemaining)
		if err := instagram.RefreshToken(); err != nil {
			log.Printf("❌ Token refresh failed: %v", err)
			if daysRemaining < 1 {
				log.Println("   🚨 URGENT: Token expires in less than 1 day!")
				log.Println("   ⚠️  MANUAL ACTION REQUIRED: Generate new token in Meta dashboard")
			}
		} else {
			newExpiresAt, newDaysRemaining, _ := instagram.GetTokenInfo()
			log.Printf("✅ Token refreshed successfully")
			log.Printf("   New expiration: %s (%.1f days)", newExpiresAt.Format("2006-01-02"), newDaysRemaining)
		}
	} else {
		log.Printf("✅ Token is valid (%.1f days remaining, threshold: 7 days)", daysRemaining)
	}
}
