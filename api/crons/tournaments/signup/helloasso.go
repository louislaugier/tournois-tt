package signup

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/scraper/services/helloasso"

	pw "github.com/playwright-community/playwright-go"
)

// Debug flag to control verbose logging
var Debug = false

// debugLog logs a message only if Debug is true
func debugLog(format string, args ...interface{}) {
	if Debug {
		log.Printf(format, args...)
	}
}

// FindSignupURLOnHelloAsso searches for signup URL on HelloAsso platform
func FindSignupURLOnHelloAsso(tournament cache.TournamentCache, tournamentDate time.Time, browserContext pw.BrowserContext, pwInstance *pw.Playwright) (string, error) {
	// Try different search strategies in order of likelihood
	searchStrategies := []func(cache.TournamentCache) (string, error){
		buildTournamentNameQuery,        // Strategy 1: Tournament name
		buildTournamentClubTTQuery,      // Strategy 2: "tournoi tennis de table {clubName}"
		buildTournamentClubTTShortQuery, // Strategy 3: "tournoi TT {clubName}"
		buildTournamentCityTTQuery,      // Strategy 4: "tournoi tennis de table {cityName}"
		buildTournamentCityTTShortQuery, // Strategy 5: "tournoi TT {cityName}"
	}

	// Try each search strategy in order
	for i, buildQuery := range searchStrategies {
		searchQuery, err := buildQuery(tournament)
		if err != nil {
			log.Printf("Warning: Failed to build query for strategy %d: %v", i+1, err)
			continue
		}

		debugLog("Trying search strategy %d with query: %s", i+1, searchQuery)

		// Use the helloasso package's search function
		activities, err := helloasso.SearchActivitiesWithBrowser(context.Background(), searchQuery, browserContext, pwInstance)
		if err != nil {
			// Any browser error is critical
			return "", fmt.Errorf("critical browser error in HelloAsso search: %w", err)
		}

		if len(activities) == 0 {
			debugLog("No results found with search strategy %d", i+1)
			continue
		}

		debugLog("Found %d potential activity results on HelloAsso with strategy %d", len(activities), i+1)

		// Extract URLs from the activities
		var activityURLs []string
		for _, activity := range activities {
			if activity.URL != "" && !Contains(activityURLs, activity.URL) {
				activityURLs = append(activityURLs, activity.URL)
			}
		}

		if len(activityURLs) == 0 {
			debugLog("No potential activity URLs found on HelloAsso with strategy %d", i+1)
			continue
		}

		debugLog("Extracted %d unique URLs from HelloAsso search results with strategy %d", len(activityURLs), i+1)

		// Validate each activity URL
		for _, url := range activityURLs {
			debugLog("Validating HelloAsso activity URL: %s", url)
			validURL, err := ValidateSignupURL(url, tournament, tournamentDate, browserContext)
			if err != nil {
				log.Printf("Warning: Failed to validate HelloAsso URL: %v", err)
				continue
			}

			if validURL != "" {
				log.Printf("Found valid signup URL on HelloAsso: %s", validURL)
				return validURL, nil
			}
		}

		debugLog("No valid signup URL found with strategy %d", i+1)
	}

	debugLog("No valid signup URL found on HelloAsso after trying all search strategies")
	return "", nil
}

// buildTournamentNameQuery uses the tournament name for search
func buildTournamentNameQuery(tournament cache.TournamentCache) (string, error) {
	if tournament.Name == "" {
		return "", fmt.Errorf("tournament name is empty")
	}
	return strings.ToLower(tournament.Name), nil
}

// buildTournamentClubTTQuery builds "tournoi tennis de table {clubName}" query
func buildTournamentClubTTQuery(tournament cache.TournamentCache) (string, error) {
	clubName := tournament.Club.Name
	if clubName == "" {
		return "", fmt.Errorf("club name is empty")
	}
	return fmt.Sprintf("tournoi tennis de table %s", strings.ToLower(clubName)), nil
}

// buildTournamentClubTTShortQuery builds "tournoi TT {clubName}" query
func buildTournamentClubTTShortQuery(tournament cache.TournamentCache) (string, error) {
	clubName := tournament.Club.Name
	if clubName == "" {
		return "", fmt.Errorf("club name is empty")
	}
	return fmt.Sprintf("tournoi TT %s", strings.ToLower(clubName)), nil
}

// buildTournamentCityTTQuery builds "tournoi tennis de table {cityName}" query
func buildTournamentCityTTQuery(tournament cache.TournamentCache) (string, error) {
	cityName := tournament.Address.AddressLocality
	if cityName == "" {
		return "", fmt.Errorf("city name is empty")
	}
	return fmt.Sprintf("tournoi tennis de table %s", strings.ToLower(cityName)), nil
}

// buildTournamentCityTTShortQuery builds "tournoi TT {cityName}" query
func buildTournamentCityTTShortQuery(tournament cache.TournamentCache) (string, error) {
	cityName := tournament.Address.AddressLocality
	if cityName == "" {
		return "", fmt.Errorf("city name is empty")
	}
	return fmt.Sprintf("tournoi TT %s", strings.ToLower(cityName)), nil
}

// Legacy function kept for compatibility
func buildHelloAssoSearchQuery(tournament cache.TournamentCache) (string, error) {
	return buildTournamentClubTTQuery(tournament)
}
