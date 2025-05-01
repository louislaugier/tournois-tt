package geocoding

import (
	"fmt"
	"log"
	"net/url"
	"time"
	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/fftt"
	"tournois-tt/api/pkg/geocoding"
)

// FetchTournamentsWithRetries attempts to fetch tournaments with configurable retries
func FetchTournamentsWithRetries(startDateAfter time.Time, startDateBefore *time.Time, maxRetries int) ([]fftt.Tournament, error) {
	var tournaments []fftt.Tournament
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			// Calculate exponential backoff delay: 5s, 20s, 60s
			delay := time.Duration(attempt*attempt) * 5 * time.Second
			log.Printf("Retry attempt %d/%d after %v delay", attempt, maxRetries, delay)
			time.Sleep(delay)
		}

		// Create query parameters
		params := url.Values{}
		params.Set("startDate[after]", startDateAfter.Format("2006-01-02T15:04:05"))
		if startDateBefore != nil {
			params.Set("startDate[before]", startDateBefore.Format("2006-01-02T15:04:05"))
		}
		params.Set("itemsPerPage", "999999")
		params.Set("order[startDate]", "asc")

		// Try to fetch tournaments
		tournaments, err = fftt.FetchTournaments(params)
		if err == nil && len(tournaments) > 0 {
			return tournaments, nil
		}
	}

	// If we exhausted all retries, return the last error
	return tournaments, err
}

// ProcessTournamentForCache prepares a tournament for caching and determines if it needs geocoding
func ProcessTournamentForCache(t fftt.Tournament, cachedTournaments map[string]cache.TournamentCache) (cache.TournamentCache, bool, geocoding.Address) {
	// Create a new cache entry with tournament data from API
	newCacheEntry := cache.TournamentCache{
		ID:        t.ID,
		Name:      t.Name,
		Type:      t.Type,
		StartDate: t.StartDate,
		EndDate:   t.EndDate,
		Address: geocoding.Address{
			StreetAddress:             t.Address.StreetAddress,
			PostalCode:                t.Address.PostalCode,
			AddressLocality:           t.Address.AddressLocality,
			DisambiguatingDescription: t.Address.DisambiguatingDescription,
			Latitude:                  t.Address.Latitude,
			Longitude:                 t.Address.Longitude,
			Failed:                    t.Address.Failed,
		},
		Club: cache.Club{
			ID:         t.Club.ID,
			Name:       t.Club.Name,
			Code:       t.Club.Code,
			Department: t.Club.Department,
			Region:     t.Club.Region,
			Identifier: t.Club.Identifier,
		},
		Timestamp: time.Now(),
	}

	// When endowment is 0 (null), calculate it from the tables
	if t.Endowment == 0 && len(t.Tables) > 0 {
		totalEndowment := 0
		for _, table := range t.Tables {
			totalEndowment += table.Endowment
		}
		newCacheEntry.Endowment = totalEndowment
	} else {
		newCacheEntry.Endowment = t.Endowment
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

	// Check if already in cache with complete geocoding
	cacheKey := fmt.Sprintf("%d", t.ID)
	if cachedTournament, exists := cachedTournaments[cacheKey]; exists {
		if cachedTournament.Address.Latitude != 0 && cachedTournament.Address.Longitude != 0 {
			// Tournament exists with coordinates - preserve them and update other data
			newCacheEntry.Address.Latitude = cachedTournament.Address.Latitude
			newCacheEntry.Address.Longitude = cachedTournament.Address.Longitude
			newCacheEntry.Address.Failed = cachedTournament.Address.Failed

			// Already geocoded, no need to process
			return newCacheEntry, false, geocoding.Address{}
		}
	}

	// Convert to models.Address to use IsValid
	geoAddress := geocoding.Address{
		StreetAddress:             t.Address.StreetAddress,
		PostalCode:                t.Address.PostalCode,
		AddressLocality:           t.Address.AddressLocality,
		DisambiguatingDescription: t.Address.DisambiguatingDescription,
	}

	// Check if the address already has coordinates or if it's invalid
	if !geocoding.IsAddressValid(geoAddress) {
		// Skip invalid addresses
		return newCacheEntry, false, geocoding.Address{}
	}

	// Skip addresses that already have coordinates
	if t.Address.Latitude != 0 && t.Address.Longitude != 0 {
		return newCacheEntry, false, geocoding.Address{}
	}

	// This address needs geocoding
	return newCacheEntry, true, geoAddress
}

// GeocodeAddresses processes a batch of addresses that need geocoding
func GeocodeAddresses(addressesToGeocode []geocoding.Address, tournamentsToUpdate []int, tournamentCacheEntries []cache.TournamentCache) ([]cache.TournamentCache, int, int) {
	var successCount, failureCount int

	// Process each address
	for i, addrIndex := range tournamentsToUpdate {
		address := addressesToGeocode[i]

		// Skip geocoding if we've already determined this address is invalid
		if !geocoding.IsAddressValid(address) {
			failureCount++
			continue
		}

		// Get geocoding coordinates
		location, err := geocoding.GetCoordinates(address)
		if err != nil {
			log.Printf("Error geocoding address %s, %s %s: %v",
				address.StreetAddress, address.PostalCode, address.AddressLocality, err)
			failureCount++
			continue
		}

		// Update cache entry with geocoding results
		cacheEntry := tournamentCacheEntries[addrIndex]
		cacheEntry.Address.Latitude = location.Lat
		cacheEntry.Address.Longitude = location.Lon
		cacheEntry.Address.Failed = location.Failed

		if location.Failed {
			failureCount++
		} else {
			successCount++
		}

		tournamentCacheEntries[addrIndex] = cacheEntry
	}

	return tournamentCacheEntries, successCount, failureCount
}
