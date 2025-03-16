package tournaments

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/fftt"
	"tournois-tt/api/pkg/geocoding"
	"tournois-tt/api/pkg/utils"
)

func RefreshGeocoding() {
	lastSeasonStart, _ := utils.GetLastFinishedSeason()
	if err := refreshGeocoding(&lastSeasonStart, nil); err != nil {
		log.Printf("Warning: Failed to refresh tournament geocoding data: %v", err)
	}
}

// refresh fetches and processes tournament addresses
func refreshGeocoding(startDateAfter, startDateBefore *time.Time) error {
	// Load existing cache
	cachedTournaments, err := cache.LoadTournaments()
	if err != nil {
		log.Printf("Warning: Failed to load tournament cache: %v", err)
	}

	// Create query params for future tournaments
	queryParams := url.Values{}
	queryParams.Set("startDate[after]", startDateAfter.Format("2006-01-02T15:04:05"))
	if startDateBefore != nil {
		queryParams.Set("startDate[before]", startDateBefore.Format("2006-01-02T15:04:05"))
	}
	queryParams.Set("itemsPerPage", "999999")
	queryParams.Set("order[startDate]", "asc")

	// Call FFTT API
	resp, err := fftt.GetClient().GetTournaments(queryParams)
	if err != nil {
		return fmt.Errorf("failed to fetch tournaments: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("FFTT API returned status %d", resp.StatusCode)
	}

	var tournaments []fftt.Tournament
	if err := json.NewDecoder(resp.Body).Decode(&tournaments); err != nil {
		return fmt.Errorf("failed to decode tournaments: %v", err)
	}

	// Prepare addresses for geocoding
	addressesToGeocode := make([]geocoding.Address, 0)
	newTournamentCacheEntries := make([]cache.TournamentCache, 0, len(tournaments))
	successCount := 0
	failureCount := 0

	for _, t := range tournaments {
		// Check if this tournament is already in cache by ID
		cacheKey := fmt.Sprintf("%d", t.ID)
		cachedTournament, exists := cachedTournaments[cacheKey]

		if exists {
			// Use existing cached tournament data but update other fields if needed
			newCacheEntry := cachedTournament
			// TODO: Update other fields if necessary

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

		// Check if the address is valid for geocoding
		if !t.Address.IsValid() {
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
			addressesToGeocode = append(addressesToGeocode, t.Address)
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

		// Geocode individual address
		location, err := geocoding.GetCoordinates(addr)

		// Find the tournament with this address in our new cache entries
		for j := range newTournamentCacheEntries {
			cacheAddr := newTournamentCacheEntries[j].Address
			if cacheAddr.StreetAddress == addr.StreetAddress &&
				cacheAddr.PostalCode == addr.PostalCode &&
				cacheAddr.AddressLocality == addr.AddressLocality {

				if err != nil || location.Failed {
					// Geocoding failed
					newTournamentCacheEntries[j].Address.Failed = true
					failureCount++
				} else {
					// Geocoding succeeded
					newTournamentCacheEntries[j].Address.Latitude = location.Lat
					newTournamentCacheEntries[j].Address.Longitude = location.Lon
					newTournamentCacheEntries[j].Address.Failed = false
					successCount++

					// Log successful geocoding
					log.Printf("Geocoded new address (%d/%d): %s -> (%f, %f)",
						i+1, len(addressesToGeocode),
						geocoding.ConstructFullAddress(addr), location.Lat, location.Lon)
				}
				break
			}
		}
	}

	// Save all tournament cache entries
	if len(newTournamentCacheEntries) > 0 {
		if err := cache.SaveTournamentsToCache(newTournamentCacheEntries); err != nil {
			log.Printf("Warning: Failed to save tournament cache: %v", err)
		}
	}

	log.Printf("Geocoding refresh completed: %d successful, %d failed", successCount, failureCount)
	return nil
}
