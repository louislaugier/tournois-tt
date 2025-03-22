package signup

import (
	"fmt"
	"strings"
	"time"

	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/scraper/browser"
	"tournois-tt/api/pkg/scraper/page"
	"tournois-tt/api/pkg/utils"

	pw "github.com/playwright-community/playwright-go"
)

// ValidateSignupURL checks if a URL is valid for tournament registration
func ValidateSignupURL(url string, tournament cache.TournamentCache, tournamentDate time.Time, browserContext pw.BrowserContext) (string, error) {
	// If it's a HelloAsso URL, validate it directly
	if helloAssoURLRegex.MatchString(url) {
		return validateHelloAssoURL(url, tournament, tournamentDate, browserContext)
	}

	// Otherwise try generic validation
	return validateGenericRegistrationURL(url, tournament, tournamentDate, browserContext)
}

// validateHelloAssoURL checks if a HelloAsso URL is valid for tournament registration
func validateHelloAssoURL(url string, tournament cache.TournamentCache, tournamentDate time.Time, browserContext pw.BrowserContext) (string, error) {
	debugLog("Validating HelloAsso URL: %s", url)

	// Create a new Playwright page
	pwPage, err := browser.NewPage(browserContext)
	if err != nil {
		return "", fmt.Errorf("critical error - failed to create page: %w", err)
	}
	defer pwPage.Close()

	// Create a page handler
	pageHandler := page.New(pwPage)

	// Navigate to the URL
	if err := pageHandler.NavigateToPage(url); err != nil {
		return "", fmt.Errorf("critical error - failed to navigate to URL: %w", err)
	}

	// Get page content
	content, err := pwPage.Content()
	if err != nil {
		return "", fmt.Errorf("critical error - failed to get page content: %w", err)
	}

	// Check if the URL is a valid HelloAsso event that matches our tournament
	contentLower := strings.ToLower(content)
	if isValidHelloAssoPage(contentLower, tournament, tournamentDate) {
		return url, nil
	}

	return "", nil
}

// validateGenericRegistrationURL validates a non-HelloAsso URL for tournament registration
func validateGenericRegistrationURL(url string, tournament cache.TournamentCache, tournamentDate time.Time, browserContext pw.BrowserContext) (string, error) {
	debugLog("Validating potential registration URL: %s", url)

	// Create a new Playwright page
	pwPage, err := browser.NewPage(browserContext)
	if err != nil {
		return "", fmt.Errorf("critical error - failed to create page: %w", err)
	}
	defer pwPage.Close()

	// Create a page handler
	pageHandler := page.New(pwPage)

	// Try to navigate to the URL
	if err := pageHandler.NavigateToPage(url); err != nil {
		return "", fmt.Errorf("critical error - failed to navigate to URL: %w", err)
	}

	// Get page content
	content, err := pwPage.Content()
	if err != nil {
		return "", fmt.Errorf("critical error - failed to get page content: %w", err)
	}

	contentLower := strings.ToLower(content)

	// Try to get the title
	title, err := pwPage.Title()
	if err != nil {
		// If we can't get the title, just use an empty string
		title = ""
	}
	titleLower := strings.ToLower(title)

	// Prepare tournament info for matching
	tournamentNameLower := strings.ToLower(tournament.Name)
	tournamentWords := ExtractSignificantWords(tournamentNameLower)
	monthFrench := utils.GetMonthNameFrench(tournamentDate.Month())
	yearStr := fmt.Sprintf("%d", tournamentDate.Year())

	// Check if page title or content mentions the tournament
	tournamentNameMatch := false
	for _, word := range tournamentWords {
		if strings.Contains(titleLower, word) || strings.Contains(contentLower, word) {
			tournamentNameMatch = true
			break
		}
	}

	// Check if page mentions the tournament date
	dateMatch := (strings.Contains(titleLower, monthFrench) || strings.Contains(contentLower, monthFrench)) &&
		(strings.Contains(titleLower, yearStr) || strings.Contains(contentLower, yearStr))

	// Check if page mentions registration
	registrationMatch := false
	for _, keyword := range registrationKeywords {
		if strings.Contains(titleLower, keyword) || strings.Contains(contentLower, keyword) {
			registrationMatch = true
			break
		}
	}

	// Consider valid if it mentions both the tournament name and registration
	if (tournamentNameMatch && registrationMatch) || (tournamentNameMatch && dateMatch) {
		return url, nil
	}

	return "", nil
}

// isValidHelloAssoPage checks if a HelloAsso page is valid for tournament registration
func isValidHelloAssoPage(pageContent string, tournament cache.TournamentCache, tournamentDate time.Time) bool {
	// Check if it's a HelloAsso form page
	if !strings.Contains(pageContent, "form-container") && !strings.Contains(pageContent, "checkout-container") {
		debugLog("Not a HelloAsso form page")
		return false
	}

	// Check for table tennis keywords
	if !strings.Contains(pageContent, "tennis de table") && 
	   !strings.Contains(pageContent, "ping pong") && 
	   !strings.Contains(pageContent, "ping") && 
	   !strings.Contains(pageContent, "tt") && 
	   !strings.Contains(pageContent, "TT") {
		debugLog("Page does not contain table tennis keywords")
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
		debugLog("Page does not contain tournament keywords")
		return false
	}

	// Check for mention of club name
	clubName := strings.ToLower(tournament.Club.Name)
	if clubName != "" && !strings.Contains(pageContent, clubName) {
		debugLog("Page does not mention the club name: %s", clubName)
		return false
	}

	// Check for current season or date
	currentSeason, _ := utils.GetCurrentSeason()
	currentYear := currentSeason.Year()
	nextYear := currentYear + 1

	yearMatches := strings.Contains(pageContent, fmt.Sprintf("%d", currentYear)) ||
		strings.Contains(pageContent, fmt.Sprintf("%d", nextYear))

	seasonPattern := fmt.Sprintf("%d-%d", currentYear, nextYear)
	seasonMatches := strings.Contains(pageContent, seasonPattern)

	// Try to match the month name
	monthName := utils.GetMonthNameFrench(int(tournamentDate.Month()))
	monthMatches := strings.Contains(pageContent, strings.ToLower(monthName))

	if !yearMatches && !seasonMatches && !monthMatches {
		debugLog("Page does not mention current season or tournament month")
		return false
	}

	// This is a valid HelloAsso registration page
	return true
}
