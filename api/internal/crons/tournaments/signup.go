// Package tournaments contains cron jobs for processing tournaments data
package tournaments

import (
	"log"
	"time"

	"tournois-tt/api/internal/crons/tournaments/signup"
	"tournois-tt/api/pkg/utils"
)

// RefreshSignupURLs updates the signup URLs for tournaments
func RefreshSignupURLs() {
	log.Printf("Starting to refresh tournament signup URLs")
	signup.RefreshURLs()
	log.Printf("Finished refreshing tournament signup URLs")
}

// RefreshSignupURLsInRange updates the signup URLs for tournaments in a specific date range
func RefreshSignupURLsInRange(startDateAfter, startDateBefore *time.Time) error {
	log.Printf("Starting to refresh tournament signup URLs in specified date range")
	err := signup.RefreshURLsInRange(startDateAfter, startDateBefore)
	if err != nil {
		return err
	}
	log.Printf("Finished refreshing tournament signup URLs in specified date range")
	return nil
}

// UpdateSignupURLsForCurrentSeason updates signup URLs for tournaments in the current season
func UpdateSignupURLsForCurrentSeason() error {
	currentSeasonStart, currentSeasonEnd := utils.GetCurrentSeason()
	log.Printf("Starting to refresh tournament signup URLs for current season")
	err := signup.RefreshURLsInRange(&currentSeasonStart, &currentSeasonEnd)
	if err != nil {
		// Any error for current season is critical, but this happens after multiple retries
		log.Fatalf("Critical error in current season signup URL refresh after multiple retry attempts: %v", err)
		return err // This line won't execute due to Fatalf, but included for clarity
	}
	log.Printf("Finished refreshing tournament signup URLs for current season")
	return nil
}
