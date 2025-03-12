package tournaments

import (
	"log"
	"time"
	"tournois-tt/api/pkg/utils"
)

func RefreshSignupURLs() {
	currentSeasonStart, _ := utils.GetCurrentSeason()
	if err := refreshSignupURLs(&currentSeasonStart, nil); err != nil {
		log.Printf("Warning: Failed to refresh tournament signup URLs: %v", err)
	}
}

// refreshSignupURLs fetches and processes tournament signup urls
func refreshSignupURLs(startDateAfter, startDateBefore *time.Time) error {
	return nil
}
