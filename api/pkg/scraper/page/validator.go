package page

import (
	"fmt"
	"log"
	"strings"
	"time"
	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/scraper/constants"
	"tournois-tt/api/pkg/utils"

	pw "github.com/playwright-community/playwright-go"
)

// Priority levels for signup links
const (
	PriorityHigh   = "HIGH"
	PriorityMedium = "MEDIUM"
	PriorityLow    = "LOW"
)

// Common skip keywords and domains
var (
	// Common words that typically indicate we should skip a link
	skipWords = []string{
		"facebook", "twitter", "instagram", "youtube", "linkedin",
		"privacy", "cookies", "terms", "conditions", "mailto",
		"tel:", "sms:", "whatsapp:", "viber:", "tg:",
		"javascript:", "void", "#",
	}

	// Words related to registration/signup
	registrationKeywords = []string{
		constants.Register, constants.Registers, constants.ToRegister, constants.SignUp,
		"registre", "s'enregistrer",
		constants.SignUpForm, constants.Participate, constants.Registration,
		constants.Engagement, constants.Engagements,
		constants.NextStep, constants.NextStepNoAccent, constants.Next, constants.Continue,
	}

	// Domains that we know are not relevant for tournament signup
	skipDomains = []string{
		"facebook.com", "twitter.com", "instagram.com", "youtube.com",
		"linkedin.com", "google.com", "apple.com", "microsoft.com",
		"amazon.com", "pinterest.com", "snapchat.com", "tiktok.com",
		"whatsapp.com", "telegram.org", "github.com", "gitlab.com",
		"adobe.com", "wordpress.com", "wix.com", "squarespace.com",
		"weebly.com", "godaddy.com", "namecheap.com", "shopify.com",
	}
)

// ExtractSignificantWords extracts significant words from tournament name
func ExtractSignificantWords(text string) []string {
	// Skip common words, keep only significant ones
	commonWords := map[string]bool{
		"le": true, "la": true, "les": true, "du": true, "de": true, "des": true,
		"un": true, "une": true, "et": true, "ou": true, "a": true, "à": true,
		"au": true, "aux": true, "pour": true, "par": true, "en": true, "dans": true,
		"sur": true, "sous": true, "avec": true, "sans": true, "club": true,
		"tournoi": true, "open": true, "tennis": true, "table": true, "tt": true,
	}

	// Split text into words, keeping only significant ones (3+ chars, not in common words)
	parts := strings.Fields(text)
	filteredWords := make([]string, 0, len(parts))

	for _, word := range parts {
		word = strings.ToLower(strings.Trim(word, ".,;:!?\"'()[]{}<>"))
		if len(word) >= 3 && !commonWords[word] {
			filteredWords = append(filteredWords, word)
		}
	}

	return filteredWords
}

