package common

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/scraper/browser"
	"tournois-tt/api/pkg/scraper/page"
	"tournois-tt/api/pkg/utils"

	pw "github.com/playwright-community/playwright-go"
)

// -----------------------------------------------------------------------------
// URL Helper Functions
// -----------------------------------------------------------------------------

// CleanURL sanitizes a URL by removing tracking parameters and ensuring it has https prefix
func CleanURL(urlStr string) string {
	// Remove any UTM parameters or tracking info
	if strings.Contains(urlStr, "?") {
		urlStr = strings.Split(urlStr, "?")[0]
	}

	// Remove hash fragments
	if strings.Contains(urlStr, "#") {
		urlStr = strings.Split(urlStr, "#")[0]
	}

	// Ensure the URL has https://
	if !strings.HasPrefix(urlStr, "http") {
		urlStr = "https://" + urlStr
	}

	return urlStr
}

// ExtractDomain extracts the domain from a URL
func ExtractDomain(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	return parsedURL.Hostname()
}

// IsDomainToSkip checks if a domain should be skipped during validation (common social media, etc.)
func IsDomainToSkip(domain string) bool {
	commonDomainsToSkip := []string{
		"google.com",
		"facebook.com",
		"instagram.com",
		"twitter.com",
		"youtube.com",
		"linkedin.com",
		"github.com",
		"zoom.us",
		"wikipedia.org",
		"apple.com",
		"microsoft.com",
		"amazonaws.com",
		"cloudfront.net",
		"cdn.com",
	}

	domain = strings.ToLower(domain)
	for _, skipDomain := range commonDomainsToSkip {
		if strings.Contains(domain, skipDomain) {
			return true
		}
	}
	return false
}

// IsURLToSkip checks if a URL should be skipped based on common patterns (assets, static files, etc.)
func IsURLToSkip(urlStr string) bool {
	urlPatternsToSkip := []string{
		"/image/",
		"/images/",
		"/img/",
		"/css/",
		"/js/",
		"/assets/",
		"/static/",
		"/media/",
		"/video/",
		"/downloads/",
		"/docs/",
		"/pdf/",
		".jpg",
		".jpeg",
		".png",
		".gif",
		".css",
		".js",
		".ico",
		".pdf",
		".zip",
		".doc",
		".docx",
		".xls",
		".xlsx",
		".mp4",
		".mp3",
		"privacy",
		"terms",
		"about",
		"contact",
		"admin",
		"login",
	}

	urlLower := strings.ToLower(urlStr)
	for _, pattern := range urlPatternsToSkip {
		if strings.Contains(urlLower, pattern) {
			return true
		}
	}
	return false
}

// -----------------------------------------------------------------------------
// Form and Content Detection Helpers
// -----------------------------------------------------------------------------

// DetectFormElements checks if a page contains form elements
func DetectFormElements(pg pw.Page) (bool, error) {
	jsScript := `
	() => {
		// Check for forms
		const forms = document.querySelectorAll('form');
		if (forms.length > 0) {
			return true;
		}
		
		// Check for input fields
		const inputs = document.querySelectorAll('input[type="text"], input[type="email"], input[type="tel"], input[type="number"], input[type="date"]');
		if (inputs.length >= 3) { // Multiple input fields suggest a form
			return true;
		}
		
		// Check for dropdowns and submit buttons
		const dropdowns = document.querySelectorAll('select');
		const submitButtons = document.querySelectorAll('button[type="submit"], input[type="submit"]');
		
		if ((dropdowns.length > 0 || inputs.length > 0) && submitButtons.length > 0) {
			return true;
		}
		
		// Check for payment elements
		const paymentKeywords = ['payment', 'credit card', 'carte', 'paiement', 'checkout', 'price', 'prix', 'tarif'];
		const allElements = document.body.innerText.toLowerCase();
		
		for (const keyword of paymentKeywords) {
			if (allElements.includes(keyword)) {
				return true;
			}
		}
		
		return false;
	}
	`

	result, err := pg.Evaluate(jsScript)
	if err != nil {
		return false, fmt.Errorf("failed to evaluate script: %w", err)
	}

	if hasForm, ok := result.(bool); ok {
		return hasForm, nil
	}

	return false, nil
}

// DetectSignupKeywords checks if a page contains signup-related keywords
func DetectSignupKeywords(pg pw.Page) (bool, error) {
	jsScript := `
	() => {
		const pageText = document.body.innerText.toLowerCase();
		
		// Check for signup keywords in French and English
		const signupKeywords = [
			'inscription', 'inscriptions', 'inscrire', 's\'inscrire',
			'registre', 'enregistrer', 's\'enregistrer',
			'tarif', 'tarifs', 'paiement', 'payer',
			'formulaire', 'form', 'registration', 'register', 'signup',
			'engagement', 'engagements',
			'etape suivante', 'étape suivante', 'suivant', 'continuer'
		];
		
		for (const keyword of signupKeywords) {
			if (pageText.includes(keyword)) {
				return true;
			}
		}
		
		return false;
	}
	`

	result, err := pg.Evaluate(jsScript)
	if err != nil {
		return false, fmt.Errorf("failed to evaluate script: %w", err)
	}

	if hasKeywords, ok := result.(bool); ok {
		return hasKeywords, nil
	}

	return false, nil
}

