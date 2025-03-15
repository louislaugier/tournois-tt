package tournaments

import (
	"log"
	"tournois-tt/api/pkg/utils"
)

func Refresh() {
	lastSeasonStart, _ := utils.GetLastFinishedSeason()

	if err := refreshGeocoding(&lastSeasonStart, nil); err != nil {
		log.Printf("Warning: Failed to refresh tournament geocoding data: %v", err)
	}

	// refresh signup urls
}
