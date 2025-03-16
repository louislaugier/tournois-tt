package tournaments

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/scraper/browser"
	"tournois-tt/api/pkg/scraper/services/helloasso"
	"tournois-tt/api/pkg/utils"

	pw "github.com/playwright-community/playwright-go"
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

	// Set up concurrency controls
	numWorkers := 4
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
				signupUrl, err := findSignupUrlOnHelloAsso(tournament, tournamentDate, browserContext, pwInstance)
				if err != nil {
					log.Printf("Worker %d: Warning: Failed to find signup URL for tournament %s: %v", workerID, tournament.Name, err)
					errorCh <- err
				} else if signupUrl != "" {
					tournament.SignupUrl = signupUrl
					log.Printf("Worker %d: Found signup URL for tournament %s: %s", workerID, tournament.Name, signupUrl)
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

	// Try to find by tournament name
	if tournament.Name != "" {
		activities, err := searchHelloAssoAndFilterByDate(ctx, tournament.Name, tournamentDate, tournamentPostalCode, browserContext, pwInstance)
		if err == nil && len(activities) > 0 {
			return activities[0].URL, nil
		}
	}

	// Try to find by club name
	if tournament.Club.Name != "" {
		activities, err := searchHelloAssoAndFilterByDate(ctx, tournament.Club.Name, tournamentDate, tournamentPostalCode, browserContext, pwInstance)
		if err == nil && len(activities) > 0 {
			return activities[0].URL, nil
		}
	}

	// Try to find by city name
	if tournament.Address.AddressLocality != "" {
		activities, err := searchHelloAssoAndFilterByDate(ctx, tournament.Address.AddressLocality, tournamentDate, tournamentPostalCode, browserContext, pwInstance)
		if err == nil && len(activities) > 0 {
			return activities[0].URL, nil
		}
	}

	return "", nil
}

// searchHelloAssoAndFilterByDate searches HelloAsso with the given query and filters results by date using a shared browser
func searchHelloAssoAndFilterByDate(ctx context.Context, query string, targetDate time.Time, tournamentPostalCode string, browserContext pw.BrowserContext, pwInstance *pw.Playwright) ([]helloasso.Activity, error) {
	// Search on HelloAsso using the shared browser context
	activities, err := helloasso.SearchActivitiesWithBrowser(ctx, query, browserContext, pwInstance)
	if err != nil {
		return nil, err
	}

	// Filter results by date, category, and postal code
	filtered := make([]helloasso.Activity, 0)
	for _, activity := range activities {
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
