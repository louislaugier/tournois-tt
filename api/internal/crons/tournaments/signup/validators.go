package signup

import (
	"fmt"
	"net/url"
	"time"

	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/helloasso"
	"tournois-tt/api/pkg/scraper/finder"
	"tournois-tt/api/pkg/utils"

	pw "github.com/playwright-community/playwright-go"
)

// ValidateSignupURL validates a URL as a tournament signup form
func ValidateSignupURL(urlStr string, tournament cache.TournamentCache, tournamentDate time.Time, browserContext pw.BrowserContext) (string, error) {
	utils.DebugLog("Validating potential signup URL: %s", urlStr)

	// Clean the URL first
	urlStr = utils.CleanURL(urlStr)

	// Basic URL validation
	urlObj, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("invalid URL format: %w", err)
	}

	// Skip invalid or common domains
	if utils.IsDomainToSkip(urlObj.Host) {
		return "", fmt.Errorf("domain '%s' is in skip list", urlObj.Host)
	}

	// First, check if it's a HelloAsso URL
	if helloasso.IsHelloAssoURL(urlStr) {
		utils.DebugLog("Detected HelloAsso URL: %s", urlStr)
		return helloasso.ValidateHelloAssoURL(urlStr, tournament, tournamentDate, browserContext)
	}

	// Add other specialized validators here
	// e.g., BilletWeb, Weezevent, etc.

	// Generic validator for other platforms
	utils.DebugLog("Using generic validator for URL: %s", urlStr)

	// Skip URLs unlikely to be signup forms
	if utils.IsURLToSkip(urlStr) {
		return "", fmt.Errorf("URL pattern '%s' is unlikely to be a signup form", urlStr)
	}

	// Check if the page contains signup form elements
	return finder.ValidateSignupURL(urlStr, tournament, tournamentDate, browserContext)
}