// CheckFormIsActive checks if a form is active or closed
func CheckFormIsActive(pg pw.Page) (bool, error) {
	jsScript := `
	() => {
		const pageText = document.body.innerText.toLowerCase();
		
		// Check for indicators that the form is closed
		const closedPatterns = [
			'inscription fermée',
			'inscriptions fermées',
			'inscriptions terminées',
			'inscription terminée',
			'plus disponible',
			'n\'est plus disponible',
			'formulaire clôturé',
			'event has ended',
			'enrollment is closed',
			'registration is closed',
			'registration closed',
			'closed',
			'fermé'
		];
		
		for (const pattern of closedPatterns) {
			if (pageText.includes(pattern)) {
				return false;
			}
		}
		
		// Check for form elements that indicate an active form
		const formElements = document.querySelectorAll('form, button[type="submit"]');
		if (formElements.length > 0) {
			return true;
		}
		
		// Check for price/ticket/signup indicators
		const activePatterns = [
			'tarif', 'tarifs', 'price', 'prix',
			'billetterie', 'ticketing', 'registration', 'inscription',
			'formulaire', 'form', 'signup', 'sign up',
			'paiement', 'payment', 'checkout'
		];
		
		for (const pattern of activePatterns) {
			if (pageText.includes(pattern)) {
				return true;
			}
		}
		
		return false;
	}
	`

	result, err := pg.Evaluate(jsScript)
	if err != nil {
		return false, fmt.Errorf("failed to evaluate script: %w", err)
	}

	if isActive, ok := result.(bool); ok {
		return isActive, nil
	}

	return false, nil
}

// ContainsTournamentInfo checks if a page contains tournament information
func ContainsTournamentInfo(pg pw.Page, tournament cache.TournamentCache, tournamentDate time.Time) (bool, error) {
	// Get the tournament name and normalize it for comparison
	tournamentName := strings.ToLower(tournament.Name)

	// Try to extract words from the tournament name for partial matching
	nameTokens := strings.Fields(tournamentName)
	significantTokens := []string{}

	// Filter out common words and keep only significant ones
	for _, token := range nameTokens {
		if len(token) >= 4 && !IsCommonWord(token) {
			significantTokens = append(significantTokens, token)
		}
	}

	// Generate date strings in various formats
	dateStrings := []string{
		tournamentDate.Format("02/01/2006"),
		tournamentDate.Format("2006-01-02"),
		fmt.Sprintf("%d/%d/%d", tournamentDate.Day(), tournamentDate.Month(), tournamentDate.Year()),
		fmt.Sprintf("%d/%d", tournamentDate.Day(), tournamentDate.Month()),
	}

	// Add French date formats
	monthFrench := utils.GetMonthNameFrench(tournamentDate.Month())
	frenchDateStrings := []string{
		fmt.Sprintf("%d %s %d", tournamentDate.Day(), monthFrench, tournamentDate.Year()),
		fmt.Sprintf("%d %s", tournamentDate.Day(), monthFrench),
	}

	dateStrings = append(dateStrings, frenchDateStrings...)

	jsScript := fmt.Sprintf(`
	() => {
		const pageText = document.body.innerText.toLowerCase();
		
		// Check tournament name (exact match)
		const tournamentName = %q;
		if (pageText.includes(tournamentName)) {
			return {
				match: true,
				reason: "Tournament name exact match"
			};
		}
		
		// Check tournament name tokens (partial match)
		const significantTokens = %v;
		let tokenMatches = 0;
		const matchedTokens = [];
		
		for (const token of significantTokens) {
			if (pageText.includes(token)) {
				tokenMatches++;
				matchedTokens.push(token);
			}
		}
		
		// If multiple tokens match, it's likely related to the tournament
		if (tokenMatches >= 2 || (significantTokens.length === 1 && tokenMatches === 1)) {
			return {
				match: true,
				reason: "Tournament name partial match: " + matchedTokens.join(", ")
			};
		}
		
		// Check for dates
		const dateStrings = %v;
		for (const dateStr of dateStrings) {
			if (pageText.includes(dateStr)) {
				return {
					match: true,
					reason: "Tournament date match: " + dateStr
				};
			}
		}
		
		// Check for club name
		const clubName = %q;
		if (clubName && pageText.includes(clubName.toLowerCase())) {
			return {
				match: true,
				reason: "Club name match: " + clubName
			};
		}
		
		// Check for table tennis keywords with year match
		const sportKeywords = [
			'ping-pong', 'ping pong', 'table tennis', 'tennis de table', 
			'tt', 'astt', 'ustt', 'ppc', 'pongiste'
		];
		
		const yearStr = %q;
		for (const keyword of sportKeywords) {
			if (pageText.includes(keyword) && pageText.includes(yearStr)) {
				return {
					match: true,
					reason: "Sport keyword with year match: " + keyword
				};
			}
		}
		
		return {
			match: false,
			reason: "No match found"
		};
	}
	`, tournamentName, significantTokens, dateStrings, tournament.Club.Name, fmt.Sprintf("%d", tournamentDate.Year()))

	result, err := pg.Evaluate(jsScript)
	if err != nil {
		return false, fmt.Errorf("failed to evaluate script: %w", err)
	}

	// Parse the result which should be an object with match and reason properties
	if resultObj, ok := result.(map[string]interface{}); ok {
		if match, ok := resultObj["match"].(bool); ok {
			if reason, ok := resultObj["reason"].(string); ok {
				utils.DebugLog("Tournament relation check: %v (%s)", match, reason)
			}
			return match, nil
		}
	}

	return false, nil
}

