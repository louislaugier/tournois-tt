package tournaments

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"sync"
	"time"
	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/scraper/browser"
	"tournois-tt/api/pkg/scraper/services/helloasso"
	"tournois-tt/api/pkg/utils"

	pw "github.com/playwright-community/playwright-go"
)

func RefreshSignupURLs() {
	_, currentSeasonEnd := utils.GetCurrentSeason()
	if err := refreshSignupURLs(utils.Ptr(time.Now()), &currentSeasonEnd); err != nil {
		log.Printf("Warning: Failed to refresh tournament signup URLs: %v", err)
	}
}

// concurrency control
var numWorkers = 4

func refreshSignupURLs(startDateAfter, startDateBefore *time.Time) error {
	// Load existing tournaments from cache
	cachedTournaments, err := cache.LoadTournaments()
	if err != nil {
		return err
	}

	// Filter tournaments that need processing
	var tournamentsToProcess []cache.TournamentCache
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

		// Add to list of tournaments to process
		tournamentsToProcess = append(tournamentsToProcess, tournament)
	}

	if len(tournamentsToProcess) == 0 {
		log.Printf("No tournaments need signup URL refresh")
		return nil
	}

	log.Printf("Processing %d tournaments for signup URL refresh", len(tournamentsToProcess))

	if len(tournamentsToProcess) < numWorkers {
		numWorkers = len(tournamentsToProcess)
	}

	// Initialize a shared browser instance
	cfg := browser.DefaultConfig()
	browserInstance, pwInstance, err := browser.Init(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize browser: %v", err)
	}
	defer pwInstance.Stop()
	defer browserInstance.Close()

	// Create a browser context that will be shared among all workers
	browserContext, err := browser.NewContext(browserInstance, cfg)
	if err != nil {
		return fmt.Errorf("failed to create browser context: %v", err)
	}
	defer browserContext.Close()

	// Create channels for work distribution and results collection
	tournamentCh := make(chan cache.TournamentCache, len(tournamentsToProcess))
	resultCh := make(chan cache.TournamentCache, len(tournamentsToProcess))
	errorCh := make(chan error, len(tournamentsToProcess))

	// Create a wait group to manage workers
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	// Start worker goroutines
	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			defer wg.Done()
			for tournament := range tournamentCh {
				// Parse the tournament date
				tournamentDate, err := time.Parse("2006-01-02 15:04", tournament.StartDate)
				if err != nil {
					// Try alternative format with T separator
					tournamentDate, err = time.Parse("2006-01-02T15:04:05", tournament.StartDate)
					if err != nil {
						// If still can't parse, try without time
						tournamentDate, err = time.Parse("2006-01-02", tournament.StartDate)
						if err != nil {
							errorCh <- fmt.Errorf("failed to parse date for tournament %s: %v", tournament.Name, err)
							continue
						}
					}
				}

				// Process the tournament using shared browser
				log.Printf("Worker %d: Searching for signup URL for tournament %s (date: %s, club: %s, postal code: %s)",
					workerID, tournament.Name, tournamentDate.Format("2006-01-02"),
					tournament.Club.Name, tournament.Address.PostalCode)

				signupUrl, err := findSignupUrlOnHelloAsso(tournament, tournamentDate, browserContext, pwInstance)
				if err != nil {
					log.Printf("Worker %d: Warning: Failed to find signup URL for tournament %s: %v", workerID, tournament.Name, err)
					errorCh <- err
				} else if signupUrl != "" {
					tournament.SignupUrl = signupUrl
					log.Printf("Worker %d: Found signup URL for tournament %s: %s", workerID, tournament.Name, signupUrl)
				} else {
					log.Printf("Worker %d: No signup URL found for tournament %s", workerID, tournament.Name)
				}

				// Update tournament fields for site and rules PDF checking
				if !tournament.IsSiteExistenceChecked {
					// TODO: Check if site exists and update ClubSiteUrl accordingly
					tournament.IsSiteExistenceChecked = true
				} else if tournament.SiteUrl != "" {
					// TODO: Check website for tournament signup link
				}

				if !tournament.IsRulesPdfChecked && tournament.Rules != nil && tournament.Rules.URL != "" {
					tournament.IsRulesPdfChecked = true
					// TODO: Check PDF for tournament signup link
				}

				// Send result back
				resultCh <- tournament
			}
		}(i)
	}

	// Send tournaments to workers
	for _, tournament := range tournamentsToProcess {
		tournamentCh <- tournament
	}
	close(tournamentCh)

	// Wait for all workers to complete in a separate goroutine
	go func() {
		wg.Wait()
		close(resultCh)
		close(errorCh)
	}()

	// Collect results
	var updatedTournaments []cache.TournamentCache
	for tournament := range resultCh {
		updatedTournaments = append(updatedTournaments, tournament)
	}

	// Check for errors (non-blocking)
	var errors []error
	for err := range errorCh {
		errors = append(errors, err)
	}

	// Save updated tournaments back to cache
	if len(updatedTournaments) > 0 {
		log.Printf("Saving %d updated tournaments to cache", len(updatedTournaments))
		if err := cache.SaveTournamentsToCache(updatedTournaments); err != nil {
			return err
		}
		log.Printf("Successfully saved tournaments with updated signup URLs to cache")
	} else {
		log.Printf("No tournaments were updated with signup URLs")
	}

	if len(errors) > 0 {
		log.Printf("Warning: Encountered %d errors while refreshing signup URLs", len(errors))
	}

	return nil
}

