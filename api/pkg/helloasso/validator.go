package helloasso

import (
	"fmt"
	"log"
	"strings"
	"time"

	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/scraper/browser"
	"tournois-tt/api/pkg/scraper/page"
	"tournois-tt/api/pkg/utils"

	pw "github.com/playwright-community/playwright-go"
)

// -----------------------------------------------------------------------------
// URL Validation Functions
// -----------------------------------------------------------------------------

// ValidateURL checks if a URL is valid for tournament registration
func ValidateURL(url string, tournament cache.TournamentCache, tournamentDate time.Time, browserContext pw.BrowserContext) (string, error) {
	// Create a new Playwright page
	pwPage, err := browser.NewPage(browserContext)
	if err != nil {
		return "", fmt.Errorf("critical error - failed to create page: %w", err)
	}
	defer browser.SafeClose(pwPage)

	// Create a page handler
	pageHandler := page.NewPageHandler(pwPage)
	defer pageHandler.Close()

	// Set default timeouts
	pageHandler.SetDefaultTimeouts(30*time.Second, 15*time.Second)

	// Use safer navigation with automatic recovery
	if err := pageHandler.SafeNavigation(url, 2, nil); err != nil {
		return "", fmt.Errorf("critical error - failed to navigate to URL: %w", err)
	}

	// Get page content using SafeOperation to ensure the page is healthy
	var content string
	err = pageHandler.SafeOperation("get content", func() error {
		var err error
		content, err = pageHandler.GetContent()
		return err
	})

	if err != nil {
		return "", fmt.Errorf("critical error - failed to get page content: %w", err)
	}

	// Check if the URL is a valid HelloAsso event that matches our tournament
	contentLower := strings.ToLower(content)
	if isValidPage(contentLower, tournament, tournamentDate) {
		// Take a screenshot for debugging if needed
		if utils.IsDebugMode() {
			screenshotPath := fmt.Sprintf("helloasso_valid_%d.png", time.Now().Unix())
			if err := pageHandler.TakeScreenshot(screenshotPath); err != nil {
				log.Printf("Warning: Failed to take screenshot: %v", err)
			} else {
				log.Printf("Screenshot saved to: %s", screenshotPath)
			}
		}

		return url, nil
	}

	return "", nil
}

// isValidPage checks if a HelloAsso page is valid for tournament registration
func isValidPage(pageContent string, tournament cache.TournamentCache, tournamentDate time.Time) bool {
	// Check if it's a HelloAsso form page
	if !strings.Contains(pageContent, "form-container") && !strings.Contains(pageContent, "checkout-container") {
		return false
	}

	// Check for table tennis keywords
	if !strings.Contains(pageContent, "tennis de table") &&
		!strings.Contains(pageContent, "ping pong") &&
		!strings.Contains(pageContent, "ping") &&
		!strings.Contains(pageContent, "tt") &&
		!strings.Contains(pageContent, "TT") {
		return false
	}

	// Check for tournament keywords
	tournamentKeywords := []string{"tournoi", "compétition", "competition", "open", "criterium"}
	hasTournamentKeyword := false
	for _, keyword := range tournamentKeywords {
		if strings.Contains(pageContent, keyword) {
			hasTournamentKeyword = true
			break
		}
	}

	if !hasTournamentKeyword {
		return false
	}

	// Check for mention of club name
	clubName := strings.ToLower(tournament.Club.Name)
	if clubName != "" && !strings.Contains(pageContent, clubName) {
		return false
	}

	// Check for current season or date
	currentYear := time.Now().Year()
	nextYear := currentYear + 1

	yearMatches := strings.Contains(pageContent, fmt.Sprintf("%d", currentYear)) ||
		strings.Contains(pageContent, fmt.Sprintf("%d", nextYear))

	seasonPattern := fmt.Sprintf("%d-%d", currentYear, nextYear)
	seasonMatches := strings.Contains(pageContent, seasonPattern)

	// Try to match the month name
	monthName := utils.GetMonthNameFrench(int(tournamentDate.Month()))
	monthMatches := strings.Contains(pageContent, strings.ToLower(monthName))

	if !yearMatches && !seasonMatches && !monthMatches {
		return false
	}

	// This is a valid HelloAsso registration page
	return true
}