// -----------------------------------------------------------------------------
// Common Validation Functions
// -----------------------------------------------------------------------------

// ValidateSignupURL validates if a URL is a tournament signup form
// Returns the valid URL if found, or empty string if not valid
func ValidateSignupURL(urlStr string, tournament cache.TournamentCache, tournamentDate time.Time, browserContext pw.BrowserContext) (string, error) {
	utils.DebugLog("Validating potential signup URL: %s", urlStr)

	// Clean up the URL
	cleanedURL := CleanURL(urlStr)

	// Create a new page with viewport settings
	playwrightPage, err := browser.NewPageWithViewport(browserContext, 1280, 800)
	if err != nil {
		return "", fmt.Errorf("failed to create page: %w", err)
	}
	defer browser.SafeClose(playwrightPage)

	// Create page handler for better operations
	pageHandler := page.New(playwrightPage)
	defer pageHandler.Close()

	// Set default timeouts
	pageHandler.SetDefaultTimeouts(30*time.Second, 15*time.Second)

	// Navigate to the URL using safer navigation
	if err := pageHandler.SafeNavigation(cleanedURL, 2, browser.RestartIfUnhealthy); err != nil {
		return "", fmt.Errorf("failed to navigate to URL: %w", err)
	}

	// Take a screenshot for debugging if needed
	if utils.IsDebugMode() {
		screenshotPath := fmt.Sprintf("validate_url_%d.png", time.Now().Unix())
		if err := pageHandler.TakeScreenshot(screenshotPath); err == nil {
			log.Printf("Screenshot saved to: %s", screenshotPath)
		}
	}

	// Check if this URL is a signup form
	var hasFormElements, hasSignupKeywords, hasTournamentInfo bool

	// Use SafeOperation for reliable checks
	err = pageHandler.SafeOperation("check form elements", func() error {
		var err error
		hasFormElements, err = DetectFormElements(playwrightPage)
		return err
	})
	if err != nil {
		utils.DebugLog("Error detecting form elements: %v", err)
	}

	err = pageHandler.SafeOperation("check signup keywords", func() error {
		var err error
		hasSignupKeywords, err = DetectSignupKeywords(playwrightPage)
		return err
	})
	if err != nil {
		utils.DebugLog("Error detecting signup keywords: %v", err)
	}

	err = pageHandler.SafeOperation("check tournament info", func() error {
		var err error
		hasTournamentInfo, err = ContainsTournamentInfo(playwrightPage, tournament, tournamentDate)
		return err
	})
	if err != nil {
		utils.DebugLog("Error checking for tournament info: %v", err)
	}

	// Log findings
	utils.DebugLog("URL %s validation: hasForm=%v, hasKeywords=%v, hasTournamentInfo=%v",
		cleanedURL, hasFormElements, hasSignupKeywords, hasTournamentInfo)

	// If the page has form elements or signup keywords and tournament info, it's likely a signup form
	if (hasFormElements || hasSignupKeywords) && hasTournamentInfo {
		log.Printf("Found valid signup URL: %s", cleanedURL)
		return cleanedURL, nil
	}

	// If we have strong form indicators but couldn't find tournament info,
	// we'll return it as a potential match but with a warning
	if hasFormElements && hasSignupKeywords {
		log.Printf("Found potential signup URL (missing tournament info): %s", cleanedURL)
		return cleanedURL, nil
	}

	return "", nil
}

// -----------------------------------------------------------------------------
// Utility Functions
// -----------------------------------------------------------------------------

// IsCommonWord checks if a word is too common to be significant for matching
func IsCommonWord(word string) bool {
	commonWords := []string{
		"tournoi", "open", "tournament", "competition", "championnat", "championship",
		"tennis", "table", "ping", "pong", "club", "association", "de", "du", "des", "la", "le", "les",
		"and", "the", "et", "à", "au", "aux", "en", "par", "pour", "sur", "un", "une",
	}

	word = strings.ToLower(word)
	for _, common := range commonWords {
		if word == common {
			return true
		}
	}

	return false
}

// ExtractSignificantWords extracts words from text that are meaningful for matching
func ExtractSignificantWords(text string) []string {
	words := strings.Fields(strings.ToLower(text))
	significant := []string{}

	for _, word := range words {
		// Only consider words of reasonable length and not in common word list
		if len(word) >= 4 && !IsCommonWord(word) {
			significant = append(significant, word)
		}
	}

	return significant
}

// Contains checks if a string slice contains a specific string
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
