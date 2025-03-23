package refresh

import (
	"fmt"
	"log"
	"time"

	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/utils"
)

// loadTournamentsWithRetry attempts to load tournaments with exponential backoff
func loadTournamentsWithRetry(maxRetries int, isCurrentSeason bool) (map[string]cache.TournamentCache, error) {
	var cachedTournaments map[string]cache.TournamentCache
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			// Calculate exponential backoff delay
			delaySeconds := RetryBaseDelay * attempt * attempt
			log.Printf("Tournament cache loading attempt %d/%d failed, retrying in %d seconds...",
				attempt-1, maxRetries, delaySeconds)
			time.Sleep(time.Duration(delaySeconds) * time.Second)
		}

		// Load existing tournaments from cache
		cachedTournaments, err = cache.LoadTournaments()
		if err == nil {
			return cachedTournaments, nil // Success
		}
	}

	// If all retries failed
	if isCurrentSeason {
		return nil, fmt.Errorf("critical error: failed to load tournament cache after %d attempts: %w", maxRetries, err)
	}
	return nil, fmt.Errorf("failed to load tournament cache: %w", err)
}

// retryLoadingForCurrentSeason attempts to load current season tournaments multiple times
// This handles cases where FFTT API refresh is in progress
func retryLoadingForCurrentSeason(maxRetries int, startDateAfter, startDateBefore *time.Time) []cache.TournamentCache {
	for retryAttempt := 1; retryAttempt <= maxRetries; retryAttempt++ {
		delaySeconds := WaitDelayForFFTTRefresh * retryAttempt
		log.Printf("No current season tournaments found, waiting %d seconds for possible FFTT API refresh to complete...", delaySeconds)
		time.Sleep(time.Duration(delaySeconds) * time.Second)

		// Try loading again
		cachedTournaments, err := cache.LoadTournaments()
		if err != nil {
			continue
		}

		// Convert and filter again
		tournamentsList := utils.MapToSlice(cachedTournaments)
		tournamentsToProcess := filterTournamentsForProcessing(tournamentsList, startDateAfter, startDateBefore)

		if len(tournamentsToProcess) > 0 {
			log.Printf("Found %d tournaments after retry attempt %d", len(tournamentsToProcess), retryAttempt)
			return tournamentsToProcess
		}
	}
	return nil
}