// ValidateHelloAssoURL validates a HelloAsso URL for a tournament
func ValidateHelloAssoURL(urlStr string, tournament cache.TournamentCache, tournamentDate time.Time, browserContext pw.BrowserContext) (string, error) {
	utils.DebugLog("Validating HelloAsso URL: %s", urlStr)

	// Clean up the URL (simplify for validation)
	cleanedURL := utils.CleanURL(urlStr)

	// Create a new page with viewport settings
	pwPage, err := browser.NewPageWithViewport(browserContext, 1280, 800)
	if err != nil {
		return "", fmt.Errorf("failed to create page: %w", err)
	}
	defer browser.SafeClose(pwPage)

	// Create a page handler
	pageHandler := page.NewPageHandler(pwPage)
	defer pageHandler.Close()

	// Set default timeouts - HelloAsso can be slow
	pageHandler.SetDefaultTimeouts(30*time.Second, 20*time.Second)

	// Navigate to the URL
	utils.DebugLog("Navigating to HelloAsso URL: %s", cleanedURL)
	if err := pageHandler.SafeNavigation(cleanedURL, 2, nil); err != nil {
		return "", fmt.Errorf("failed to navigate to HelloAsso URL: %w", err)
	}

	// Get the final URL after any redirects
	finalURL := pwPage.URL()
	utils.DebugLog("Final HelloAsso URL after navigation: %s", finalURL)

	// Check if the page is still a HelloAsso page
	if !IsHelloAssoURL(finalURL) {
		return "", fmt.Errorf("redirected to non-HelloAsso URL: %s", finalURL)
	}

	// Check if the page is an active registration form
	isActive, err := CheckFormIsActive(pwPage)
	if err != nil {
		utils.DebugLog("Error checking if HelloAsso form is active: %v", err)
	}

	if !isActive {
		// Take screenshot for debugging
		if utils.IsDebugMode() {
			screenshotPath := fmt.Sprintf("helloasso_inactive_%d.png", time.Now().Unix())
			pageHandler.TakeScreenshot(screenshotPath)
			log.Printf("Inactive form screenshot saved to: %s", screenshotPath)
		}
		return "", fmt.Errorf("HelloAsso form is not active (registration closed)")
	}

	// Check if the page contains the tournament name or date
	isRelated, err := ContainsTournamentInfo(pwPage, tournament, tournamentDate)
	if err != nil {
		utils.DebugLog("Error checking if HelloAsso form is related to tournament: %v", err)
	}

	if !isRelated {
		utils.DebugLog("HelloAsso form doesn't appear to be related to this tournament")
		// For HelloAsso, we'll still return the URL even if we're not 100% sure it's for this tournament
		// since HelloAsso URLs are highly likely to be registration forms
		log.Printf("Found HelloAsso URL (might not be for this specific tournament): %s", finalURL)
		return finalURL, nil
	}

	log.Printf("Found valid HelloAsso signup URL for tournament: %s", finalURL)
	return finalURL, nil
}

// -----------------------------------------------------------------------------
// Helper Functions
// -----------------------------------------------------------------------------

// CleanHelloAssoURL simplifies a HelloAsso URL to its canonical form
func CleanHelloAssoURL(url string) string {
	return utils.CleanURL(url)
}

// CheckFormIsActive checks if a form is active on the page
func CheckFormIsActive(page pw.Page) (bool, error) {
	// Check for closed form indicators
	script := `
	() => {
		const closedElements = document.querySelectorAll('.registration-closed, .form-closed, .h-rXiJeRZz3ISiX5K07gkk');
		return closedElements.length === 0;
	}
	`
	result, err := page.Evaluate(script)
	if err != nil {
		return false, err
	}

	if boolResult, ok := result.(bool); ok {
		return boolResult, nil
	}

	return false, fmt.Errorf("unexpected result type from form check")
}

// ContainsTournamentInfo checks if the page contains tournament information
func ContainsTournamentInfo(page pw.Page, tournament cache.TournamentCache, tournamentDate time.Time) (bool, error) {
	// Get page content
	content, err := page.Content()
	if err != nil {
		return false, err
	}

	contentLower := strings.ToLower(content)

	// Check for tournament name
	if tournament.Name != "" && strings.Contains(contentLower, strings.ToLower(tournament.Name)) {
		return true, nil
	}

	// Check for club name
	if tournament.Club.Name != "" && strings.Contains(contentLower, strings.ToLower(tournament.Club.Name)) {
		return true, nil
	}

	// Check for month name
	monthName := utils.GetMonthNameFrench(int(tournamentDate.Month()))
	if strings.Contains(contentLower, strings.ToLower(monthName)) {
		return true, nil
	}

	// Check for year
	year := tournamentDate.Year()
	if strings.Contains(content, fmt.Sprintf("%d", year)) {
		return true, nil
	}

	return false, nil
}
