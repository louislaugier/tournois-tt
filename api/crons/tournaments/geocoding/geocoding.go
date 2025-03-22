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

// RefreshTournamentsAndGeocoding fetches and updates tournament geocoding data
func RefreshTournamentsAndGeocoding(startDateAfter, startDateBefore *time.Time) error {
	// Check if we're querying for current season
	isCurrentSeason := isCurrentSeasonQuery(*startDateAfter, startDateBefore)

	// Configure retry parameters
	maxRetries := 3
	if !isCurrentSeason {
		maxRetries = 1 // Only retry once for historical data
	}

	// Fetch tournaments from FFTT with retries
	var tournaments []fftt.Tournament
	var err error
	var attempt int

	for attempt = 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			// Calculate exponential backoff delay: 5s, 20s, 60s
			delaySeconds := 5 * attempt * attempt
			log.Printf("FFTT API call attempt %d/%d failed, retrying in %d seconds...",
				attempt-1, maxRetries, delaySeconds)
			time.Sleep(time.Duration(delaySeconds) * time.Second)
		}

		// Attempt to fetch tournaments
		tournaments, err = fftt.GetFutureTournaments(*startDateAfter, startDateBefore)
		if err == nil {
			break // Success, exit retry loop
		}
	}

	// If all retries failed and we're in current season, it's critical
	if err != nil {
		if isCurrentSeason {
			if attempt > 1 {
				return fmt.Errorf("failed to fetch current season tournaments after %d attempts: %v", maxRetries, err)
			}
			return fmt.Errorf("failed to fetch tournaments: %v", err)
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

	// Prepare addresses for geocoding
	addressesToGeocode := make([]geocoding.Address, 0)
	newTournamentCacheEntries := make([]cache.TournamentCache, 0, len(tournaments))
	successCount := 0
	failureCount := 0

	// Load existing cache
	cachedTournaments, err := cache.LoadTournaments()
	if err != nil {
		log.Printf("Warning: Failed to load tournament cache: %v", err)
	}

	for _, t := range tournaments {
		// Check if this tournament is already in cache by ID
		cacheKey := fmt.Sprintf("%d", t.ID)
		cachedTournament, exists := cachedTournaments[cacheKey]

		if exists {
			// Use existing cached tournament data but update other fields if needed
			newCacheEntry := cachedTournament

			// Keep this tournament in the cache
			newTournamentCacheEntries = append(newTournamentCacheEntries, newCacheEntry)
			continue
		}

		// Create a new cache entry for this tournament
		newCacheEntry := cache.TournamentCache{
			ID:        t.ID,
			Name:      t.Name,
			Type:      utils.MapTournamentType(t.Type),
			StartDate: t.StartDate,
			EndDate:   t.EndDate,
			Address: cache.Address{
				StreetAddress:             t.Address.StreetAddress,
				PostalCode:                t.Address.PostalCode,
				AddressLocality:           t.Address.AddressLocality,
				DisambiguatingDescription: t.Address.DisambiguatingDescription,
			},
			Club: cache.Club{
				ID:         t.Club.ID,
				Name:       t.Club.Name,
				Code:       t.Club.Code,
				Department: t.Club.Department,
				Region:     t.Club.Region,
				Identifier: t.Club.Identifier,
			},
			Endowment:              t.Endowment,
			IsRulesPdfChecked:      false,
			IsSiteExistenceChecked: false,
			Timestamp:              time.Now(),
		}

		// Add Rules if available
		if t.Rules != nil {
			newCacheEntry.Rules = &cache.Rules{
				AgeMin:  t.Rules.AgeMin,
				AgeMax:  t.Rules.AgeMax,
				Points:  t.Rules.Points,
				Ranking: t.Rules.Ranking,
				URL:     t.Rules.URL,
			}
		}

		// Convert to geocoding.Address to use IsValid from the geocoding package
		geoAddress := geocoding.Address{
			StreetAddress:             t.Address.StreetAddress,
			PostalCode:                t.Address.PostalCode,
			AddressLocality:           t.Address.AddressLocality,
			DisambiguatingDescription: t.Address.DisambiguatingDescription,
		}

		// Check if the address is valid for geocoding
		if !geoAddress.IsValid() {
			newCacheEntry.Address.Failed = true
			newTournamentCacheEntries = append(newTournamentCacheEntries, newCacheEntry)
			continue
		}

		// Check if address is already in cache by address key
		addressKey := cache.GenerateAddressCacheKey(newCacheEntry.Address)
		addressFound := false

		// Look for an existing tournament with the same address
		for _, existingTournament := range cachedTournaments {
			existingAddressKey := cache.GenerateAddressCacheKey(existingTournament.Address)
			if existingAddressKey == addressKey && !existingTournament.Address.Failed {
				// Found a tournament with the same address that has coordinates
				newCacheEntry.Address.Latitude = existingTournament.Address.Latitude
				newCacheEntry.Address.Longitude = existingTournament.Address.Longitude
				addressFound = true
				break
			}
		}

		if !addressFound {
			// If address not found in cache, add to list for geocoding
			addressesToGeocode = append(addressesToGeocode, geoAddress)
			// Store the index in the newTournamentCacheEntries for updating later
			newCacheEntry.Address.Failed = true // Mark as failed by default, will update if geocoding succeeds
			newTournamentCacheEntries = append(newTournamentCacheEntries, newCacheEntry)
		} else {
			// Address found, no need to geocode
			newTournamentCacheEntries = append(newTournamentCacheEntries, newCacheEntry)
		}
	}

	// Perform individual geocoding for addresses not in cache
	for i, addr := range addressesToGeocode {
		// Rate limit between requests
		time.Sleep(geocoding.RateLimitDelay)

		// Geocode individual address using the geocoding package
		geocodeResult, err := geocoding.GetCoordinatesNominatim(addr)

		// If Nominatim fails, try Google as fallback (the package handles this internally)
		if err != nil || geocodeResult.Failed {
			// Try with Google as fallback
			geocodeResult, err = geocoding.GetCoordinatesGoogle(addr)
		}

		// Find the tournament with this address in our new cache entries
		for j := range newTournamentCacheEntries {
			cacheAddr := newTournamentCacheEntries[j].Address
			if cacheAddr.StreetAddress == addr.StreetAddress &&
				cacheAddr.PostalCode == addr.PostalCode &&
				cacheAddr.AddressLocality == addr.AddressLocality {

				if err != nil || geocodeResult.Failed {
					// Geocoding failed
					newTournamentCacheEntries[j].Address.Failed = true
					failureCount++
				} else {
					// Geocoding succeeded
					newTournamentCacheEntries[j].Address.Latitude = geocodeResult.Latitude
					newTournamentCacheEntries[j].Address.Longitude = geocodeResult.Longitude
					newTournamentCacheEntries[j].Address.Failed = false
					successCount++

					// Log successful geocoding in debug mode
					debugLog("Geocoded new address (%d/%d): %s -> (%f, %f)",
						i+1, len(addressesToGeocode),
						geocoding.ConstructFullAddress(addr),
						geocodeResult.Latitude, geocodeResult.Longitude)
				}

				// Save this single tournament to cache immediately
				tournamentToSave := []cache.TournamentCache{newTournamentCacheEntries[j]}
				if err := cache.SaveTournamentsToCache(tournamentToSave); err != nil {
					log.Printf("Warning: Failed to save geocoded tournament to cache: %v", err)
				} else {
					debugLog("Saved geocoded tournament %d to cache", newTournamentCacheEntries[j].ID)
				}

				break
			}
		}
	}

	// Log completion statistics
	log.Printf("Geocoding refresh completed: %d successful, %d failed", successCount, failureCount)
	return nil
}
