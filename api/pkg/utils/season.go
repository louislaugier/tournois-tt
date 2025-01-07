package utils

import (
	"time"
	_ "time/tzdata"
)

// GetLatestFinishedSeason returns the latest completed season year.
// A season starts on July 1st and ends on June 30th of the following year.
// For example, if current date is May 2024, it will return 2022-2023 as the latest finished season.
// Returns the start and end of the season in France time zone.
func GetLatestFinishedSeason() (time.Time, time.Time) {
	loc, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		// Fallback to UTC if timezone loading fails
		loc = time.UTC
	}

	now := time.Now().In(loc)
	currentYear := now.Year()

	// Check if we're in July-December
	if now.Month() >= time.July {
		// Current season started this year, so last finished season was previous year
		return time.Date(currentYear-1, time.July, 1, 0, 0, 0, 0, loc), time.Date(currentYear, time.June, 30, 23, 59, 59, 999999999, loc)
	}

	// We're in January-June
	// If we haven't passed June 30th, we're still in previous season
	return time.Date(currentYear-2, time.July, 1, 0, 0, 0, 0, loc), time.Date(currentYear-1, time.June, 30, 23, 59, 59, 999999999, loc)
}
