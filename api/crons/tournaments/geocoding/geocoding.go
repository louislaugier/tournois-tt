package geocoding

import (
	"fmt"
	"log"
	"time"
	"tournois-tt/api/pkg/cache"
	geocodingPkg "tournois-tt/api/pkg/geocoding"
	"tournois-tt/api/pkg/utils"
)

// Debug flag to control verbose logging
var Debug = false

// debugLog logs a message only if Debug is true
func debugLog(format string, args ...interface{}) {
	if Debug {
		log.Printf(format, args...)
	}
}

// isCurrentSeasonQuery checks if the query date range is part of the current season
func isCurrentSeasonQuery(startDateAfter time.Time, startDateBefore *time.Time) bool {
	currentSeasonStart, currentSeasonEnd := utils.GetCurrentSeason()

	// Check if startDateAfter is within or after the current season start
	isWithinCurrentSeason := !startDateAfter.Before(currentSeasonStart)

	// Check if startDateBefore is within or equal to current season end (if provided)
	if startDateBefore != nil {
		isWithinCurrentSeason = isWithinCurrentSeason && !startDateBefore.After(currentSeasonEnd)
	}

	return isWithinCurrentSeason
}

// RefreshTournamentsAndGeocoding fetches and updates tournament geocoding data
func RefreshTournamentsAndGeocoding(startDateAfter, startDateBefore *time.Time) error {
	if startDateAfter == nil {
		now := time.Now()
		startDateAfter = &now
	}

	// Check if we're querying for current season
	isCurrentSeason := isCurrentSeasonQuery(*startDateAfter, startDateBefore)

	// Configure retry parameters
	maxRetries := 3
	if !isCurrentSeason {
		maxRetries = 1 // Only retry once for historical data
	}

	// Fetch tournaments from FFTT with retries
	tournaments, err := FetchTournamentsWithRetries(*startDateAfter, startDateBefore, maxRetries)

	// Handle errors based on whether it's current season or historical data
	if err != nil {
		if isCurrentSeason {
			return fmt.Errorf("failed to fetch current season tournaments after %d attempts: %v", maxRetries, err)
		} else {
			// For historical data, just log a warning
			log.Printf("Warning: Failed to fetch historical tournaments: %v", err)
			return err
		}
	}

	// Check for empty response
	if len(tournaments) == 0 {
		if isCurrentSeason {
			// Empty response for current season is a critical error
			return fmt.Errorf("no tournaments found in current season date range")
		} else {
			// Empty response for historical data is just a warning
			log.Printf("No tournaments found in specified date range from %v to %v",
				startDateAfter.Format("2006-01-02"),
				startDateBefore.Format("2006-01-02"))
		}
	}

	// Load existing cache
	cachedTournaments, err := cache.LoadTournaments()
	if err != nil {
		log.Printf("Warning: Failed to load tournament cache: %v", err)
	}

	// Prepare for processing
	addressesToGeocode := make([]geocodingPkg.Address, 0)
	newTournamentCacheEntries := make([]cache.TournamentCache, 0, len(tournaments))
	tournamentsNeedingGeocoding := make([]int, 0)

	// Process each tournament
	for _, t := range tournaments {
		cacheEntry, needsGeocoding, address := ProcessTournamentForCache(t, cachedTournaments)

		// Add to our list of cache entries
		newTournamentCacheEntries = append(newTournamentCacheEntries, cacheEntry)

		// If this entry needs geocoding, add to our list
		if needsGeocoding {
			addressesToGeocode = append(addressesToGeocode, address)
			tournamentsNeedingGeocoding = append(tournamentsNeedingGeocoding, len(newTournamentCacheEntries)-1)
		}
	}

	// Process geocoding for addresses that need it
	if len(addressesToGeocode) > 0 {
		updatedEntries, successCount, failureCount := GeocodeAddresses(
			addressesToGeocode,
			tournamentsNeedingGeocoding,
			newTournamentCacheEntries,
		)
		newTournamentCacheEntries = updatedEntries

		// Log completion statistics
		log.Printf("Geocoding refresh completed: %d successful, %d failed", successCount, failureCount)
	} else {
		log.Printf("Geocoding refresh completed: no new addresses to geocode")
	}

	// Save all tournaments at once after processing is complete
	if err := cache.SaveTournamentsToCache(newTournamentCacheEntries); err != nil {
		log.Printf("Warning: Failed to save all geocoded tournaments to cache: %v", err)
	} else {
		debugLog("Saved all %d geocoded tournaments to cache", len(newTournamentCacheEntries))
	}

	return nil
}
