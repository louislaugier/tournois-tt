package signup

import (
	"fmt"
	"log"
	"sync"
	"time"

	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/utils"

	pw "github.com/playwright-community/playwright-go"
)

// RefreshURLs updates signup URLs for tournaments in the specified date range
func RefreshURLs() {
	_, currentSeasonEnd := utils.GetCurrentSeason()
	if err := RefreshURLsInRange(utils.Ptr(time.Now()), &currentSeasonEnd); err != nil {
		log.Fatalf("Critical error in tournament signup URL refresh: %v", err)
	}
}

// RefreshURLsInRange updates signup URLs for tournaments within the specified date range
func RefreshURLsInRange(startDateAfter, startDateBefore *time.Time) error {
	// Check if we're querying for current season
	isCurrentSeason := isCurrentSeasonQuery(startDateAfter, startDateBefore)

	// Configure retry parameters
	maxRetries := 3
	if !isCurrentSeason {
		maxRetries = 1 // Only retry once for historical data
	}
	
	var cachedTournaments map[string]cache.TournamentCache
	var err error
	var attempt int
	
	// Try loading tournaments with retries
	for attempt = 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			// Calculate exponential backoff delay: 5s, 20s, 60s
			delaySeconds := 5 * attempt * attempt
			log.Printf("Tournament cache loading attempt %d/%d failed, retrying in %d seconds...", 
				attempt-1, maxRetries, delaySeconds)
			time.Sleep(time.Duration(delaySeconds) * time.Second)
		}
		
		// Load existing tournaments from cache
		cachedTournaments, err = cache.LoadTournaments()
		if err == nil {
			break // Success, exit retry loop
		}
	}
	
	// If all retries failed and we're in current season, it's a critical error
	if err != nil {
		if isCurrentSeason {
			return fmt.Errorf("critical error: failed to load tournament cache after %d attempts: %w", maxRetries, err)
		}
		return fmt.Errorf("failed to load tournament cache: %w", err)
	}

	// Convert map to slice for processing
	var tournamentsList []cache.TournamentCache
	for _, tournament := range cachedTournaments {
		tournamentsList = append(tournamentsList, tournament)
	}

	// Filter tournaments that need processing
	tournamentsToProcess := filterTournamentsForProcessing(tournamentsList, startDateAfter, startDateBefore)

	// For current season, if no tournaments are found, retry a few times
	// This could happen if the FFTT API refresh just ran and hasn't completed yet
	if isCurrentSeason && len(tournamentsToProcess) == 0 {
		for retryAttempt := 1; retryAttempt <= maxRetries; retryAttempt++ {
			delaySeconds := 30 * retryAttempt // Wait longer (30s, 60s, 90s) as this is waiting for geocoding to complete
			log.Printf("No current season tournaments found, waiting %d seconds for possible FFTT API refresh to complete...", delaySeconds)
			time.Sleep(time.Duration(delaySeconds) * time.Second)
			
			// Try loading again
			cachedTournaments, err = cache.LoadTournaments()
			if err != nil {
				continue
			}
			
			// Convert and filter again
			tournamentsList = []cache.TournamentCache{}
			for _, tournament := range cachedTournaments {
				tournamentsList = append(tournamentsList, tournament)
			}
			tournamentsToProcess = filterTournamentsForProcessing(tournamentsList, startDateAfter, startDateBefore)
			
			if len(tournamentsToProcess) > 0 {
				log.Printf("Found %d tournaments after retry attempt %d", len(tournamentsToProcess), retryAttempt)
				break
			}
		}
	}

	if len(tournamentsToProcess) == 0 {
		if isCurrentSeason {
			// For current season, no tournaments to process is a critical error
			return fmt.Errorf("critical error: no tournaments found to process for current season signup URL refresh after multiple attempts")
		}
		log.Printf("No tournaments need signup URL refresh")
		return nil
	}

	log.Printf("Processing %d tournaments for signup URL refresh", len(tournamentsToProcess))

	// Adjust worker count if needed
	workerCount := numWorkers
	if len(tournamentsToProcess) < workerCount {
		workerCount = len(tournamentsToProcess)
	}

	// Initialize shared browser
	browserInstance, pwInstance, browserContext, err := browserSetup()
	if err != nil {
		// This is a critical error - browser setup failed
		return fmt.Errorf("browser setup failed (critical error): %w", err)
	}
	defer pwInstance.Stop()
	defer browserInstance.Close()
	defer browserContext.Close()

	// Create channels for work distribution and results collection
	tournamentCh := make(chan cache.TournamentCache, len(tournamentsToProcess))
	resultCh := make(chan cache.TournamentCache, len(tournamentsToProcess))
	errorCh := make(chan error, len(tournamentsToProcess))

	// Create a wait group to manage workers
	var wg sync.WaitGroup
	wg.Add(workerCount)

	// Start worker goroutines
	for i := 0; i < workerCount; i++ {
		go processWorker(i, tournamentCh, resultCh, errorCh, &wg, browserContext, pwInstance)
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

	// Collect results (now just for statistics)
	processedTournaments := collectResults(resultCh)

	// Check for errors (non-blocking)
	errors := collectErrors(errorCh)

	// Log statistics
	log.Printf("Processed %d tournaments for signup URLs", len(processedTournaments))
	
	if len(errors) > 0 {
		log.Printf("Warning: Encountered %d errors while refreshing signup URLs", len(errors))
	}

	return nil
}

