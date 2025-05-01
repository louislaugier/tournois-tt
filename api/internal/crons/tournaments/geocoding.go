package tournaments

import (
	"log"
	"tournois-tt/api/internal/crons/tournaments/geocoding"
	"tournois-tt/api/pkg/utils"
)

func RefreshListWithGeocoding() {
	lastSeasonStart, _ := utils.GetLastFinishedSeason()
	currentSeasonStart, currentSeasonEnd := utils.GetCurrentSeason()

	// First refresh historical tournaments (non-critical operation)
	if err := geocoding.RefreshGeocoding(&lastSeasonStart, &currentSeasonStart); err != nil {
		log.Printf("Warning: Failed to refresh historical tournament geocoding data: %v", err)
	}

	// Then refresh current season tournaments (critical operation)
	if err := geocoding.RefreshGeocoding(&currentSeasonStart, &currentSeasonEnd); err != nil {
		// Fatal error after multiple retry attempts
		log.Fatalf("Critical error: Failed to refresh current season tournament geocoding data after multiple attempts: %v", err)
	}
}
