package refresh

import (
	"log"
	"time"

	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/utils"
)

// filterTournamentsForProcessing filters tournaments that need signup URL refresh
func filterTournamentsForProcessing(tournaments []cache.TournamentCache, startDateAfter, startDateBefore *time.Time) []cache.TournamentCache {
	var result []cache.TournamentCache

	for _, tournament := range tournaments {
		// Parse the tournament date
		tournamentDate, err := utils.ParseTournamentDate(tournament.StartDate)
		if err != nil {
			// Tournaments should always have a valid date, log error but don't crash
			log.Printf("Error: Failed to parse tournament date for tournament %s (ID: %d): %v",
				tournament.Name, tournament.ID, err)
			continue
		}

		// Skip tournaments outside our date range
		if startDateAfter != nil && tournamentDate.Before(*startDateAfter) {
			continue
		}
		if startDateBefore != nil && tournamentDate.After(*startDateBefore) {
			continue
		}

		// Skip tournaments that already have page URLs
		if tournament.Page != "" {
			continue
		}

		// Add to list of tournaments to process
		result = append(result, tournament)
	}

	return result
}
