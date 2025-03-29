package refresh

import (
	"fmt"
	"log"
	"time"

	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/utils"
)

// URLs refreshes signup URLs for all tournaments in the cache
func URLs() {
	// refresh all tournaments in the current season
	err := URLsInRange(nil, nil)
	if err != nil {
		log.Printf("Error refreshing signup URLs: %v", err)
	}
}

// URLsInRange refreshes signup URLs for tournaments within a date range
// If startDateAfter and startDateBefore are both nil, refreshes the current season
func URLsInRange(startDateAfter, startDateBefore *time.Time) error {
	// Check if we're querying for current season
	isCurrentSeason := isCurrentSeasonQuery(startDateAfter, startDateBefore)

	// Configure retry parameters based on season type
	maxRetries := MaxRetriesCurrentSeason
	if !isCurrentSeason {
		maxRetries = MaxRetriesHistorical
	}

	// Load tournaments with retry mechanism
	cachedTournaments, err := loadTournamentsWithRetry(maxRetries, isCurrentSeason)
	if err != nil {
		return err
	}

	// Convert map to slice for processing
	tournamentsList := utils.MapToSlice(cachedTournaments)

	// Filter tournaments that need processing
	tournamentsToProcess := filterTournamentsForProcessing(tournamentsList, startDateAfter, startDateBefore)

	// For current season, retry if no tournaments found (may be during FFTT API refresh)
	if isCurrentSeason && len(tournamentsToProcess) == 0 {
		tournamentsToProcess = retryLoadingForCurrentSeason(maxRetries, startDateAfter, startDateBefore)
	}

	// Check if we have tournaments to process
	if len(tournamentsToProcess) == 0 {
		return handleNoTournamentsCase(isCurrentSeason)
	}

	log.Printf("Processing %d tournaments for signup URL refresh", len(tournamentsToProcess))

	// Process tournaments using worker pool
	processedTournaments, err := processTournamentBatch(tournamentsToProcess)
	if err != nil {
		return fmt.Errorf("error processing tournament batch: %w", err)
	}

	// Save results if needed
	if len(processedTournaments) > 0 {
		if err := cache.SaveTournamentsToCache(processedTournaments); err != nil {
			return fmt.Errorf("error saving processed tournaments: %w", err)
		}
		log.Printf("Successfully saved %d processed tournaments", len(processedTournaments))
	}

	log.Printf("Successfully processed %d tournaments", len(processedTournaments))
	return nil
}

// handleNoTournamentsCase handles the case when no tournaments need processing
func handleNoTournamentsCase(isCurrentSeason bool) error {
	if isCurrentSeason {
		// For current season, no tournaments to process is a critical error
		return fmt.Errorf("critical error: no tournaments found to process for current season signup URL refresh after multiple attempts")
	}
	log.Printf("No tournaments need signup URL refresh")
	return nil
}
