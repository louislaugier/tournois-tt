package geocoding

import (
	"fmt"
	"log"
	"time"
	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/fftt"
	geocodingPkg "tournois-tt/api/pkg/geocoding"
	"tournois-tt/api/pkg/utils"
)

// FetchTournamentsWithRetries attempts to fetch tournaments with configurable retries
func FetchTournamentsWithRetries(startDateAfter time.Time, startDateBefore *time.Time, maxRetries int) ([]fftt.Tournament, error) {
	var tournaments []fftt.Tournament
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			// Calculate exponential backoff delay: 5s, 20s, 60s
			delaySeconds := 5 * attempt * attempt
			log.Printf("FFTT API call attempt %d/%d failed, retrying in %d seconds...",
				attempt-1, maxRetries, delaySeconds)
			time.Sleep(time.Duration(delaySeconds) * time.Second)
		}

		// Attempt to fetch tournaments
		tournaments, err = fftt.GetFutureTournaments(startDateAfter, startDateBefore)
		if err == nil {
			return tournaments, nil // Success, exit retry loop
		}
	}

	return nil, err // All retries failed
}

// ProcessTournamentForCache prepares a tournament for caching and determines if geocoding is needed
// Returns: the new cache entry, whether the address needs geocoding, and the address if needed
func ProcessTournamentForCache(t fftt.Tournament, cachedTournaments map[string]cache.TournamentCache) (cache.TournamentCache, bool, geocodingPkg.Address) {
	// Check if this tournament is already in cache by ID
	cacheKey := fmt.Sprintf("%d", t.ID)
	cachedTournament, exists := cachedTournaments[cacheKey]

	if exists {
		// Use existing cached tournament data
		return cachedTournament, false, geocodingPkg.Address{}
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
	geoAddress := geocodingPkg.Address{
		StreetAddress:             t.Address.StreetAddress,
		PostalCode:                t.Address.PostalCode,
		AddressLocality:           t.Address.AddressLocality,
		DisambiguatingDescription: t.Address.DisambiguatingDescription,
	}

	// Check if the address is valid for geocoding
	if !geoAddress.IsValid() {
		newCacheEntry.Address.Failed = true
		return newCacheEntry, false, geocodingPkg.Address{}
	}

	// Check if address is already in cache by address key
	addressKey := cache.GenerateAddressCacheKey(newCacheEntry.Address)

	// Look for an existing tournament with the same address
	for _, existingTournament := range cachedTournaments {
		existingAddressKey := cache.GenerateAddressCacheKey(existingTournament.Address)
		if existingAddressKey == addressKey && !existingTournament.Address.Failed {
			// Found a tournament with the same address that has coordinates
			newCacheEntry.Address.Latitude = existingTournament.Address.Latitude
			newCacheEntry.Address.Longitude = existingTournament.Address.Longitude
			return newCacheEntry, false, geocodingPkg.Address{}
		}
	}

	// If we got here, the address needs geocoding
	newCacheEntry.Address.Failed = true // Mark as failed by default, will update if geocoding succeeds
	return newCacheEntry, true, geoAddress
}

// GeocodeAddresses processes a batch of addresses that need geocoding
// It updates the tournament cache entries with geocoding results
func GeocodeAddresses(addressesToGeocode []geocodingPkg.Address, tournamentsToUpdate []int, tournamentCacheEntries []cache.TournamentCache) ([]cache.TournamentCache, int, int) {
	successCount := 0
	failureCount := 0

	// Create a copy to avoid modifying the original slice during iteration
	updatedEntries := make([]cache.TournamentCache, len(tournamentCacheEntries))
	copy(updatedEntries, tournamentCacheEntries)

	for i, addr := range addressesToGeocode {
		// Rate limit between requests
		time.Sleep(geocodingPkg.RateLimitDelay)

		// Geocode individual address using the geocoding package
		geocodeResult, err := geocodingPkg.GetCoordinatesNominatim(addr)

		// If Nominatim fails, try Google as fallback
		if err != nil || geocodeResult.Failed {
			geocodeResult, err = geocodingPkg.GetCoordinatesGoogle(addr)
		}

		// Find the tournament to update
		tournamentIndex := tournamentsToUpdate[i]

		if err != nil || geocodeResult.Failed {
			// Geocoding failed
			updatedEntries[tournamentIndex].Address.Failed = true
			failureCount++
		} else {
			// Geocoding succeeded
			updatedEntries[tournamentIndex].Address.Latitude = geocodeResult.Latitude
			updatedEntries[tournamentIndex].Address.Longitude = geocodeResult.Longitude
			updatedEntries[tournamentIndex].Address.Failed = false
			successCount++

			// Log successful geocoding in debug mode
			debugLog("Geocoded new address (%d/%d): %s -> (%f, %f)",
				i+1, len(addressesToGeocode),
				geocodingPkg.ConstructFullAddress(addr),
				geocodeResult.Latitude, geocodeResult.Longitude)
		}
	}

	return updatedEntries, successCount, failureCount
}
