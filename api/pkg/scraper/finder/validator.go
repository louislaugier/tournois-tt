// Package finder provides navigation services for finding tournament signup URLs
package finder

import (
	"fmt"
	"strings"
	"time"

	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/constants"
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

	// Skip invalid or irrelevant URLs
	if utils.IsDomainToSkip(utils.ExtractDomain(urlStr)) {
		return "", fmt.Errorf("domain is in skip list: %s", urlStr)
	}

	// Create a new page to validate the URL
	page, err := browserContext.NewPage()
	if err != nil {
		return "", fmt.Errorf("failed to create page: %w", err)
	}
	defer page.Close()

	// Navigate to the URL
	if _, err := page.Goto(urlStr, pw.PageGotoOptions{
		WaitUntil: pw.WaitUntilStateNetworkidle,
		Timeout:   pw.Float(DefaultPageTimeout),
	}); err != nil {
		return "", fmt.Errorf("failed to navigate to URL: %w", err)
	}

	// Check if the URL appears to be a signup form
	isSignupForm, err := IsSignupForm(page, tournament, tournamentDate)
	if err != nil {
		return "", fmt.Errorf("error checking if URL is a signup form: %w", err)
	}

	if isSignupForm {
		utils.DebugLog("URL appears to be a signup form: %s", urlStr)
		return urlStr, nil
	}

	return "", fmt.Errorf("URL does not appear to be a signup form: %s", urlStr)
}

// IsSignupForm checks if a page is likely to be a registration/signup form
func IsSignupForm(page pw.Page, tournament cache.TournamentCache, tournamentDate time.Time) (bool, error) {
	// Match specific keywords in the URL
	currentURL := page.URL()

	urlLower := strings.ToLower(currentURL)
	if strings.Contains(urlLower, "inscription") || strings.Contains(urlLower, "register") ||
		strings.Contains(urlLower, "signup") || strings.Contains(urlLower, "enroll") ||
		strings.Contains(urlLower, "registration") {
		return true, nil
	}

	// Check for tournament-specific name patterns
	isSignupForm, err := CheckIfPageIsSignupForm(page, tournament, tournamentDate)
	if err != nil {
		return false, fmt.Errorf("error checking if page is a signup form: %w", err)
	}

	if isSignupForm {
		return true, nil
	}

	// Detect form elements on the page
	hasForm, err := ContainsFormElements(page)
	if err != nil {
		return false, fmt.Errorf("error detecting form elements: %w", err)
	}

	if hasForm {
		// Check if the page has tournament-related keywords
		pageTitle, err := page.Title()
		if err != nil {
			return false, fmt.Errorf("failed to get page title: %w", err)
		}

		pageContent, err := page.TextContent("body")
		if err != nil {
			return false, fmt.Errorf("failed to get page content: %w", err)
		}

		pageContentLower := strings.ToLower(pageContent)
		pageTitleLower := strings.ToLower(pageTitle)

		// Check for registration-related keywords
		for _, keyword := range registrationKeywords {
			if strings.Contains(pageContentLower, keyword) || strings.Contains(pageTitleLower, keyword) {
				return true, nil
			}
		}

		// Check for tournament name patterns
		for _, pattern := range constants.TournamentPatterns {
			if strings.Contains(pageContentLower, pattern) {
				return true, nil
			}
		}

		// Check for capitalized variations
		for _, pattern := range constants.TournamentPatternsCapitalized {
			if strings.Contains(pageContent, pattern) {
				return true, nil
			}
		}

		// Check for special tournament names
		for _, pattern := range constants.SpecialTournaments {
			if strings.Contains(pageContentLower, pattern) {
				return true, nil
			}
		}

		// Check for capitalized special tournament names
		for _, pattern := range constants.SpecialTournamentsCapitalized {
			if strings.Contains(pageContent, pattern) {
				return true, nil
			}
		}
	}

	return false, nil
}

// ContainsFormElements checks if a page contains form elements
func ContainsFormElements(page pw.Page) (bool, error) {
	jsScript := `
	() => {
		// Check for forms
		const forms = document.querySelectorAll('form');
		if (forms.length > 0) {
			return true;
		}
		
		// Check for input fields
		const inputs = document.querySelectorAll('input, select, textarea');
		if (inputs.length > 3) {  // Typically a form has multiple input fields
			return true;
		}
		
		return false;
	}`

	result, err := page.Evaluate(jsScript)
	if err != nil {
		return false, fmt.Errorf("failed to evaluate script: %w", err)
	}

	hasFormElements, ok := result.(bool)
	if !ok {
		return false, fmt.Errorf("failed to convert result to boolean")
	}

	return hasFormElements, nil
}

