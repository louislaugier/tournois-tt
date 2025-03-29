// Package finder provides navigation services for finding tournament signup URLs
package finder

import (
	"fmt"
	"log"
	"strings"
	"time"

	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/constants"
	"tournois-tt/api/pkg/scraper/browser"
	"tournois-tt/api/pkg/utils"

	pw "github.com/playwright-community/playwright-go"
)

// Default timeout values
const DefaultPageTimeout = 30000 // 30 seconds

// Registration keywords used to detect signup-related content
var registrationKeywords = []string{
	constants.Register, constants.Registers, constants.ToRegister, constants.SignUp,
	"registre", "s'enregistrer",
	constants.SignUpForm, constants.Participate, constants.Registration,
	constants.Engagement, constants.Engagements,
	constants.NextStep, constants.NextStepNoAccent, constants.Next, constants.Continue,
}

// ValidateSignupURL checks if a given URL is likely to be a registration/signup form
// for a tournament matching the provided details
func ValidateSignupURL(urlStr string, tournament cache.TournamentCache, tournamentDate time.Time, browserContext pw.BrowserContext) (string, error) {
	utils.DebugLog("Validating signup URL: %s", urlStr)
	log.Printf("Starting validation of URL: %s (Docker-optimized)", urlStr)

	// Skip invalid or irrelevant URLs
	if utils.IsDomainToSkip(utils.ExtractDomain(urlStr)) {
		return "", fmt.Errorf("domain is in skip list: %s", urlStr)
	}

	// Create a special header for different browser emulation on retry
	alternateBrowserHeaders := map[string]string{
		"User-Agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.5 Safari/605.1.15",
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language": "fr-FR,fr;q=0.9",
		"Accept-Encoding": "gzip, deflate, br",
		"Connection":      "keep-alive",
		"DNT":             "1",
	}

	// DOCKER-OPTIMIZED: Perform a robust check on the browser context
	if browserContext == nil {
		log.Printf("ERROR: Browser context is nil, cannot validate URL")
		return "", fmt.Errorf("browser context is nil")
	}

	// Check if parent browser is accessible and get version info
	if browserContext.Browser() == nil {
		log.Printf("ERROR: Browser instance not accessible from context")
		return "", fmt.Errorf("browser instance not accessible from context")
	}

	log.Printf("DOCKER DIAGNOSTICS: Using browser version: %s", browserContext.Browser().Version())

	// DOCKER-OPTIMIZED: Create a new context specifically for this validation
	// to avoid potential issues with shared contexts
	log.Printf("Creating fresh context for this validation")
	config := browser.DefaultConfig()
	freshContext, err := browser.NewContext(browserContext.Browser(), config)
	if err != nil {
		log.Printf("ERROR: Failed to create fresh context: %v", err)
		return "", fmt.Errorf("failed to create fresh context: %w", err)
	}

	// Ensure we clean up the fresh context regardless of outcome
	defer func() {
		log.Printf("Closing fresh context used for validation")
		if err := freshContext.Close(); err != nil {
			log.Printf("WARNING: Error closing fresh context: %v", err)
		}
	}()

	// Set aggressive timeouts on the fresh context
	freshContext.SetDefaultNavigationTimeout(60000) // 60 seconds
	freshContext.SetDefaultTimeout(60000)           // 60 seconds

	// DOCKER-OPTIMIZED: Create a page with robust error handling
	var page pw.Page
	log.Printf("Creating page from fresh context for validation")

	// Add a page creation timeout
	pageCreationCh := make(chan struct {
		page pw.Page
		err  error
	}, 1)

	go func() {
		newPage, err := freshContext.NewPage()
		pageCreationCh <- struct {
			page pw.Page
			err  error
		}{newPage, err}
	}()

	// Wait for page creation with timeout
	select {
	case result := <-pageCreationCh:
		if result.err != nil {
			log.Printf("ERROR: Failed to create page: %v", result.err)
			return "", fmt.Errorf("failed to create page: %w", result.err)
		}
		page = result.page
		log.Printf("Successfully created page for validation")
	case <-time.After(15 * time.Second):
		log.Printf("ERROR: Page creation timed out after 15 seconds")
		return "", fmt.Errorf("page creation timed out after 15 seconds")
	}

	// Set up deferred cleanup
	defer func() {
		log.Printf("Closing page for validation")
		if err := page.Close(); err != nil {
			log.Printf("WARNING: Error closing page: %v", err)
		}
	}()

	// Attempt the navigation with simplified error handling
	log.Printf("Navigating to URL with simplified error handling: %s", urlStr)
	var resp pw.Response

	// Try to prepend https:// if not specified
	fullURL := urlStr
	if !strings.HasPrefix(urlStr, "http") {
		fullURL = "https://" + urlStr
		log.Printf("Adding https:// prefix: %s", fullURL)
	}

	// Attempt navigation with timeout channel
	navigateCh := make(chan struct {
		resp pw.Response
		err  error
	}, 1)

	go func() {
		resp, err := page.Goto(fullURL, pw.PageGotoOptions{
			WaitUntil: pw.WaitUntilStateDomcontentloaded, // Less strict wait
			Timeout:   pw.Float(45000),                   // 45 seconds - reduced from 60
		})
		navigateCh <- struct {
			resp pw.Response
			err  error
		}{resp, err}
	}()

	// Wait for navigation with timeout
	select {
	case result := <-navigateCh:
		resp = result.resp
		err = result.err
	case <-time.After(50 * time.Second): // Slightly longer than the goto timeout
		log.Printf("ERROR: Navigation timed out after 50 seconds")
		return "", fmt.Errorf("navigation timed out after 50 seconds")
	}

	// If first navigation attempt fails, try with alternate headers
	if err != nil {
		log.Printf("First navigation attempt failed: %v. Trying with alternate browser profile.", err)

		// Try with a different browser profile
		log.Printf("Setting alternate browser headers for retry")
		if err := page.SetExtraHTTPHeaders(alternateBrowserHeaders); err != nil {
			log.Printf("Failed to set alternate browser headers: %v", err)
		}

		// Try navigation again with different browser profile
		log.Printf("Retrying navigation with alternate browser profile")
		resp, err = page.Goto(fullURL, pw.PageGotoOptions{
			WaitUntil: pw.WaitUntilStateLoad, // Even less strict wait
			Timeout:   pw.Float(30000),       // 30 seconds - reduced timeout for retry
		})

		if err != nil {
			log.Printf("Both navigation attempts failed: %v", err)
			return "", fmt.Errorf("both navigation attempts failed: %w", err)
		}
	}

	// Check HTTP status if response is available
	if resp != nil {
		status := resp.Status()
		log.Printf("HTTP status code: %d", status)
		if status >= 400 {
			log.Printf("HTTP error status %d for URL: %s", status, urlStr)
			// If we got a 403 Forbidden, it's likely anti-crawling protection
			if status == 403 {
				return "", fmt.Errorf("access forbidden (403) - site likely has anti-crawling protection: %s", fullURL)
			}
			return "", fmt.Errorf("HTTP error status %d for URL: %s", status, fullURL)
		}
	}

	// DOCKER-OPTIMIZED: Simplified URL and content check without JS
	currentURL := page.URL()
	log.Printf("Successfully navigated to: %s", currentURL)

	// Check URL patterns first (fastest check)
	urlLower := strings.ToLower(currentURL)
	if strings.Contains(urlLower, "inscription") || strings.Contains(urlLower, "register") ||
		strings.Contains(urlLower, "signup") || strings.Contains(urlLower, "enroll") ||
		strings.Contains(urlLower, "registration") {
		log.Printf("URL contains registration keyword: %s", currentURL)
		return currentURL, nil
	}

	// Get page content directly with a timeout
	log.Printf("Getting page content for analysis")
	contentCh := make(chan struct {
		content string
		err     error
	}, 1)

	go func() {
		content, err := page.Content()
		contentCh <- struct {
			content string
			err     error
		}{content, err}
	}()

	var content string
	select {
	case result := <-contentCh:
		if result.err != nil {
			log.Printf("ERROR: Failed to get page content: %v", result.err)
			return "", fmt.Errorf("failed to get page content: %w", result.err)
		}
		content = result.content
		log.Printf("Successfully got page content (%d bytes)", len(content))
	case <-time.After(20 * time.Second):
		log.Printf("ERROR: Content retrieval timed out after 20 seconds")
		return "", fmt.Errorf("content retrieval timed out")
	}

	// Get page title with a timeout
	log.Printf("Getting page title")
	titleCh := make(chan struct {
		title string
		err   error
	}, 1)

	go func() {
		title, err := page.Title()
		titleCh <- struct {
			title string
			err   error
		}{title, err}
	}()

	var title string
	select {
	case result := <-titleCh:
		if result.err != nil {
			log.Printf("Warning: Could not get page title: %v", result.err)
			title = "" // Use empty string if title can't be retrieved
		} else {
			title = result.title
			log.Printf("Page title: %s", title)
		}
	case <-time.After(5 * time.Second):
		log.Printf("WARNING: Title retrieval timed out, using empty title")
		title = ""
	}

	// Basic form element check
	hasFormElements := strings.Contains(content, "<form") ||
		strings.Contains(content, "<input") ||
		strings.Contains(content, "type=\"submit\"") ||
		strings.Contains(content, "type=\"text\"") ||
		strings.Contains(content, "type=\"email\"") ||
		strings.Contains(content, "type=\"password\"")

	// Quick check for tournament name and registration keywords
	contentLower := strings.ToLower(content)
	titleLower := strings.ToLower(title)
	tournamentNameLower := strings.ToLower(tournament.Name)

	tournamentMatch := strings.Contains(contentLower, tournamentNameLower) ||
		strings.Contains(titleLower, tournamentNameLower)

	registrationMatch := false
	for _, keyword := range registrationKeywords {
		if strings.Contains(titleLower, keyword) || strings.Contains(contentLower, keyword) {
			registrationMatch = true
			break
		}
	}

	// If it has form elements and mentions the tournament or registration keywords
	if hasFormElements && (tournamentMatch || registrationMatch) {
		log.Printf("Found likely registration form at: %s", currentURL)
		return currentURL, nil
	}

	// Print the HTML content for debugging
	log.Printf("DEBUG - HTML content of page that failed registration form validation:\n%s", content)

	log.Printf("URL does not appear to be a registration form: %s", fullURL)
	return "", fmt.Errorf("URL does not appear to be a registration form")
}
