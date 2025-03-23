// Package navigation provides services for navigating tournament websites and finding registration forms
package navigation

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/utils"

	pw "github.com/playwright-community/playwright-go"
)

// Configuration constants
const (
	// Maximum number of URLs to check in recursive navigation
	MaxURLsToCheck = 10
	// Maximum recursion depth
	MaxRecursionDepth = 2
)

// RecursivelyFindRegistrationForm navigates recursively through URLs looking for registration forms
func RecursivelyFindRegistrationForm(baseURL string, tournament cache.TournamentCache, tournamentDate time.Time, browserContext pw.BrowserContext) (string, error) {
	// Parse the URL to extract domain for subdomain generation
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	// Generate potential URLs to check
	urlsToCheck := []string{baseURL}

	// Add common tournament subdomains if we're checking a base domain
	host := parsedURL.Host
	if !strings.HasPrefix(host, "www.") && !strings.Contains(host, "tournament") && !strings.Contains(host, "tournoi") {
		// Generate common tournament subdomains
		subdomains := []string{
			"www." + host,
			"tournament." + host,
			"tournoi." + host,
			"inscriptions." + host,
			"signup." + host,
			"register." + host,
		}

		for _, subdomain := range subdomains {
			subdURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, subdomain)
			urlsToCheck = append(urlsToCheck, subdURL)
		}
	}

	// Add common tournament paths
	paths := []string{
		"/tournament", "/tournaments", "/tournoi", "/tournois",
		"/inscription", "/inscriptions", "/register", "/signup",
		"/events", "/evenements", "/competitions", "/competition",
	}

	// Check if the base URL already has a path
	hasPath := len(parsedURL.Path) > 1 // More than just "/"
	if !hasPath {
		// Only add path variants if the base URL doesn't already have a significant path
		for _, path := range paths {
			pathURL := fmt.Sprintf("%s://%s%s", parsedURL.Scheme, parsedURL.Host, path)
			urlsToCheck = append(urlsToCheck, pathURL)
		}
	}

	// Try to find a registration form by recursively following links
	return SearchRegistrationFormRecursively(urlsToCheck, tournament, tournamentDate, browserContext, 0)
}

// SearchRegistrationFormRecursively attempts to find a registration form by recursively following links
func SearchRegistrationFormRecursively(urlsToCheck []string, tournament cache.TournamentCache, tournamentDate time.Time, browserContext pw.BrowserContext, depth int) (string, error) {
	// Limit recursion depth to prevent infinite loops
	if depth > MaxRecursionDepth {
		return "", fmt.Errorf("maximum recursion depth reached")
	}

	// Limit the number of URLs to check
	if len(urlsToCheck) > MaxURLsToCheck {
		urlsToCheck = urlsToCheck[:MaxURLsToCheck]
	}

	for _, urlToCheck := range urlsToCheck {
		// Skip if already checked or obvious social media URLs
		if utils.IsURLToSkip(urlToCheck) {
			continue
		}

		utils.DebugLog("Checking URL for registration form: %s", urlToCheck)

		// First, try direct validation
		signupURL, err := ValidateSignupURL(urlToCheck, tournament, tournamentDate, browserContext)
		if err == nil && signupURL != "" {
			log.Printf("Found registration form at %s", signupURL)
			return signupURL, nil
		}

		// If direct validation fails, look for links on the page
		newLinks, err := GatherRegistrationLinksFromPage(urlToCheck, browserContext)
		if err != nil {
			utils.DebugLog("Error finding links on %s: %v", urlToCheck, err)
			continue
		}

		if len(newLinks) > 0 {
			utils.DebugLog("Found %d potential registration links on %s", len(newLinks), urlToCheck)

			// Recursively check these links
			result, err := SearchRegistrationFormRecursively(newLinks, tournament, tournamentDate, browserContext, depth+1)
			if err == nil && result != "" {
				return result, nil
			}
		}
	}

	return "", fmt.Errorf("no registration form found")
}

// GatherRegistrationLinksFromPage finds potential registration links from a page
func GatherRegistrationLinksFromPage(pageURL string, browserContext pw.BrowserContext) ([]string, error) {
	// Create a new page
	page, err := browserContext.NewPage()
	if err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}
	defer page.Close()

	// Navigate to the URL
	if _, err := page.Goto(pageURL, pw.PageGotoOptions{
		WaitUntil: pw.WaitUntilStateNetworkidle,
		Timeout:   pw.Float(DefaultPageTimeout),
	}); err != nil {
		return nil, fmt.Errorf("failed to navigate to %s: %w", pageURL, err)
	}

	// Find all links on the page
	jsScript := `
	() => {
		const links = Array.from(document.querySelectorAll('a'));
		return links.map(link => {
			return {
				href: link.href,
				text: link.innerText.trim(),
				hasRegistrationKeyword: /inscription|register|signup|enroll|participate|engage|tournament|tournoi|competition/.test(link.innerText.toLowerCase())
			};
		}).filter(link => link.href && link.href.startsWith('http'));
	}
	`

	result, err := page.Evaluate(jsScript)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate script: %w", err)
	}

	// Process the results
	var links []string
	if resultSlice, ok := result.([]interface{}); ok {
		for _, item := range resultSlice {
			if linkData, ok := item.(map[string]interface{}); ok {
				linkHref, ok := linkData["href"].(string)
				if !ok {
					continue
				}

				hasKeyword, _ := linkData["hasRegistrationKeyword"].(bool)
				linkText, _ := linkData["text"].(string)

				// Prioritize links with registration keywords
				if hasKeyword {
					links = append(links, linkHref)
				} else if strings.Contains(strings.ToLower(linkText), "tournament") ||
					strings.Contains(strings.ToLower(linkText), "tournoi") ||
					strings.Contains(strings.ToLower(linkText), "competition") ||
					strings.Contains(strings.ToLower(linkText), "compétition") {
					links = append(links, linkHref)
				}
			}
		}
	}

	return links, nil
}
