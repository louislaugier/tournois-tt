package refresh

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/helloasso"
	"tournois-tt/api/pkg/scraper/browser"
	"tournois-tt/api/pkg/utils"

	pw "github.com/playwright-community/playwright-go"
)

// processTournament refreshes the signup URL for a single tournament
func processTournament(tournament cache.TournamentCache, resultChan chan<- cache.TournamentCache,
	errChan chan<- error) {

	// Skip tournaments that already have signup URLs
	if tournament.SignupUrl != "" {
		log.Printf("Skipping tournament ID %d (%s) - already has signup URL",
			tournament.ID, tournament.Name)
		return
	}

	log.Printf("Processing tournament ID %d: %s", tournament.ID, tournament.Name)

	// Attempt to fetch signup URL
	signupURL, err := fetchSignupURL(tournament)
	if err != nil {
		errChan <- fmt.Errorf("error processing tournament %d (%s): %w",
			tournament.ID, tournament.Name, err)
		return
	}

	// Update tournament with new signup URL if one was found
	if signupURL != "" {
		tournament.SignupUrl = signupURL
		// Note: We don't set LastRefreshed as it's not in the TournamentCache struct
		resultChan <- tournament
	}
}

// ProcessTournament processes a tournament using an already established browser context
// This function is used by the Worker function in worker.go
func ProcessTournament(workerID int, tournament cache.TournamentCache,
	browserContext pw.BrowserContext, pwInstance *pw.Playwright) (cache.TournamentCache, bool, error) {

	tournamentModified := false

	// Parse the tournament date
	tournamentDate, err := utils.ParseTournamentDate(tournament.StartDate)
	if err != nil {
		return tournament, false, fmt.Errorf("failed to parse tournament date for tournament %s (ID: %d): %w",
			tournament.Name, tournament.ID, err)
	}

	// Process the tournament using shared browser
	utils.DebugLog("Worker %d: Searching for signup URL for tournament %s (date: %s, club: %s, postal code: %s)",
		workerID, tournament.Name, tournamentDate.Format("2006-01-02"),
		tournament.Club.Name, tournament.Address.PostalCode)

	// Try to find signup URL on HelloAsso
	signupURL, err := FindSignupURLOnHelloAsso(tournament, tournamentDate, browserContext, pwInstance)
	if err != nil {
		// Handle navigation errors
		if utils.IsNavigationError(err.Error()) {
			log.Printf("Worker %d: Warning: Navigation error for tournament %s: %v",
				workerID, tournament.Name, err)
			return tournament, false, fmt.Errorf("navigation error for tournament %s: %w", tournament.Name, err)
		}
		return tournament, false, err
	} else if signupURL != "" {
		tournament.SignupUrl = signupURL
		tournamentModified = true
		log.Printf("Worker %d: Found signup URL for tournament %s: %s", workerID, tournament.Name, signupURL)
	} else {
		utils.DebugLog("Worker %d: No signup URL found for tournament %s", workerID, tournament.Name)
	}

	// Check PDF if HelloAsso search didn't yield results
	if tournament.SignupUrl == "" && !tournament.IsRulesPdfChecked && tournament.Rules != nil && tournament.Rules.URL != "" {
		tournament, pdfModified, err := checkPDFForSignupURL(workerID, tournament, tournamentDate, browserContext)
		if err != nil {
			return tournament, tournamentModified || pdfModified, err
		}
		tournamentModified = tournamentModified || pdfModified
	}

	return tournament, tournamentModified, nil
}

// checkPDFForSignupURL searches for signup URLs in tournament rule PDFs
func checkPDFForSignupURL(workerID int, tournament cache.TournamentCache,
	tournamentDate time.Time, browserContext pw.BrowserContext) (cache.TournamentCache, bool, error) {

	// Check PDF for tournament signup link
	rulesURL := tournament.Rules.URL
	signupURL, err := ExtractSignupURLFromPDFFile(tournament, tournamentDate, rulesURL, browserContext)
	if err != nil {
		// Handle PDF navigation errors
		if utils.IsNavigationError(err.Error()) {
			log.Printf("Worker %d: Warning: PDF navigation error for tournament %s: %v",
				workerID, tournament.Name, err)
			// Still mark PDF as checked to prevent retrying on subsequent runs
			tournament.IsRulesPdfChecked = true
			return tournament, true, fmt.Errorf("PDF navigation error for tournament %s: %w", tournament.Name, err)
		}
		return tournament, false, err
	} else if signupURL != "" {
		tournament.SignupUrl = signupURL
		tournament.IsRulesPdfChecked = true
		log.Printf("Worker %d: Found signup URL in PDF rules for tournament %s: %s",
			workerID, tournament.Name, signupURL)
		return tournament, true, nil
	}

	// Mark as checked even if no URL found
	tournament.IsRulesPdfChecked = true
	return tournament, true, nil
}

