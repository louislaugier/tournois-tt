package scripts

import (
	"log"
	"time"

	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/geocoding"
)

func RegeocodeFailedTournaments() {
	// Initialize cache
	if err := cache.InitCache(); err != nil {
		log.Fatalf("Failed to initialize cache: %v", err)
	}

	// Load all tournaments
	tournaments, err := cache.LoadTournaments()
	if err != nil {
		log.Fatalf("Failed to load tournaments: %v", err)
	}

	// Find tournaments with failed geocoding
	var failedTournaments []cache.TournamentCache
	for _, tournament := range tournaments {
		if tournament.Address.Failed || (tournament.Address.Latitude == 0 && tournament.Address.Longitude == 0) {
			failedTournaments = append(failedTournaments, tournament)
		}
	}

	if len(failedTournaments) == 0 {
		log.Println("No failed tournaments found to re-geocode")
		return
	}

	log.Printf("Found %d tournaments with failed geocoding", len(failedTournaments))

	// Try to geocode each failed tournament
	successCount := 0
	for i, tournament := range failedTournaments {
		log.Printf("Processing tournament %d/%d: %s", i+1, len(failedTournaments), tournament.Name)

		// Try geocoding with both Nominatim and Google Maps fallback
		location, err := geocoding.GetCoordinates(tournament.Address)
		if err != nil {
			log.Printf("Geocoding failed for tournament %s: %v", tournament.Name, err)
			continue
		}

		// Update tournament with new coordinates
		tournament.Address.Latitude = location.Lat
		tournament.Address.Longitude = location.Lon
		tournament.Address.Failed = location.Failed
		tournament.Timestamp = time.Now()

		// Save updated tournament to cache
		cache.SetCachedTournament(tournament)
		successCount++

		// Add a small delay to avoid rate limiting
		time.Sleep(1 * time.Second)
	}

	// Save all tournaments to the cache file
	// This ensures all tournaments are properly saved, not just the updated ones
	allTournamentsMap, err := cache.LoadTournaments()
	if err != nil {
		log.Printf("Warning: Failed to load tournaments for saving: %v", err)
		return
	}

	allTournaments := []cache.TournamentCache{}
	for _, t := range allTournamentsMap {
		allTournaments = append(allTournaments, t)
	}

	if err := cache.SaveTournamentsToCache(allTournaments); err != nil {
		log.Printf("Warning: Failed to save all tournaments to cache: %v", err)
	} else {
		log.Printf("Successfully saved all %d tournaments to cache", len(allTournaments))
	}

	log.Printf("Re-geocoding completed: %d successful, %d failed", successCount, len(failedTournaments)-successCount)
}