// ValidateSignupURL validates and potentially follows a URL to check if it's a tournament signup page
func ValidateSignupURL(url string, tournament cache.TournamentCache, tournamentDate time.Time, browserContext pw.BrowserContext) (string, error) {
	// Skip common social media platforms, external services, etc.
	if IsURLToSkip(url) {
		return "", nil
	}

	// Attempt to navigate to the URL
	page, err := browserContext.NewPage()
	if err != nil {
		return "", fmt.Errorf("failed to create new page: %w", err)
	}
	defer page.Close()

	// Set navigation timeout to 30 seconds
	// Using Goto with timeout options instead of setting default timeout

	// Try to navigate to the URL
	resp, err := page.Goto(url, pw.PageGotoOptions{
		WaitUntil: pw.WaitUntilStateNetworkidle,
		Timeout:   pw.Float(30000), // 30 seconds in milliseconds
	})
	if err != nil {
		return "", fmt.Errorf("failed to navigate to %s: %w", url, err)
	}

	// Check HTTP status code
	statusCode := resp.Status()
	if statusCode >= 400 {
		return "", fmt.Errorf("received HTTP error %d when accessing %s", statusCode, url)
	}

	// Check if the page looks like a signup form
	isSignupForm, err := checkIfPageIsSignupForm(page, tournament, tournamentDate)
	if err != nil {
		log.Printf("Warning: Failed to check if page is signup form: %v", err)
	}

	// If it appears to be a signup form, return the final URL after navigation
	if isSignupForm {
		return page.URL(), nil
	}

	// Look for potential signup links on the page using JavaScript
	jsScript := `
	() => {
		const PRIORITY = {
			HIGH: 'HIGH',
			MEDIUM: 'MEDIUM',
			LOW: 'LOW'
		};
		
		const CTA_REASON = {
			DIRECT_SIGNUP: 'direct_signup',
			TOURNAMENT_AND_YEAR: 'tournament_and_year',
			ACCOUNT_LOGIN: 'account_login',
			PARTICIPATION_BUTTON: 'participation_button',
			NEXT_STEP: 'next_step'
		};
		
		const tournamentWords = ['` + strings.Join(ExtractSignificantWords(tournament.Name), "', '") + `'];
		const year = '` + fmt.Sprintf("%d", tournamentDate.Year()) + `';
		
		const signupKeywords = ['` + constants.Register + `', '` + constants.ENRegister + `', '` + constants.ENSignUp + `', "` + constants.SignUp + `", '` + constants.SignUpForm + `', '` + constants.Participate + `', '` + constants.Registration + `'];
		
		// Look for signup links by their text content
		const links = Array.from(document.querySelectorAll('a, button'));
		
		for (const link of links) {
			// Skip links with no text
			const text = link.textContent.toLowerCase().trim();
			const title = (link.getAttribute('title') || '').toLowerCase().trim();
			const href = link.tagName === 'A' ? link.getAttribute('href') : null;
			
			// Skip empty links
			if (text === '' && title === '') continue;
			
			// Skip links that are just JavaScript actions or anchors
			if (!href || href.startsWith('javascript:') || href === '#') continue;
			
			// Check if the link mentions both tournament and year
			const mentionsTournament = tournamentWords.some(word => text.includes(word) || title.includes(word));
			const mentionsYear = text.includes(year.toString()) || title.includes(year.toString());
			
			if (mentionsTournament && mentionsYear) {
				return { url: href, priority: PRIORITY.LOW, reason: CTA_REASON.TOURNAMENT_AND_YEAR };
			}
			
			// Check for "créer un compte" or "se connecter" links as they often lead to signup portals
			if (text.includes('` + constants.CreateAccount + `') || text.includes('` + constants.Login + `') || 
				text.includes('` + constants.ENCreateAccount + `') || text.includes('` + constants.ENLogin + `') || 
				text.includes('` + constants.ENSignIn + `') || text.includes('` + constants.Connection + `')) {
				return { url: href, priority: PRIORITY.MEDIUM, reason: CTA_REASON.ACCOUNT_LOGIN };
			}

			// Check for participation buttons (new CTA type)
			if ((text.includes('` + constants.Participate + `') || text.includes('` + constants.ENParticipate + `') || title.includes('` + constants.Participate + `')) &&
				 (mentionsTournament || mentionsYear)) {
				return { url: href, priority: PRIORITY.MEDIUM, reason: CTA_REASON.PARTICIPATION_BUTTON };
			}
			
			// Check for "étape suivante" (next step) links that indicate progression in a signup flow
			if (text.includes('` + constants.NextStepNoAccent + `') || text.includes('` + constants.NextStep + `') || 
				text.includes('` + constants.ENNextStep + `') || text.includes('` + constants.Next + `') || 
				text.includes('` + constants.Continue + `') || text.includes('` + constants.ENContinue + `')) {
				return { url: href, priority: PRIORITY.MEDIUM, reason: CTA_REASON.NEXT_STEP };
			}
		}
		
		return null;
	}
	`

	result, err := page.Evaluate(jsScript)
	if err != nil {
		return "", fmt.Errorf("failed to evaluate JavaScript: %w", err)
	}

	if result != nil {
		// Try to extract link information
		if linkInfo, ok := result.(map[string]interface{}); ok {
			if linkURL, ok := linkInfo["url"].(string); ok {
				// If this is a high or medium priority link, try to validate it recursively
				priority, _ := linkInfo["priority"].(string)
				// Use the same priority levels as defined in the JavaScript
				if priority == PriorityHigh || priority == PriorityMedium {
					// Navigate to this link and validate recursively
					followedURL, err := ValidateSignupURL(linkURL, tournament, tournamentDate, browserContext)
					if err == nil && followedURL != "" {
						return followedURL, nil
					}
				}
			}
		}
	}

	// No valid signup URL found
	return "", nil
}

// checkIfPageIsSignupForm determines if a page is likely a tournament signup form
func checkIfPageIsSignupForm(page pw.Page, tournament cache.TournamentCache, tournamentDate time.Time) (bool, error) {
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
	tournamentWords := ExtractSignificantWords(tournamentNameLower)

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

	return isSignupForm, nil
}

