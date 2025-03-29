package utils

import (
	"time"
	_ "time/tzdata"
)

// GetLastFinishedSeason returns the last completed season year.
// A season starts on July 1st and ends on June 30th of the following year.
// For example, if current date is May 2024, it will return 2022-2023 as the last finished season.
// Returns the start and end of the season in France time zone.
func GetLastFinishedSeason() (time.Time, time.Time) {
	return getSeasonDates(-1)
}

// GetCurrentSeason returns the current ongoing season year.
// A season starts on July 1st and ends on June 30th of the following year.
// For example, if current date is May 2024, it will return 2023-2024 as the current season.
// Returns the start and end of the season in France time zone.
func GetCurrentSeason() (time.Time, time.Time) {
	return getSeasonDates(0)
}

// getSeasonDates calculates season dates based on offset years
// offset can be 0 for current season, -1 for last finished season, etc.
func getSeasonDates(offset int) (time.Time, time.Time) {
	loc, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		// Fallback to UTC if timezone loading fails
		loc = time.UTC
	}

	now := time.Now().In(loc)
	currentYear := now.Year()

	// Check if we're in July-December
	if now.Month() >= time.July {
		// Current season started this year
		startYear := currentYear + offset
		endYear := currentYear + offset + 1
		return time.Date(startYear, time.July, 1, 0, 0, 0, 0, loc), time.Date(endYear, time.June, 30, 23, 59, 59, 999999999, loc)
	}

	// We're in January-June
	startYear := currentYear + offset - 1
	endYear := currentYear + offset
	return time.Date(startYear, time.July, 1, 0, 0, 0, 0, loc), time.Date(endYear, time.June, 30, 23, 59, 59, 999999999, loc)
}
