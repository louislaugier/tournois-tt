package tournaments

import (
	"log"
	"time"
	"tournois-tt/api/pkg/utils"
)

func RefreshSignupURLs() {
	currentSeasonStart, _ := utils.GetCurrentSeason()
	if err := refreshSignupURLs(&currentSeasonStart, nil); err != nil {
		log.Printf("Warning: Failed to refresh tournament geocoding data: %v", err)
	}
}

func refreshSignupURLs(startDateAfter, startDateBefore *time.Time) error {
	// get tournaments from cache

	// for each, if signup url not empty
	// check hello asso with tournament name, tournament club, tournament city

	// if signup url not found, if !siteExistenceChecked, try to get site url
	// if site url found, check if tournament signup link available on site
	// if signup url found, add to json

	// if signup url not found, if pdf rules not checked, if has pdf rules, check pdf rules
	// if signup url found, add to json
	return nil
}
