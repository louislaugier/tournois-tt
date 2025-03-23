package signup

import (
	"time"

	"tournois-tt/api/internal/crons/tournaments/signup/refresh"
)

// RefreshURLs refreshes signup URLs for all tournaments in the cache
func RefreshURLs() {
	refresh.URLs()
}

// RefreshURLsInRange refreshes signup URLs for tournaments within a date range
// If startDateAfter and startDateBefore are both nil, refreshes the current season
func RefreshURLsInRange(startDateAfter, startDateBefore *time.Time) error {
	return refresh.URLsInRange(startDateAfter, startDateBefore)
}