// findSignupUrlOnHelloAsso searches for tournament signup URLs on HelloAsso using a shared browser
func findSignupUrlOnHelloAsso(tournament cache.TournamentCache, tournamentDate time.Time, browserContext pw.BrowserContext, pwInstance *pw.Playwright) (string, error) {
	ctx := context.Background()

	// Get tournament postal code
	tournamentPostalCode := tournament.Address.PostalCode

	// Format the date for potential inclusion in search
	formattedDate := tournamentDate.Format("02/01/2006")
	monthName := getMonthNameFrench(int(tournamentDate.Month()))
	year := tournamentDate.Year()
	day := tournamentDate.Day()

	// Try to find by tournament name with date
	if tournament.Name != "" {
		// First try with full name and exact date
		dateSpecificQuery := fmt.Sprintf("%s %s", tournament.Name, formattedDate)
		log.Printf("Searching HelloAsso for tournament name with date: %s", dateSpecificQuery)
		activities, err := searchHelloAssoAndFilterByDate(ctx, dateSpecificQuery, tournamentDate, tournamentPostalCode, tournament.ID, tournament.Name, browserContext, pwInstance)
		if err == nil && len(activities) > 0 {
			log.Printf("Found %d activities by tournament name with date, using first result: %s (%s)",
				len(activities), activities[0].Title, activities[0].URL)
			return activities[0].URL, nil
		}

		// Then try with just the name
		log.Printf("Searching HelloAsso for tournament name: %s", tournament.Name)
		activities, err = searchHelloAssoAndFilterByDate(ctx, tournament.Name, tournamentDate, tournamentPostalCode, tournament.ID, tournament.Name, browserContext, pwInstance)
		if err == nil && len(activities) > 0 {
			log.Printf("Found %d activities by tournament name, using first result: %s (%s)",
				len(activities), activities[0].Title, activities[0].URL)
			return activities[0].URL, nil
		} else if err != nil {
			log.Printf("Error searching by tournament name: %v", err)
		} else {
			log.Printf("No activities found by tournament name")
		}
	}

	// Try to find by club name with date
	if tournament.Club.Name != "" {
		// First try with club name and exact date
		dateSpecificQuery := fmt.Sprintf("%s %d %s", tournament.Club.Name, day, monthName)
		log.Printf("Searching HelloAsso for club name with date: %s", dateSpecificQuery)
		activities, err := searchHelloAssoAndFilterByDate(ctx, dateSpecificQuery, tournamentDate, tournamentPostalCode, tournament.ID, tournament.Name, browserContext, pwInstance)
		if err == nil && len(activities) > 0 {
			log.Printf("Found %d activities by club name with date, using first result: %s (%s)",
				len(activities), activities[0].Title, activities[0].URL)
			return activities[0].URL, nil
		}

		// Then try with just the club name
		log.Printf("Searching HelloAsso for club name: %s", tournament.Club.Name)
		activities, err = searchHelloAssoAndFilterByDate(ctx, tournament.Club.Name, tournamentDate, tournamentPostalCode, tournament.ID, tournament.Name, browserContext, pwInstance)
		if err == nil && len(activities) > 0 {
			log.Printf("Found %d activities by club name, using first result: %s (%s)",
				len(activities), activities[0].Title, activities[0].URL)
			return activities[0].URL, nil
		} else if err != nil {
			log.Printf("Error searching by club name: %v", err)
		} else {
			log.Printf("No activities found by club name")
		}
	}

	// Try to find by city name and date
	if tournament.Address.AddressLocality != "" {
		// First try with city name, date and "tennis de table" (table tennis)
		dateSpecificQuery := fmt.Sprintf("%s tennis de table %d %s %d",
			tournament.Address.AddressLocality, day, monthName, year)
		log.Printf("Searching HelloAsso for city name with date and sport: %s", dateSpecificQuery)
		activities, err := searchHelloAssoAndFilterByDate(ctx, dateSpecificQuery, tournamentDate, tournamentPostalCode, tournament.ID, tournament.Name, browserContext, pwInstance)
		if err == nil && len(activities) > 0 {
			log.Printf("Found %d activities by city name with date and sport, using first result: %s (%s)",
				len(activities), activities[0].Title, activities[0].URL)
			return activities[0].URL, nil
		}

		// Then try with just the city name
		log.Printf("Searching HelloAsso for city name: %s", tournament.Address.AddressLocality)
		activities, err = searchHelloAssoAndFilterByDate(ctx, tournament.Address.AddressLocality, tournamentDate, tournamentPostalCode, tournament.ID, tournament.Name, browserContext, pwInstance)
		if err == nil && len(activities) > 0 {
			log.Printf("Found %d activities by city name, using first result: %s (%s)",
				len(activities), activities[0].Title, activities[0].URL)
			return activities[0].URL, nil
		} else if err != nil {
			log.Printf("Error searching by city name: %v", err)
		} else {
			log.Printf("No activities found by city name")
		}
	}

	log.Printf("No signup URL found after exhausting all search options")
	return "", nil
}

