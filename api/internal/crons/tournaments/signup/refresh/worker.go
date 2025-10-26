package refresh

import (
	"fmt"
	"log"
	"sync"

	"tournois-tt/api/pkg/cache"

	pw "github.com/playwright-community/playwright-go"
)

// processTournamentBatch processes a batch of tournaments to refresh Page field from FFTT
func processTournamentBatch(tournaments []cache.TournamentCache) ([]cache.TournamentCache, error) {
	if len(tournaments) == 0 {
		return nil, nil
	}

	// No browser needed anymore - just process with nil browser context
	return ProcessWithExistingBrowser(tournaments, nil, nil)
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

// Worker processes tournaments from the input channel, refreshes Page field from FFTT,
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
		// Note: browserContext and pwInstance are no longer used but kept for backward compatibility
		processedTournament, processed, err := ProcessTournament(workerID, tournament, browserContext, pwInstance)

		if err != nil {
			log.Printf("Worker %d: Error processing tournament %s: %v", workerID, tournament.Name, err)
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