// CheckWebsiteHeaderForSignupLink checks a website's header/navigation for signup links
func CheckWebsiteHeaderForSignupLink(websiteURL string, browserContext pw.BrowserContext) (string, error) {
	// Create a new page
	page, err := browserContext.NewPage()
	if err != nil {
		return "", fmt.Errorf("failed to create page: %w", err)
	}
	defer page.Close()

	// Navigate to the website
	response, err := page.Goto(websiteURL, pw.PageGotoOptions{
		WaitUntil: pw.WaitUntilStateNetworkidle,
		Timeout:   pw.Float(15 * 1000), // 15 seconds timeout
	})
	if err != nil {
		return "", fmt.Errorf("failed to navigate to URL: %w", err)
	}

	if response == nil || response.Status() >= 400 {
		return "", fmt.Errorf("received error status: %d", response.Status())
	}

	// JavaScript to find links in the header/navigation containing registration keywords
	jsScript := `
	() => {
		// Helper function to check if an element is likely a navigation/header element
		function isHeaderOrNav(element) {
			const tag = element.tagName.toLowerCase();
			
			// Check tag name
			if (tag === 'header' || tag === 'nav' || tag === 'menu') {
				return true;
			}
			
			// Check element id
			const id = element.id.toLowerCase();
			if (id.includes('header') || id.includes('nav') || id.includes('menu') || 
				id.includes('top') || id.includes('main-menu')) {
				return true;
			}
			
			// Check element class
			const className = element.className.toLowerCase();
			if (className.includes('header') || className.includes('nav') || 
				className.includes('menu') || className.includes('top-bar') || 
				className.includes('navbar')) {
				return true;
			}
			
			return false;
		}
		
		// Keywords for registration related links
		const keywords = ['inscription', 'register', 'signup', 'sign up', 'sign-up',
		  's\'inscrire', 'formulaire', 'form', 'tournoi', 'tournament',
		  'participer', 'engagement', 'registration', 'compétition',
		  'competition', 'évènement', 'event'];
		
		// Find all potential navigation/header elements
		const headerElements = Array.from(document.querySelectorAll('*')).filter(isHeaderOrNav);
		
		// Find all links in headers that contain registration keywords
		const signupLinks = [];
		
		// Process links in header elements first
		headerElements.forEach(header => {
			const links = Array.from(header.querySelectorAll('a'));
			
			links.forEach(link => {
				// Skip if no href or it's just a hash or javascript
				if (!link.href || link.href === '' || link.href.startsWith('#') || 
					link.href.startsWith('javascript:') || link.href.includes('mailto:')) {
					return;
				}
				
				const text = (link.textContent || '').toLowerCase();
				const href = link.href.toLowerCase();
				
				// Check if the link text contains any of our keywords
				for (const keyword of keywords) {
					if (text.includes(keyword) || href.includes(keyword)) {
						signupLinks.push({
							url: link.href,
							text: link.textContent.trim(),
							isHeader: true,
							relevanceScore: calculateRelevanceScore(text, href, keyword)
						});
						break;
					}
				}
			});
		});
		
		// If we didn't find any in headers, look for links in the main content that match
		if (signupLinks.length === 0) {
			const allLinks = Array.from(document.querySelectorAll('a'));
			
			allLinks.forEach(link => {
				// Skip if no href or it's just a hash or javascript
				if (!link.href || link.href === '' || link.href.startsWith('#') || 
					link.href.startsWith('javascript:') || link.href.includes('mailto:')) {
					return;
				}
				
				const text = (link.textContent || '').toLowerCase();
				const href = link.href.toLowerCase();
				
				// Check if the link text contains any of our keywords
				for (const keyword of keywords) {
					if (text.includes(keyword) || href.includes(keyword)) {
						signupLinks.push({
							url: link.href,
							text: link.textContent.trim(),
							isHeader: false,
							relevanceScore: calculateRelevanceScore(text, href, keyword) * 0.8 // Lower score for non-header links
						});
						break;
					}
				}
			});
		}
		
		// Calculate a relevance score based on how likely this is a tournament signup link
		function calculateRelevanceScore(text, href, matchedKeyword) {
			let score = 1.0;
			
			// Higher score for more specific keywords
			const highValueKeywords = ['inscription', 'signup', 'register', 's\'inscrire'];
			if (highValueKeywords.includes(matchedKeyword)) {
				score += 0.5;
			}
			
			// Higher score if the text explicitly mentions tournament or competition
			if (text.includes('tournoi') || text.includes('tournament') || 
				text.includes('compétition') || text.includes('competition')) {
				score += 0.5;
			}
			
			// Higher score if the link has current year
			const currentYear = new Date().getFullYear();
			if (text.includes(currentYear.toString()) || href.includes(currentYear.toString())) {
				score += 0.5;
			}
			
			// Higher score if the keyword appears in URL
			if (href.includes(matchedKeyword)) {
				score += 0.3;
			}
			
			// Lower score for login links (might be user accounts, not signup)
			if (text.includes('login') || text.includes('connexion')) {
				score -= 0.3;
			}
			
			return score;
		}
		
		// Sort links by relevance score in descending order
		return signupLinks.sort((a, b) => b.relevanceScore - a.relevanceScore);
	}
	`

	// Execute the JavaScript to find signup links
	result, err := page.Evaluate(jsScript)
	if err != nil {
		return "", fmt.Errorf("failed to evaluate JavaScript: %w", err)
	}

	// Parse the result
	links := []map[string]interface{}{}
	if resultArr, ok := result.([]interface{}); ok {
		for _, item := range resultArr {
			if linkObj, ok := item.(map[string]interface{}); ok {
				links = append(links, linkObj)
			}
		}
	}

	// Process found links
	if len(links) == 0 {
		return "", nil
	}

	// Create a mock tournament with the site domain as the club name
	// This helps with validation when checking if the page is a signup form
	domain := strings.TrimPrefix(websiteURL, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	if strings.Contains(domain, "/") {
		domain = strings.Split(domain, "/")[0]
	}

	mockTournament := cache.TournamentCache{
		Name:      "Tournament " + time.Now().Format("2006"),
		StartDate: time.Now().AddDate(0, 1, 0).Format("2006-01-02"), // 1 month from now
		Club: cache.Club{
			Name: domain,
		},
	}

	tournamentDate := time.Now().AddDate(0, 1, 0) // 1 month from now

	// For each link, create a new page and check if it's actually a signup form
	for _, link := range links {
		url, _ := link["url"].(string)
		score, _ := link["relevanceScore"].(float64)

		if url == "" {
			continue
		}

		// Create a new page for this link to avoid conflicts with the main page
		linkPage, err := browserContext.NewPage()
		if err != nil {
			continue
		}

		// Navigate to the potential signup URL
		_, err = linkPage.Goto(url, pw.PageGotoOptions{
			WaitUntil: pw.WaitUntilStateNetworkidle,
			Timeout:   pw.Float(15 * 1000),
		})

		if err != nil {
			linkPage.Close()
			continue
		}

		// Check if this page is a signup form
		isSignupForm, err := checkIfPageIsSignupForm(linkPage, mockTournament, tournamentDate)

		// Always close the page when done
		linkPage.Close()

		if err != nil {
			continue
		}

		if isSignupForm {
			return url, nil
		} else {
			// If it's not a signup form but has high relevance,
			// try to search for signup links on that page too (one level recursion)
			if score > 1.5 {
				recursiveURL, err := CheckWebsiteHeaderForSignupLink(url, browserContext)
				if err == nil && recursiveURL != "" {
					return recursiveURL, nil
				}
			}
		}
	}

	// No valid signup links found
	return "", nil
}

// IsDomainToSkip checks if a domain should be skipped during validation
func IsDomainToSkip(domain string) bool {
	domain = strings.ToLower(domain)
	for _, skipDomain := range skipDomains {
		if strings.Contains(domain, skipDomain) {
			return true
		}
	}
	return false
}

// IsURLToSkip checks if a URL should be skipped based on common patterns
func IsURLToSkip(urlStr string) bool {
	urlLower := strings.ToLower(urlStr)
	for _, pattern := range skipWords {
		if strings.Contains(urlLower, pattern) {
			return true
		}
	}
	return false
}

// ValidateGenericSignupURL validates a generic URL as a potential signup form
func ValidateGenericSignupURL(urlStr string, tournament cache.TournamentCache, tournamentDate time.Time, browserContext pw.BrowserContext) (string, error) {
	// Create a new page for validation
	page, err := browserContext.NewPage()
	if err != nil {
		return "", fmt.Errorf("failed to create page: %w", err)
	}
	defer page.Close()

	// Set a timeout for navigation
	page.SetDefaultTimeout(20000) // 20 seconds

	utils.DebugLog("Navigating to URL: %s", urlStr)

	// Navigate to the URL
	resp, err := page.Goto(urlStr, pw.PageGotoOptions{
		WaitUntil: pw.WaitUntilStateNetworkidle,
	})

	if err != nil {
		return "", fmt.Errorf("failed to navigate to URL: %w", err)
	}

	// Check response status
	if resp.Status() >= 400 {
		return "", fmt.Errorf("page returned error status: %d", resp.Status())
	}

	utils.DebugLog("Successfully loaded URL: %s", urlStr)

	// Get the final URL after any redirects
	finalURL := page.URL()

	// Check if page contains form elements
	hasFormElements, err := detectFormElements(page)
	if err != nil {
		utils.DebugLog("Error detecting form elements: %v", err)
	}

	// Check if page contains signup keywords
	hasSignupKeywords, err := detectSignupKeywords(page)
	if err != nil {
		utils.DebugLog("Error detecting signup keywords: %v", err)
	}

	// Check if the page contains the tournament name or date
	containsTournamentInfo, err := containsTournamentInfo(page, tournament, tournamentDate)
	if err != nil {
		utils.DebugLog("Error checking for tournament info: %v", err)
	}

	// Log findings
	utils.DebugLog("URL %s validation: hasForm=%v, hasKeywords=%v, hasTournamentInfo=%v",
		urlStr, hasFormElements, hasSignupKeywords, containsTournamentInfo)

	// If the page has form elements or signup keywords and tournament info, it's likely a signup form
	if (hasFormElements || hasSignupKeywords) && containsTournamentInfo {
		log.Printf("Found valid signup URL: %s", finalURL)
		return finalURL, nil
	}

	// If we have strong form indicators but couldn't find tournament info,
	// we'll return it as a potential match but with a warning
	if hasFormElements && hasSignupKeywords {
		log.Printf("Found potential signup URL (missing tournament info): %s", finalURL)
		return finalURL, nil
	}

	return "", fmt.Errorf("URL does not appear to be a signup form")
}

// detectFormElements checks if the page contains form elements
func detectFormElements(page pw.Page) (bool, error) {
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

// detectSignupKeywords checks if the page contains signup-related keywords
func detectSignupKeywords(page pw.Page) (bool, error) {
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

// containsTournamentInfo checks if the page contains information about the tournament
func containsTournamentInfo(page pw.Page, tournament cache.TournamentCache, tournamentDate time.Time) (bool, error) {
	// Get the tournament name and normalize it for comparison
	tournamentName := strings.ToLower(tournament.Name)

	// Try to extract words from the tournament name for partial matching
	nameTokens := strings.Fields(tournamentName)

	// Generate date strings in various formats
	dateStrings := generateDateStrings(tournamentDate)

	jsScript := fmt.Sprintf(`
	() => {
		const pageText = document.body.innerText.toLowerCase();
		
		// Check tournament name (exact match)
		const tournamentName = %q;
		if (pageText.includes(tournamentName)) {
			return true;
		}
		
		// Check tournament name tokens (partial match)
		const nameTokens = %v;
		let tokenMatches = 0;
		
		for (const token of nameTokens) {
			if (token.length >= 4 && pageText.includes(token)) { // Only check tokens of meaningful length
				tokenMatches++;
			}
		}
		
		// If multiple tokens match, it's likely related to the tournament
		if (tokenMatches >= 2 || (nameTokens.length === 1 && tokenMatches === 1)) {
			return true;
		}
		
		// Check for dates
		const dateStrings = %v;
		for (const dateStr of dateStrings) {
			if (pageText.includes(dateStr)) {
				return true;
			}
		}
		
		return false;
	}
	`, tournamentName, nameTokens, dateStrings)

	result, err := page.Evaluate(jsScript)
	if err != nil {
		return false, fmt.Errorf("failed to evaluate script: %w", err)
	}

	if hasTournamentInfo, ok := result.(bool); ok {
		return hasTournamentInfo, nil
	}

	return false, nil
}

// generateDateStrings creates a list of date string formats for the given date
func generateDateStrings(date time.Time) []string {
	dateStrings := []string{
		date.Format("02/01/2006"),
		date.Format("2006-01-02"),
		date.Format("January 2, 2006"),
		date.Format("2 January 2006"),
		fmt.Sprintf("%d/%d/%d", date.Day(), date.Month(), date.Year()),
		fmt.Sprintf("%d %s %d", date.Day(), date.Month().String(), date.Year()),
		fmt.Sprintf("%d %s", date.Day(), date.Month().String()),
		fmt.Sprintf("%d/%d", date.Day(), date.Month()),
	}

	// Add French month names
	frenchMonths := map[time.Month]string{
		time.January:   "janvier",
		time.February:  "février",
		time.March:     "mars",
		time.April:     "avril",
		time.May:       "mai",
		time.June:      "juin",
		time.July:      "juillet",
		time.August:    "août",
		time.September: "septembre",
		time.October:   "octobre",
		time.November:  "novembre",
		time.December:  "décembre",
	}

	frenchDateStrings := []string{
		fmt.Sprintf("%d %s %d", date.Day(), frenchMonths[date.Month()], date.Year()),
		fmt.Sprintf("%d %s", date.Day(), frenchMonths[date.Month()]),
	}

	dateStrings = append(dateStrings, frenchDateStrings...)

	return dateStrings
}