// isCurrentSeasonQuery checks if the query date range is part of the current season
func isCurrentSeasonQuery(startDateAfter, startDateBefore *time.Time) bool {
	currentSeasonStart, currentSeasonEnd := utils.GetCurrentSeason()
	
	// If startDateAfter is nil, we're starting from before current season
	if startDateAfter == nil {
		return false
	}
	
	// Check if startDateAfter is within or after the current season start
	isWithinCurrentSeason := !startDateAfter.Before(currentSeasonStart)
	
	// Check if startDateBefore is within or equal to current season end (if provided)
	if startDateBefore != nil {
		isWithinCurrentSeason = isWithinCurrentSeason && !startDateBefore.After(currentSeasonEnd)
	}
	
	return isWithinCurrentSeason
}

// filterTournamentsForProcessing filters tournaments that need signup URL refresh
func filterTournamentsForProcessing(tournaments []cache.TournamentCache, startDateAfter, startDateBefore *time.Time) []cache.TournamentCache {
	var result []cache.TournamentCache

	for _, tournament := range tournaments {
		// Parse the tournament date
		tournamentDate, err := parseTournamentDate(tournament.StartDate)
		if err != nil {
			// Tournaments should always have a valid date, so this is an error condition
			log.Fatalf("Critical error: Failed to parse tournament date for tournament %s (ID: %d): %v",
				tournament.Name, tournament.ID, err)
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
		result = append(result, tournament)
	}

	return result
}

// processWorker is a worker function that processes tournaments to find signup URLs
func processWorker(workerID int, tournamentCh <-chan cache.TournamentCache, resultCh chan<- cache.TournamentCache,
	errorCh chan<- error, wg *sync.WaitGroup, browserContext pw.BrowserContext, pwInstance *pw.Playwright) {

	defer wg.Done()

	for tournament := range tournamentCh {
		tournamentModified := false
		
		// Parse the tournament date
		tournamentDate, err := parseTournamentDate(tournament.StartDate)
		if err != nil {
			// Tournaments should always have a valid date, so this is an error condition
			log.Printf("Worker %d: Critical error: Failed to parse tournament date for tournament %s (ID: %d): %v",
				workerID, tournament.Name, tournament.ID, err)
			errorCh <- fmt.Errorf("failed to parse tournament date for tournament %s (ID: %d): %v",
				tournament.Name, tournament.ID, err)
			continue
		}

		// Process the tournament using shared browser
		debugLog("Worker %d: Searching for signup URL for tournament %s (date: %s, club: %s, postal code: %s)",
			workerID, tournament.Name, tournamentDate.Format("2006-01-02"),
			tournament.Club.Name, tournament.Address.PostalCode)

		signupUrl, err := FindSignupURLOnHelloAsso(tournament, tournamentDate, browserContext, pwInstance)
		if err != nil {
			// Any browser error is critical
			log.Printf("Worker %d: Critical error in browser operation: %v", workerID, err)
			errorCh <- fmt.Errorf("critical browser error: %w", err)
			// Exit the worker loop on critical browser errors
			return
		} else if signupUrl != "" {
			tournament.SignupUrl = signupUrl
			tournamentModified = true
			log.Printf("Worker %d: Found signup URL for tournament %s: %s", workerID, tournament.Name, signupUrl)
		} else {
			debugLog("Worker %d: No signup URL found for tournament %s", workerID, tournament.Name)
		}

		// This section implements a fallback strategy to find signup links in tournament rule PDFs
		// when HelloAsso search doesn't yield results. The PDF may contain registration information
		// or direct links to registration platforms. We mark the PDF as checked to avoid redundant processing.
		if tournament.SignupUrl == "" && !tournament.IsRulesPdfChecked && tournament.Rules != nil && tournament.Rules.URL != "" {
			// Check PDF for tournament signup link
			rulesURL := tournament.Rules.URL
			signupUrl, err := ExtractSignupURLFromPDFFile(tournament, tournamentDate, rulesURL, browserContext)
			if err != nil {
				// Browser errors are critical
				log.Printf("Worker %d: Critical error in PDF processing: %v", workerID, err)
				errorCh <- fmt.Errorf("critical browser error in PDF processing: %w", err)
				// Exit the worker on browser errors
				return
			} else if signupUrl != "" {
				tournament.SignupUrl = signupUrl
				tournamentModified = true
				log.Printf("Worker %d: Found signup URL in PDF rules for tournament %s: %s",
					workerID, tournament.Name, signupUrl)
			}
			tournament.IsRulesPdfChecked = true
			tournamentModified = true
		}

		// TODO search on google if nothing found above

		// Save the tournament to cache if it was modified
		if tournamentModified {
			tournamentToSave := []cache.TournamentCache{tournament}
			if err := cache.SaveTournamentsToCache(tournamentToSave); err != nil {
				log.Printf("Worker %d: Warning: Failed to save tournament %d to cache: %v", 
					workerID, tournament.ID, err)
				errorCh <- fmt.Errorf("failed to save tournament %d to cache: %v", tournament.ID, err)
			} else {
				debugLog("Worker %d: Saved tournament %d to cache", workerID, tournament.ID)
			}
		}

		// Send result back to the collector for statistics
		resultCh <- tournament
	}
}

// collectResults collects processed tournaments from the result channel (now just for statistics)
func collectResults(resultCh <-chan cache.TournamentCache) []cache.TournamentCache {
	var processedTournaments []cache.TournamentCache
	for tournament := range resultCh {
		processedTournaments = append(processedTournaments, tournament)
	}
	return processedTournaments
}

// collectErrors collects errors from the error channel
func collectErrors(errorCh <-chan error) []error {
	var errors []error
	for err := range errorCh {
		errors = append(errors, err)
	}
	return errors
}
