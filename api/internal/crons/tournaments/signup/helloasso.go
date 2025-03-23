package signup

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/helloasso"
	"tournois-tt/api/pkg/utils"

	pw "github.com/playwright-community/playwright-go"
)

// FindSignupURLOnHelloAsso searches for signup URL on HelloAsso platform
func FindSignupURLOnHelloAsso(tournament cache.TournamentCache, tournamentDate time.Time, browserContext pw.BrowserContext, pwInstance *pw.Playwright) (string, error) {
	// Maximum number of retry attempts for navigation errors
	const maxNavigationRetries = 3
	// Delay between retries (increases with each retry)
	var retryDelayBase = 2 * time.Second

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

		utils.DebugLog("Trying search strategy %d with query: %s", i+1, searchQuery)

		// Use the helloasso package's search function with retries for navigation errors
		var activities []helloasso.Activity
		var searchErr error
		var attemptsMade int

		for attemptsMade = 0; attemptsMade < maxNavigationRetries; attemptsMade++ {
			// Exponential backoff on retries
			if attemptsMade > 0 {
				retryDelay := retryDelayBase * time.Duration(attemptsMade)
				log.Printf("Navigation error in HelloAsso search, retrying in %v (attempt %d/%d)",
					retryDelay, attemptsMade+1, maxNavigationRetries)
				time.Sleep(retryDelay)
			}

			activities, searchErr = helloasso.SearchActivitiesWithBrowser(context.Background(), searchQuery, browserContext, pwInstance)

			// If no error or not a navigation error, break the retry loop
			if searchErr == nil || !utils.IsNavigationError(searchErr.Error()) {
				break
			}
		}

		if searchErr != nil {
			// If we've exhausted retries or it's not a navigation error, propagate the error
			if attemptsMade >= maxNavigationRetries || !utils.IsNavigationError(searchErr.Error()) {
				return "", fmt.Errorf("critical browser error in HelloAsso search: %w", searchErr)
			}
		}

		if len(activities) == 0 {
			utils.DebugLog("No results found with search strategy %d", i+1)
			continue
		}

		utils.DebugLog("Found %d potential activity results on HelloAsso with strategy %d", len(activities), i+1)

		// Extract URLs from the activities
		var activityURLs []string
		for _, activity := range activities {
			if activity.URL != "" && !utils.StringSliceContains(activityURLs, activity.URL) {
				activityURLs = append(activityURLs, activity.URL)
			}
		}

		if len(activityURLs) == 0 {
			utils.DebugLog("No potential activity URLs found on HelloAsso with strategy %d", i+1)
			continue
		}

		utils.DebugLog("Extracted %d unique URLs from HelloAsso search results with strategy %d", len(activityURLs), i+1)

		// Validate each activity URL with retries for navigation errors
		for _, url := range activityURLs {
			utils.DebugLog("Validating HelloAsso activity URL: %s", url)

			var validURL string
			var validationErr error
			var validationAttempts int

			for validationAttempts = 0; validationAttempts < maxNavigationRetries; validationAttempts++ {
				// Exponential backoff on retries
				if validationAttempts > 0 {
					retryDelay := retryDelayBase * time.Duration(validationAttempts)
					log.Printf("Navigation error in HelloAsso URL validation, retrying in %v (attempt %d/%d)",
						retryDelay, validationAttempts+1, maxNavigationRetries)
					time.Sleep(retryDelay)
				}

				validURL, validationErr = ValidateSignupURL(url, tournament, tournamentDate, browserContext)

				// If no error or not a navigation error, break the retry loop
				if validationErr == nil || !utils.IsNavigationError(validationErr.Error()) {
					break
				}
			}

			if validationErr != nil {
				// Non-navigation errors or exhausted retries are propagated
				if !utils.IsNavigationError(validationErr.Error()) || validationAttempts >= maxNavigationRetries {
					log.Printf("Warning: Failed to validate HelloAsso URL: %v", validationErr)
					continue
				}
			}

			if validURL != "" {
				log.Printf("Found valid signup URL on HelloAsso: %s", validURL)
				return validURL, nil
			}
		}

		utils.DebugLog("No valid signup URL found with strategy %d", i+1)
	}

	utils.DebugLog("No valid signup URL found on HelloAsso after trying all search strategies")
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
