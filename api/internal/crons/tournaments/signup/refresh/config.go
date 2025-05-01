// Package refresh provides functionality for refreshing tournament signup URLs
package refresh

import (
	"time"
	"tournois-tt/api/pkg/utils"
)

// Constants for retry and timeout configuration
const (
	// MaxRetriesCurrentSeason is the number of retries for current season operations
	MaxRetriesCurrentSeason = 3
	// MaxRetriesHistorical is the number of retries for historical data operations
	MaxRetriesHistorical = 1
	// RetryBaseDelay is the base delay in seconds for retry attempts
	RetryBaseDelay = 5
	// WaitDelayForFFTTRefresh is the delay in seconds between retries when waiting for FFTT data
	WaitDelayForFFTTRefresh = 30
	// NumWorkers is the default number of concurrent workers
	NumWorkers = 3
)

// IsCurrentSeasonQuery checks if the query date range is part of the current season
func IsCurrentSeasonQuery(startDateAfter, startDateBefore *time.Time) bool {
	currentSeasonStart, currentSeasonEnd := utils.GetCurrentSeason()

	// If startDateAfter is nil, we're starting from before current season
	if startDateAfter == nil {
		return false
	}

	// Check if startDateAfter is within or after the current season start
	isWithinCurrentSeason := !startDateAfter.Before(currentSeasonStart)

	// Check if startDateBefore is within or equal to current season end (if provided)
	if startDateBefore != nil {
		isWithinCurrentSeason = isWithinCurrentSeason && !startDateBefore.After(currentSeasonEnd)
	}

	return isWithinCurrentSeason
}
