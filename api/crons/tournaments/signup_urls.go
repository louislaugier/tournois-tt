package tournaments

import (
	"context"
	"log"
	"time"
	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/scraper/services/helloasso"
	"tournois-tt/api/pkg/utils"
)

func RefreshSignupURLs() {
	currentSeasonStart, currentSeasonEnd := utils.GetCurrentSeason()
	if err := refreshSignupURLs(&currentSeasonStart, &currentSeasonEnd); err != nil {
		log.Printf("Warning: Failed to refresh tournament signup URLs: %v", err)
	}
}

func refreshSignupURLs(startDateAfter, startDateBefore *time.Time) error {
	// Load existing tournaments from cache
	cachedTournaments, err := cache.LoadTournaments()
	if err != nil {
		return err
	}

	// Create a slice to hold updated tournaments
	updatedTournaments := make([]cache.TournamentCache, 0)

	// For each tournament in the cache
	for _, tournament := range cachedTournaments {
		// Skip tournaments outside our date range
		tournamentDate, err := time.Parse("2006-01-02 15:04", tournament.StartDate)
		if err != nil {
			// Try alternative format with T separator
			tournamentDate, err = time.Parse("2006-01-02T15:04:05", tournament.StartDate)
			if err != nil {
				// If still can't parse, try without time
				tournamentDate, err = time.Parse("2006-01-02", tournament.StartDate)
				if err != nil {
					// Skip this tournament if we can't parse the date
					continue
				}
			}
		}

		// Skip tournaments outside our date range
		if startDateAfter != nil && tournamentDate.Before(*startDateAfter) {
			continue
		}
		if startDateBefore != nil && tournamentDate.After(*startDateBefore) {
			continue
		}

		// Skip tournaments that already have signup URLs
		if tournament.SignupUrl != "" {
			continue
		}

		// check on helloasso with (tournament title / tournament club name / tournament city name) if any results found, only pick results whose dates match current seasons dates
		signupUrl, err := findSignupUrlOnHelloAsso(tournament, tournamentDate)
		if err != nil {
			log.Printf("Warning: Failed to find signup URL for tournament %s: %v", tournament.Name, err)
		} else if signupUrl != "" {
			tournament.SignupUrl = signupUrl
			log.Printf("Found signup URL for tournament %s: %s", tournament.Name, signupUrl)
		}

		// Update tournament fields for site and rules PDF checking
		if !tournament.IsSiteExistenceChecked {
			// TODO: Check if site exists and update ClubSiteUrl accordingly
			tournament.IsSiteExistenceChecked = true
		}

		if !tournament.IsRulesPdfChecked && tournament.Rules != nil && tournament.Rules.URL != "" {
			tournament.IsRulesPdfChecked = true
			// TODO: Check PDF for signup link
		}

		// Add to the list of updated tournaments
		updatedTournaments = append(updatedTournaments, tournament)
	}

	// Save updated tournaments back to cache
	if len(updatedTournaments) > 0 {
		log.Printf("Saving %d updated tournaments to cache", len(updatedTournaments))
		if err := cache.SaveTournamentsToCache(updatedTournaments); err != nil {
			return err
		}
	}

	return nil
}

// findSignupUrlOnHelloAsso searches for tournament signup URLs on HelloAsso
func findSignupUrlOnHelloAsso(tournament cache.TournamentCache, tournamentDate time.Time) (string, error) {
	ctx := context.Background()

	// Get tournament postal code
	tournamentPostalCode := tournament.Address.PostalCode

	// Try to find by tournament name
	if tournament.Name != "" {
		activities, err := searchHelloAssoAndFilterByDate(ctx, tournament.Name, tournamentDate, tournamentPostalCode)
		if err == nil && len(activities) > 0 {
			return activities[0].URL, nil
		}
	}

	// Try to find by club name
	if tournament.Club.Name != "" {
		activities, err := searchHelloAssoAndFilterByDate(ctx, tournament.Club.Name, tournamentDate, tournamentPostalCode)
		if err == nil && len(activities) > 0 {
			return activities[0].URL, nil
		}
	}

	// Try to find by city name
	if tournament.Address.AddressLocality != "" {
		activities, err := searchHelloAssoAndFilterByDate(ctx, tournament.Address.AddressLocality, tournamentDate, tournamentPostalCode)
		if err == nil && len(activities) > 0 {
			return activities[0].URL, nil
		}
	}

	return "", nil
}

// searchHelloAssoAndFilterByDate searches HelloAsso with the given query and filters results by date
func searchHelloAssoAndFilterByDate(ctx context.Context, query string, targetDate time.Time, tournamentPostalCode string) ([]helloasso.Activity, error) {
	// Search on HelloAsso
	activities, err := helloasso.SearchActivities(ctx, query)
	if err != nil {
		return nil, err
	}

	// Filter results by date, category, and postal code
	filtered := make([]helloasso.Activity, 0)
	for _, activity := range activities {
		// Check if category contains "tennis de table" (case insensitive)
		// if !strings.Contains(strings.ToLower(activity.Category), "tennis de table") {
		// 	continue
		// }

		// Parse activity date
		activityDate, err := utils.ParseHelloAssoDate(activity.Date)
		if err != nil {
			continue
		}

		// Check if date is close to target date (within 3 days)
		if !utils.IsDateCloseEnough(targetDate, activityDate, 3) {
			continue
		}

		// Check if postal code matches if available
		if tournamentPostalCode != "" && activity.Location != "" {
			activityPostalCode := utils.ExtractPostalCode(activity.Location)
			if activityPostalCode == "" || activityPostalCode != tournamentPostalCode {
				log.Printf("Skipping activity due to postal code mismatch: activity=%s, location=%s, expectedPostalCode=%s, foundPostalCode=%s",
					activity.Title, activity.Location, tournamentPostalCode, activityPostalCode)
				continue
			}
		}

		filtered = append(filtered, activity)
	}

	return filtered, nil
}