// getMonthNameFrench returns the French name of the month for the given month number (1-12)
func getMonthNameFrench(month int) string {
	months := []string{
		"janvier", "février", "mars", "avril", "mai", "juin",
		"juillet", "août", "septembre", "octobre", "novembre", "décembre",
	}

	if month < 1 || month > 12 {
		return ""
	}

	return months[month-1]
}

// searchHelloAssoAndFilterByDate searches HelloAsso with the given query and filters results by date using a shared browser
func searchHelloAssoAndFilterByDate(ctx context.Context, query string, targetDate time.Time, tournamentPostalCode string, tournamentID int, tournamentName string, browserContext pw.BrowserContext, pwInstance *pw.Playwright) ([]helloasso.Activity, error) {
	// Search on HelloAsso using the shared browser context
	activities, err := helloasso.SearchActivitiesWithBrowser(ctx, query, browserContext, pwInstance)
	if err != nil {
		return nil, err
	}

	log.Printf("Found %d raw activities for query '%s' (tournament ID: %d, name: %s, target date: %s)",
		len(activities), query, tournamentID, tournamentName, targetDate.Format("2006-01-02"))

	// Debug: List all activities with their fields
	for i, activity := range activities {
		log.Printf("Activity %d: Title='%s', Date='%s', URL='%s', Location='%s'",
			i, activity.Title, activity.Date, activity.URL, activity.Location)
	}

	// Filter results by date, category, postal code and name similarity
	filtered := make([]helloasso.Activity, 0)

	// First, look for title containing "tennis de table" or "tennis", sorting by relevance and date
	relevantActivities := make([]helloasso.Activity, 0)
	otherActivities := make([]helloasso.Activity, 0)

	for _, activity := range activities {
		// Validate required activity fields
		if activity.Title == "" {
			log.Printf("Skipping activity with empty title (tournament ID: %d)", tournamentID)
			continue
		}

		if activity.URL == "" {
			log.Printf("Skipping activity '%s' with empty URL (tournament ID: %d)", activity.Title, tournamentID)
			continue
		}

		// Check for empty date
		if activity.Date == "" {
			log.Printf("Skipping activity '%s' due to empty date field (tournament ID: %d)", activity.Title, tournamentID)
			continue
		}

		titleLower := strings.ToLower(activity.Title)
		if strings.Contains(titleLower, "tennis de table") ||
			strings.Contains(titleLower, "ping") ||
			strings.Contains(titleLower, "tournoi national") {
			relevantActivities = append(relevantActivities, activity)
		} else {
			otherActivities = append(otherActivities, activity)
		}
	}

	// Process relevant activities first, then others
	allActivitiesToProcess := append(relevantActivities, otherActivities...)

	for _, activity := range allActivitiesToProcess {
		// Parse activity date
		activityDate, err := utils.ParseHelloAssoDate(activity.Date)
		if err != nil {
			log.Printf("Skipping activity '%s' due to date parsing error: %v for date '%s' (tournament ID: %d, tournament name: %s)",
				activity.Title, err, activity.Date, tournamentID, tournamentName)
			continue
		}

		// Look for exact date match within +/- 1 day to account for variations
		dateDiff := math.Abs(float64(activityDate.Sub(targetDate).Hours() / 24))
		if dateDiff > 1 {
			log.Printf("Skipping activity due to date mismatch: activity=%s, activityDate=%s, targetDate=%s, diff=%.1f days (tournament ID: %d, tournament name: %s)",
				activity.Title, activityDate.Format("2006-01-02"), targetDate.Format("2006-01-02"), dateDiff, tournamentID, tournamentName)
			continue
		}

		// For tournaments with generic names (less than 3 words), do a more strict title check
		tournamentWords := len(strings.Fields(strings.ToLower(tournamentName)))
		activityTitleLower := strings.ToLower(activity.Title)

		if tournamentWords < 3 && !strings.Contains(activityTitleLower, strings.ToLower(tournamentName)) {
			// For generic tournament names, check if the activity title contains the tournament name
			// If not, it might be an unrelated tournament that happens to match the date
			log.Printf("Skipping activity with non-matching title for generic tournament: activity=%s, tournament=%s (tournament ID: %d)",
				activity.Title, tournamentName, tournamentID)
			continue
		}

		log.Printf("Found activity with matching date: title=%s, date=%s, target=%s (tournament ID: %d, tournament name: %s)",
			activity.Title, activityDate.Format("2006-01-02"), targetDate.Format("2006-01-02"), tournamentID, tournamentName)

		// Check if postal code matches if available
		if tournamentPostalCode != "" && activity.Location != "" {
			activityPostalCode := utils.ExtractPostalCode(activity.Location)

			// Try exact match first
			if activityPostalCode == tournamentPostalCode {
				filtered = append(filtered, activity)
				log.Printf("Added activity with exact postal code match: %s (postal code: %s, tournament ID: %d, tournament name: %s)",
					activity.Title, activityPostalCode, tournamentID, tournamentName)
				continue
			}

			// If no exact match, check if first two digits match (same department)
			if len(activityPostalCode) >= 2 && len(tournamentPostalCode) >= 2 &&
				activityPostalCode[:2] == tournamentPostalCode[:2] {
				filtered = append(filtered, activity)
				log.Printf("Added activity with department match: %s (postal code: %s, tournament postal code: %s, tournament ID: %d, tournament name: %s)",
					activity.Title, activityPostalCode, tournamentPostalCode, tournamentID, tournamentName)
				continue
			}

			log.Printf("Skipping activity due to postal code mismatch: activity=%s, activityPostalCode=%s, tournamentPostalCode=%s (tournament ID: %d, tournament name: %s)",
				activity.Title, activityPostalCode, tournamentPostalCode, tournamentID, tournamentName)
			continue
		}

		// If no postal code to check or postal code not available, add the activity
		filtered = append(filtered, activity)
	}

	log.Printf("Filtered to %d activities matching criteria (tournament ID: %d, tournament name: %s)", len(filtered), tournamentID, tournamentName)
	return filtered, nil
}