// CheckIfPageIsSignupForm determines if a page is likely a tournament signup form
func CheckIfPageIsSignupForm(page pw.Page, tournament cache.TournamentCache, tournamentDate time.Time) (bool, error) {
	// Get the page content
	content, err := page.Content()
	if err != nil {
		return false, fmt.Errorf("failed to get page content: %w", err)
	}

	// Get the page title
	title, err := page.Title()
	if err != nil {
		title = "" // Use empty string if title can't be retrieved
	}

	// Convert to lowercase for case-insensitive matching
	contentLower := strings.ToLower(content)
	titleLower := strings.ToLower(title)

	// Prepare tournament info for matching
	tournamentNameLower := strings.ToLower(tournament.Name)
	tournamentWords := utils.ExtractSignificantWordsFromText(tournamentNameLower)

	// Get month name in French and English
	monthFrench := utils.GetMonthNameFrench(int(tournamentDate.Month()))
	monthEnglish := tournamentDate.Month().String()
	yearStr := fmt.Sprintf("%d", tournamentDate.Year())

	// Check if page contains form elements
	hasFormElements := strings.Contains(content, "<form") ||
		strings.Contains(content, "<input") ||
		strings.Contains(content, "type=\"submit\"") ||
		strings.Contains(content, "type=\"text\"") ||
		strings.Contains(content, "type=\"email\"") ||
		strings.Contains(content, "type=\"password\"")

	// Check if page mentions the tournament
	tournamentNameMatch := false
	for _, word := range tournamentWords {
		if strings.Contains(titleLower, word) || strings.Contains(contentLower, word) {
			tournamentNameMatch = true
			break
		}
	}

	// Check if page mentions the tournament date
	dateMatch := (strings.Contains(titleLower, monthFrench) || strings.Contains(contentLower, monthFrench) ||
		strings.Contains(titleLower, monthEnglish) || strings.Contains(contentLower, monthEnglish)) &&
		(strings.Contains(titleLower, yearStr) || strings.Contains(contentLower, yearStr))

	// Check if page mentions registration
	registrationMatch := false
	for _, keyword := range registrationKeywords {
		if strings.Contains(titleLower, keyword) || strings.Contains(contentLower, keyword) {
			registrationMatch = true
			break
		}
	}

	// The page is likely a signup form if:
	// 1. It has form elements AND
	// 2. (It mentions both the tournament name and registration OR it mentions the tournament name and date)
	isSignupForm := hasFormElements && ((tournamentNameMatch && registrationMatch) || (tournamentNameMatch && dateMatch))

	// Additional checks for common signup patterns
	if !isSignupForm && hasFormElements {
		// Check for "créer un compte" or "se connecter" forms for tournament platforms
		if (strings.Contains(contentLower, constants.CreateAccount) || strings.Contains(contentLower, constants.ENCreateAccount)) &&
			(tournamentNameMatch || dateMatch) {
			isSignupForm = true
		}

		// Check for "inscription" in the URL
		currentURL := page.URL()
		if strings.Contains(strings.ToLower(currentURL), constants.Register) ||
			strings.Contains(strings.ToLower(currentURL), constants.ENSignUp) ||
			strings.Contains(strings.ToLower(currentURL), constants.ENRegister) ||
			strings.Contains(strings.ToLower(currentURL), constants.Engagement) {
			// If the URL itself suggests it's a signup page
			if tournamentNameMatch || dateMatch {
				isSignupForm = true
			}
		}

		// Check for "étape suivante" (next step) in form context
		if !isSignupForm && hasFormElements {
			if (strings.Contains(contentLower, constants.NextStepNoAccent) ||
				strings.Contains(contentLower, constants.NextStep) ||
				strings.Contains(contentLower, constants.Next) ||
				strings.Contains(contentLower, constants.Continue)) &&
				(tournamentNameMatch || dateMatch) {
				isSignupForm = true
			}
		}
	}

	// Check for special tournament patterns
	if !isSignupForm && hasFormElements {
		// Check for patterns like "tournoi de pâques", "tournoi de noël", etc.
		for _, pattern := range constants.TournamentPatterns {
			if strings.Contains(contentLower, pattern) {
				isSignupForm = true
				break
			}
		}
	}

	return isSignupForm, nil
}

// DetectSignupKeywords checks if the page contains signup-related keywords
func DetectSignupKeywords(page pw.Page) (bool, error) {
	jsScript := `
	() => {
		const signupKeywords = [
			'inscription', 'inscrivez', 'inscrire', 's\'inscrire', 'inscriptions', 
			'signup', 'sign up', 'sign-up', 'register', 'registration', 'enroll', 'enregistrement',
			'participer', 'participation', 'engagements', 'engagement',
			'tournoi', 'tournament', 'competition', 'compétition',
			'paiement', 'payment', 'prix', 'price', 'tarif', 'fee', 'checkout'
		];
		
		const pageText = document.body.innerText.toLowerCase();
		
		for (const keyword of signupKeywords) {
			if (pageText.includes(keyword)) {
				return true;
			}
		}
		
		return false;
	}
	`

	result, err := page.Evaluate(jsScript)
	if err != nil {
		return false, fmt.Errorf("failed to evaluate script: %w", err)
	}

	if hasKeywords, ok := result.(bool); ok {
		return hasKeywords, nil
	}

	return false, nil
}

// DetectFormElements checks if the page contains form elements - more detailed than ContainsFormElements
func DetectFormElements(page pw.Page) (bool, error) {
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

	result, err := page.Evaluate(jsScript)
	if err != nil {
		return false, fmt.Errorf("failed to evaluate script: %w", err)
	}

	if hasForm, ok := result.(bool); ok {
		return hasForm, nil
	}

	return false, nil
}
