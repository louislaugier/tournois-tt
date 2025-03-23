package refresh

import (
	"fmt"
	"log"
	"sync"

	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/scraper/browser"
	"tournois-tt/api/pkg/utils"

	pw "github.com/playwright-community/playwright-go"
)

// processTournamentBatch processes a batch of tournaments using a worker pool to refresh signup URLs
func processTournamentBatch(tournaments []cache.TournamentCache) ([]cache.TournamentCache, error) {
	if len(tournaments) == 0 {
		return nil, nil
	}

	// Create a channel to communicate tournaments to workers
	tournamentChan := make(chan cache.TournamentCache, len(tournaments))
	resultChan := make(chan cache.TournamentCache, len(tournaments))
	errorChan := make(chan error, len(tournaments))

	// Start worker pool
	var wg sync.WaitGroup
	for i := 0; i < NumWorkers; i++ {
		wg.Add(1)
		go worker(tournamentChan, resultChan, errorChan, &wg)
	}

	// Send tournaments to process
	for _, tournament := range tournaments {
		tournamentChan <- tournament
	}
	close(tournamentChan)

	// Wait for all workers to finish
	wg.Wait()
	close(resultChan)
	close(errorChan)

	// Check for errors
	errCount := 0
	for err := range errorChan {
		errCount++
		log.Printf("Error refreshing tournament signup URL: %v", err)
	}

	// Collect results
	var processedTournaments []cache.TournamentCache
	for result := range resultChan {
		processedTournaments = append(processedTournaments, result)
	}

	log.Printf("Processed %d/%d tournaments, encountered %d errors",
		len(processedTournaments), len(tournaments), errCount)

	return processedTournaments, nil
}

// ProcessWithExistingBrowser processes a batch of tournaments using an existing browser context
// This is useful when the browser is already initialized elsewhere
func ProcessWithExistingBrowser(tournaments []cache.TournamentCache, browserContext pw.BrowserContext,
	pwInstance *pw.Playwright) ([]cache.TournamentCache, error) {

	if len(tournaments) == 0 {
		return nil, nil
	}

	// Adjust worker count if needed
	workerCount := NumWorkers
	if len(tournaments) < workerCount {
		workerCount = len(tournaments)
	}

	// Create channels for work distribution and results collection
	tournamentCh := make(chan cache.TournamentCache, len(tournaments))
	resultCh := make(chan cache.TournamentCache, len(tournaments))
	errorCh := make(chan error, len(tournaments))

	// Process tournaments using workers with the provided browser
	results, errors := RunWorkers(workerCount, tournaments, tournamentCh, resultCh, errorCh, browserContext, pwInstance)

	// Log statistics
	log.Printf("Processed %d tournaments for signup URLs", len(results))

	if len(errors) > 0 {
		log.Printf("Warning: Encountered %d errors while refreshing signup URLs", len(errors))
	}

	return results, nil
}

// RunWorkers starts worker goroutines and collects results
func RunWorkers(workerCount int, tournamentsToProcess []cache.TournamentCache,
	tournamentCh chan cache.TournamentCache, resultCh chan cache.TournamentCache,
	errorCh chan error, browserContext pw.BrowserContext, pwInstance *pw.Playwright) ([]cache.TournamentCache, []error) {

	// Create a wait group to manage workers
	var wg sync.WaitGroup
	wg.Add(workerCount)

	// Start worker goroutines
	for i := 0; i < workerCount; i++ {
		go Worker(i, tournamentCh, resultCh, errorCh, &wg, browserContext, pwInstance)
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

	// Collect results and errors
	processedTournaments := CollectResults(resultCh)
	errors := CollectErrors(errorCh)

	return processedTournaments, errors
}

// Worker processes tournaments from the input channel, refreshes signup URLs,
// and sends results to the result channel
func Worker(workerID int, tournamentCh <-chan cache.TournamentCache, resultCh chan<- cache.TournamentCache,
	errorCh chan<- error, wg *sync.WaitGroup, browserContext pw.BrowserContext, pwInstance *pw.Playwright) {

	defer wg.Done()

	// Add panic recovery to prevent goroutine crashes from affecting the whole process
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Worker %d: Recovered from panic: %v", workerID, r)
			errorCh <- fmt.Errorf("worker %d recovered from panic: %v", workerID, r)
		}
	}()

	// Collect modified tournaments locally to save all at once
	var modifiedTournaments []cache.TournamentCache

	for tournament := range tournamentCh {
		processedTournament, processed, err := ProcessTournament(workerID, tournament, browserContext, pwInstance)

		if err != nil {
			// Check if this is a critical browser error
			if !IsNavigationError(err) {
				// This is likely a critical browser error - terminate worker
				log.Printf("Worker %d: Critical error in browser operation: %v", workerID, err)
				errorCh <- fmt.Errorf("critical browser error: %w", err)

				// Try to recover the browser if possible
				if restarted, recoverErr := browser.RestartIfUnhealthy(); recoverErr != nil {
					log.Printf("Worker %d: Failed to recover browser: %v", workerID, recoverErr)
					// Exit the worker loop on critical browser errors
					return
				} else if restarted {
					log.Printf("Worker %d: Browser successfully restarted, continuing operation", workerID)
					// We could continue here, but it's safer to exit and let the next run handle remaining tournaments
					return
				}

				// Exit the worker loop on critical browser errors
				return
			}

			// For navigation errors, report but continue processing
			errorCh <- err
		}

		if processed {
			modifiedTournaments = append(modifiedTournaments, processedTournament)
		}

		// Send result back to the collector for statistics
		resultCh <- processedTournament
	}

	// Save all modified tournaments at once
	SaveModifiedTournaments(workerID, modifiedTournaments, errorCh)
}

// worker is the internal worker function called by the processTournamentBatch method
func worker(input <-chan cache.TournamentCache, output chan<- cache.TournamentCache,
	errChan chan<- error, wg *sync.WaitGroup) {

	defer wg.Done()

	for tournament := range input {
		processTournament(tournament, output, errChan)
	}
}

// CollectResults collects processed tournaments from the result channel
func CollectResults(resultCh <-chan cache.TournamentCache) []cache.TournamentCache {
	var processedTournaments []cache.TournamentCache
	for tournament := range resultCh {
		processedTournaments = append(processedTournaments, tournament)
	}
	return processedTournaments
}

// CollectErrors collects errors from the error channel
func CollectErrors(errorCh <-chan error) []error {
	var errors []error
	for err := range errorCh {
		errors = append(errors, err)
	}
	return errors
}

// SaveModifiedTournaments saves a batch of modified tournaments to cache
func SaveModifiedTournaments(workerID int, modifiedTournaments []cache.TournamentCache, errorCh chan<- error) {
	if len(modifiedTournaments) > 0 {
		if err := cache.SaveTournamentsToCache(modifiedTournaments); err != nil {
			log.Printf("Worker %d: Warning: Failed to save %d modified tournaments to cache: %v",
				workerID, len(modifiedTournaments), err)
			errorCh <- fmt.Errorf("failed to save %d modified tournaments to cache: %v",
				len(modifiedTournaments), err)
		} else {
			log.Printf("Worker %d: Saved %d modified tournaments to cache", workerID, len(modifiedTournaments))
		}
	}
}

// IsNavigationError determines if an error is a navigation error (timeout, network issue)
// that can be skipped rather than a critical browser error that requires termination
func IsNavigationError(err error) bool {
	if err == nil {
		return false
	}
	return utils.IsNavigationError(err.Error())
}
