package geocoding

import (
	"fmt"
	"log"
	"time"
	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/fftt"
	"tournois-tt/api/pkg/geocoding"
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

// fetchAndValidateTournaments fetches tournaments and validates the response based on season context
func fetchAndValidateTournaments(startDateAfter time.Time, startDateBefore *time.Time, isCurrentSeason bool) ([]fftt.Tournament, error) {
	// Configure retry parameters
	maxRetries := 3
	if !isCurrentSeason {
		maxRetries = 1 // Only retry once for historical data
	}

	// Fetch tournaments from FFTT with retries
	tournaments, err := FetchTournamentsWithRetries(startDateAfter, startDateBefore, maxRetries)

	// Handle errors based on whether it's current season or historical data
	if err != nil {
		if isCurrentSeason {
			return nil, fmt.Errorf("failed to fetch current season tournaments after %d attempts: %v", maxRetries, err)
		} else {
			// For historical data, just log a warning
			log.Printf("Warning: Failed to fetch historical tournaments: %v", err)
			return nil, err
		}
	}

	// Check for empty response
	if len(tournaments) == 0 {
		if isCurrentSeason {
			// Empty response for current season is a critical error
			return nil, fmt.Errorf("no tournaments found in current season date range")
		} else {
			// Empty response for historical data is just a warning
			log.Printf("No tournaments found in specified date range from %v to %v",
				startDateAfter.Format("2006-01-02"),
				startDateBefore.Format("2006-01-02"))
		}
	}

	return tournaments, nil
}

// prepareTournamentsForGeocoding processes tournaments and identifies those needing geocoding
func prepareTournamentsForGeocoding(tournaments []fftt.Tournament) ([]cache.TournamentCache, []geocoding.Address, []int, error) {
	// Load existing cache
	cachedTournaments, err := cache.LoadTournaments()
	if err != nil {
		log.Printf("Warning: Failed to load tournament cache: %v", err)
	}

	// Prepare for processing
	addressesToGeocode := make([]geocoding.Address, 0)
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

	return newTournamentCacheEntries, addressesToGeocode, tournamentsNeedingGeocoding, nil
}

// performGeocoding executes geocoding for addresses that need it
func performGeocoding(
	tournamentCacheEntries []cache.TournamentCache,
	addressesToGeocode []geocoding.Address,
	tournamentsNeedingGeocoding []int,
) ([]cache.TournamentCache, error) {
	if len(addressesToGeocode) > 0 {
		updatedEntries, successCount, failureCount := GeocodeAddresses(
			addressesToGeocode,
			tournamentsNeedingGeocoding,
			tournamentCacheEntries,
		)

		// Log completion statistics
		log.Printf("Geocoding refresh completed: %d successful, %d failed", successCount, failureCount)

		return updatedEntries, nil
	}

	log.Printf("Geocoding refresh completed: no new addresses to geocode")
	return tournamentCacheEntries, nil
}

// saveTournamentCache saves the updated tournaments to cache
func saveTournamentCache(tournamentCacheEntries []cache.TournamentCache) error {
	err := cache.SaveTournamentsToCache(tournamentCacheEntries)
	if err != nil {
		log.Printf("Warning: Failed to save all geocoded tournaments to cache: %v", err)
		return err
	}

	debugLog("Saved all %d geocoded tournaments to cache", len(tournamentCacheEntries))
	return nil
}

// RefreshTournamentsAndGeocoding fetches and updates tournament geocoding data
func RefreshTournamentsAndGeocoding(startDateAfter, startDateBefore *time.Time) error {
	if startDateAfter == nil {
		now := time.Now()
		startDateAfter = &now
	}

	// Check if we're querying for current season
	isCurrentSeason := isCurrentSeasonQuery(*startDateAfter, startDateBefore)

	// Fetch and validate tournaments
	tournaments, err := fetchAndValidateTournaments(*startDateAfter, startDateBefore, isCurrentSeason)
	if err != nil {
		return err
	}

	// Skip further processing if no tournaments were found
	if len(tournaments) == 0 {
		return nil
	}

	// Prepare tournaments for geocoding
	tournamentCacheEntries, addressesToGeocode, tournamentsNeedingGeocoding, err := prepareTournamentsForGeocoding(tournaments)
	if err != nil {
		return fmt.Errorf("error preparing tournaments for geocoding: %v", err)
	}

	// Perform geocoding for addresses that need it
	updatedEntries, err := performGeocoding(tournamentCacheEntries, addressesToGeocode, tournamentsNeedingGeocoding)
	if err != nil {
		return fmt.Errorf("error during geocoding: %v", err)
	}

	// Save all tournaments to cache
	if err := saveTournamentCache(updatedEntries); err != nil {
		return fmt.Errorf("error saving tournaments to cache: %v", err)
	}

	return nil
}