// fetchSignupURL attempts to retrieve the signup URL for a tournament
func fetchSignupURL(tournament cache.TournamentCache) (string, error) {
	// Extract tournament information
	// Note: We don't have ExtractTournamentNumber in utils, so we'll use a different approach

	// Parse the tournament date
	tournamentDate, err := utils.ParseTournamentDate(tournament.StartDate)
	if err != nil {
		return "", fmt.Errorf("failed to parse tournament date: %w", err)
	}

	// Initialize shared browser
	_, pwInstance, browserContext, err := browser.Setup()
	if err != nil {
		return "", fmt.Errorf("browser setup failed: %w", err)
	}

	// Ensure browser is always closed, even if errors occur
	defer browser.ShutdownBrowser()

	// First try to find signup URL on HelloAsso
	signupURL, err := searchForSignupURL(tournament, tournamentDate, browserContext, pwInstance)
	if err != nil {
		// Handle navigation errors differently than critical errors
		if utils.IsNavigationError(err.Error()) {
			log.Printf("Navigation error while searching for signup URL: %v", err)
			return "", nil
		}
		return "", fmt.Errorf("failed to search for signup URL: %w", err)
	}

	return signupURL, nil
}

// searchForSignupURL attempts to find a signup URL for the tournament
func searchForSignupURL(tournament cache.TournamentCache, tournamentDate time.Time,
	browserContext pw.BrowserContext, pwInstance *pw.Playwright) (string, error) {

	// Try to find signup URL on HelloAsso
	signupURL, err := FindSignupURLOnHelloAsso(tournament, tournamentDate, browserContext, pwInstance)
	if err != nil {
		return "", err
	}

	// If we found a URL on HelloAsso, return it
	if signupURL != "" {
		return signupURL, nil
	}

	// If HelloAsso search didn't yield results, check PDF if available
	if tournament.Rules != nil && tournament.Rules.URL != "" && !tournament.IsRulesPdfChecked {
		pdfURL := tournament.Rules.URL
		signupURLFromPDF, err := ExtractSignupURLFromPDFFile(tournament, tournamentDate, pdfURL, browserContext)
		if err != nil {
			if strings.Contains(err.Error(), "navigation error") {
				// Navigation errors are not critical
				return "", nil
			}
			return "", err
		}

		if signupURLFromPDF != "" {
			return signupURLFromPDF, nil
		}
	}

	// No signup URL found
	return "", nil
}

// FindSignupURLOnHelloAsso searches for signup URL on HelloAsso platform
func FindSignupURLOnHelloAsso(tournament cache.TournamentCache, tournamentDate time.Time,
	browserContext pw.BrowserContext, pwInstance *pw.Playwright) (string, error) {

	log.Printf("Searching for signup URL on HelloAsso for tournament %s", tournament.Name)

	// Build search query for HelloAsso
	searchQuery := ""
	if tournament.Name != "" {
		searchQuery = tournament.Name
	} else if tournament.Club.Name != "" {
		searchQuery = fmt.Sprintf("tournoi tennis de table %s", tournament.Club.Name)
	} else {
		return "", fmt.Errorf("insufficient tournament information for HelloAsso search")
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Extract activities from HelloAsso
	activities, err := helloasso.SearchActivitiesWithBrowser(ctx, searchQuery, browserContext, pwInstance)
	if err != nil {
		return "", fmt.Errorf("failed to search HelloAsso: %w", err)
	}

	// Check each activity URL
	for _, activity := range activities {
		if activity.URL == "" {
			continue
		}

		// Validate the URL using the helloasso validator
		validURL, err := helloasso.ValidateHelloAssoURL(activity.URL, tournament, tournamentDate, browserContext)
		if err == nil && validURL != "" {
			return validURL, nil
		}
	}

	return "", nil
}

// ExtractSignupURLFromPDFFile extracts a signup URL from a tournament rules PDF
// This uses the pdf_processing package to eliminate code duplication
func ExtractSignupURLFromPDFFile(tournament cache.TournamentCache, tournamentDate time.Time,
	pdfURL string, browserContext pw.BrowserContext) (string, error) {

	log.Printf("Extracting signup URL from PDF for tournament %s", tournament.Name)

	// Call the implementation from the pdf_processing package
	// return pdf.ExtractSignupURLFromPDF(tournament, tournamentDate, pdfURL, browserContext)
	return "", nil // TODO
}
