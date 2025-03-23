package pdf_processor

import (
	"fmt"
	"strings"
	"time"

	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/pdf"
	"tournois-tt/api/pkg/scraper/services/common"
	"tournois-tt/api/pkg/utils"

	pw "github.com/playwright-community/playwright-go"
)

// tryRecursiveFormNavigation recursively navigates through a list of URLs looking for registration forms
func tryRecursiveFormNavigation(urls []string, tournament cache.TournamentCache, tournamentDate time.Time, browserContext pw.BrowserContext,
	validator func(string, cache.TournamentCache, time.Time, pw.BrowserContext) (string, error)) (string, bool, error) {
	// To avoid excessive processing, limit the number of URLs to check
	expandedUrls := []string{}

	// For each URL, also add common tournament subdomains to try
	for _, url := range urls {
		// Add the original URL
		expandedUrls = append(expandedUrls, url)

		// Extract domain and generate subdomains to try
		domain := strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "http://")
		// Only generate subdomains if this looks like a base domain (no path components)
		if !strings.Contains(domain, "/") {
			subdomains := pdf.GenerateCommonTournamentSubdomains(domain)
			// Skip the first one which is just the original domain with https
			if len(subdomains) > 1 {
				for _, subdomain := range subdomains[1:] {
					if !common.Contains(expandedUrls, subdomain) {
						expandedUrls = append(expandedUrls, subdomain)
					}
				}
			}
		}
	}

	utils.DebugLog("Expanded %d URLs to %d URLs including potential subdomains",
		len(urls), len(expandedUrls))

	// Try each URL (original and subdomains) with recursive navigation
	for _, url := range expandedUrls {
		utils.DebugLog("Starting recursive form navigation from: %s", url)
		finalURL, found, err := findRegistrationFormRecursively(url, tournament, tournamentDate, browserContext, 0, validator)
		if err != nil {
			utils.DebugLog("Error in recursive form navigation for %s: %v", url, err)
			continue
		}

		if found {
			return finalURL, true, nil
		}
	}

	return "", false, nil
}

// findRegistrationFormRecursively recursively navigates through pages looking for a registration form
func findRegistrationFormRecursively(url string, tournament cache.TournamentCache, tournamentDate time.Time,
	browserContext pw.BrowserContext, depth int,
	validator func(string, cache.TournamentCache, time.Time, pw.BrowserContext) (string, error)) (string, bool, error) {

	// Check recursion depth limit
	if depth >= MaxRedirections {
		utils.DebugLog("Reached maximum recursion depth (%d) for URL: %s", MaxRedirections, url)
		return "", false, nil
	}

	utils.DebugLog("Checking for registration form at depth %d: %s", depth, url)

	// First, check if the current URL is a valid signup form directly
	validURL, err := validator(url, tournament, tournamentDate, browserContext)
	if err != nil {
		return "", false, fmt.Errorf("error validating URL %s: %w", url, err)
	}

	if validURL != "" {
		utils.DebugLog("Found valid signup form at depth %d: %s", depth, validURL)
		return validURL, true, nil
	}

	// If HTTPS navigation fails, we might want to try HTTP
	if strings.HasPrefix(url, "https://") && err != nil &&
		(strings.Contains(err.Error(), "ERR_CONNECTION_REFUSED") ||
			strings.Contains(err.Error(), "ERR_SSL_PROTOCOL_ERROR") ||
			strings.Contains(err.Error(), "ERR_CERT_") ||
			strings.Contains(err.Error(), "certificate")) {

		httpURL := "http://" + strings.TrimPrefix(url, "https://")
		utils.DebugLog("HTTPS navigation failed, retrying with HTTP: %s", httpURL)
		return findRegistrationFormRecursively(httpURL, tournament, tournamentDate, browserContext, depth, validator)
	}

	// If the current URL is not a signup form, try to find registration links on it
	page, err := browserContext.NewPage()
	if err != nil {
		return "", false, fmt.Errorf("failed to create page: %w", err)
	}
	defer page.Close()

	// Navigate to the URL with enhanced error handling
	_, err = page.Goto(url, pw.PageGotoOptions{
		Timeout:   pw.Float(15000), // 15 seconds timeout
		WaitUntil: pw.WaitUntilStateNetworkidle,
	})

	if err != nil {
		return "", false, fmt.Errorf("failed to navigate to URL: %w", err)
	}

	// Find registration-related links on the page
	links, err := findRegistrationLinksOnPage(page)
	if err != nil {
		return "", false, fmt.Errorf("failed to find registration links: %w", err)
	}

	// If we found links, recursively check each of them
	currentURL := page.URL() // Get the current URL after possible redirects

	utils.DebugLog("Found %d potential registration links on page %s", len(links), currentURL)

	for _, link := range links {
		utils.DebugLog("Following registration link at depth %d: %s", depth, link)
		validURL, found, err := findRegistrationFormRecursively(link, tournament, tournamentDate, browserContext, depth+1, validator)
		if err != nil {
			utils.DebugLog("Error following link %s: %v", link, err)
			continue
		}

		if found {
			return validURL, true, nil
		}
	}

	// No valid signup form found recursively
	return "", false, nil
}

// findRegistrationLinksOnPage finds registration-related links on a page
func findRegistrationLinksOnPage(page pw.Page) ([]string, error) {
	// JavaScript to find registration-related links on the page
	jsScript := `
	() => {
		const registrationKeywords = [
			'inscription', 'inscriptions', 'inscrire', "s'inscrire",
			'registre', 'enregistrer', "s'enregistrer",
			'tarif', 'tarifs', 'paiement', 'payer',
			'formulaire', 'form', 'registration', 'register', 'signup',
			'engagement', 'engagements',
			'etape suivante', 'étape suivante', 'suivant', 'continuer',
			'tournoi', 'tournament', 'competition', 'compétition'
		];
		
		const links = Array.from(document.querySelectorAll('a'));
		const result = [];
		
		for (const link of links) {
			// Skip if link has no href or is a fragment/javascript link
			if (!link.href || link.href === '' || link.href === '#' || 
				link.href.startsWith('javascript:') || link.href.includes('mailto:')) {
				continue;
			}
			
			// Get text content and normalize
			const text = (link.textContent || '').toLowerCase();
			const href = link.href.toLowerCase();
			const title = (link.getAttribute('title') || '').toLowerCase();
			
			// Check if link contains any registration keyword
			for (const keyword of registrationKeywords) {
				if (text.includes(keyword) || href.includes(keyword) || title.includes(keyword)) {
					result.push(link.href);
					break;
				}
			}
		}
		
		return result;
	}
	`

	// Execute the JavaScript to find registration-related links
	result, err := page.Evaluate(jsScript)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate JavaScript: %w", err)
	}

	// Parse the result into a string array
	var links []string
	if resultArray, ok := result.([]interface{}); ok {
		for _, item := range resultArray {
			if url, ok := item.(string); ok {
				links = append(links, url)
			}
		}
	}

	return links, nil
}
