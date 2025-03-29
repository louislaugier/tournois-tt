// Package finder provides services for navigating tournament websites and finding registration forms
package finder

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
	MaxRecursionDepth = 3
)

// RecursivelyFindRegistrationForm searches a tournament website recursively to find the registration form URL
func RecursivelyFindRegistrationForm(baseURL string, tournament cache.TournamentCache, tournamentDate time.Time, browserContext pw.BrowserContext) (string, error) {
	utils.DebugLog("Starting recursive search for registration form on: %s", baseURL)

	// Skip if URL is known to be irrelevant
	if utils.IsURLToSkip(baseURL) {
		return "", fmt.Errorf("URL is in skip list: %s", baseURL)
	}

	// First, check if the baseURL itself is a registration form
	registrationURL, err := ValidateSignupURL(baseURL, tournament, tournamentDate, browserContext)
	if err == nil && registrationURL != "" {
		utils.DebugLog("Base URL is already a registration form: %s", baseURL)
		return registrationURL, nil
	}

	// Check the website's header navigation for signup links
	headerSignupURL, err := CheckWebsiteHeaderForSignupLink(baseURL, browserContext)
	if err == nil && headerSignupURL != "" {
		utils.DebugLog("Found signup link in header: %s", headerSignupURL)
		// Validate the found URL
		if validatedURL, err := ValidateSignupURL(headerSignupURL, tournament, tournamentDate, browserContext); err == nil {
			return validatedURL, nil
		}
	}

	// Gather all links from the starting page
	links, err := GatherRegistrationLinksFromPage(baseURL, browserContext)
	if err != nil {
		return "", fmt.Errorf("failed to gather links from page: %w", err)
	}

	// Sort and filter links to check the most promising ones first
	urlsToCheck := FilterRelevantLinks(links)
	if len(urlsToCheck) > MaxURLsToCheck {
		urlsToCheck = urlsToCheck[:MaxURLsToCheck]
	}

	// Start recursive search
	return SearchRegistrationFormRecursively(urlsToCheck, tournament, tournamentDate, browserContext, 1)
}

// FilterRelevantLinks filters and sorts links by their relevance to registration
func FilterRelevantLinks(links []string) []string {
	// First, filter out irrelevant links
	var filteredLinks []string
	for _, link := range links {
		if !utils.IsURLToSkip(link) && !utils.IsDomainToSkip(utils.ExtractDomain(link)) {
			filteredLinks = append(filteredLinks, link)
		}
	}

	// Then sort by relevance
	return PrioritizeLinksByRelevance(filteredLinks)
}

// SearchRegistrationFormRecursively searches links recursively to find a registration form
func SearchRegistrationFormRecursively(urlsToCheck []string, tournament cache.TournamentCache, tournamentDate time.Time, browserContext pw.BrowserContext, depth int) (string, error) {
	if depth > MaxRecursionDepth {
		return "", fmt.Errorf("reached maximum recursion depth")
	}

	utils.DebugLog("Checking %d URLs at depth %d", len(urlsToCheck), depth)

	for _, linkURL := range urlsToCheck {
		// Skip if URL is known to be irrelevant
		if utils.IsURLToSkip(linkURL) {
			continue
		}

		utils.DebugLog("Checking URL: %s", linkURL)

		// Check if this URL is a registration form
		registrationURL, err := ValidateSignupURL(linkURL, tournament, tournamentDate, browserContext)
		if err == nil && registrationURL != "" {
			utils.DebugLog("Found registration form at: %s", registrationURL)
			return registrationURL, nil
		}

		// If not a registration form, gather links from this page and continue the search
		if depth < MaxRecursionDepth {
			subLinks, err := GatherRegistrationLinksFromPage(linkURL, browserContext)
			if err != nil {
				log.Printf("Error gathering links from %s: %v", linkURL, err)
				continue
			}

			// Filter links we've already seen or that are irrelevant
			newLinks := []string{}
			for _, subLink := range subLinks {
				alreadyChecked := false
				for _, checkedURL := range urlsToCheck {
					if subLink == checkedURL {
						alreadyChecked = true
						break
					}
				}
				if !alreadyChecked && !utils.IsURLToSkip(subLink) {
					newLinks = append(newLinks, subLink)
				}
			}

			// Sort and filter links to check the most promising ones first
			relevantLinks := FilterRelevantLinks(newLinks)
			if len(relevantLinks) > MaxURLsToCheck {
				relevantLinks = relevantLinks[:MaxURLsToCheck]
			}

			if len(relevantLinks) > 0 {
				result, err := SearchRegistrationFormRecursively(relevantLinks, tournament, tournamentDate, browserContext, depth+1)
				if err == nil {
					return result, nil
				}
			}
		}
	}

	return "", fmt.Errorf("no registration form found")
}

// GatherRegistrationLinksFromPage collects links from a page that might lead to registration forms
func GatherRegistrationLinksFromPage(pageURL string, browserContext pw.BrowserContext) ([]string, error) {
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
		return nil, fmt.Errorf("failed to navigate to URL: %w", err)
	}

	// Extract all links from the page
	links, err := page.Evaluate(`
	() => {
		const links = Array.from(document.querySelectorAll('a')).map(a => a.href).filter(href => href.startsWith('http'));
		return [...new Set(links)]; // Remove duplicates
	}
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to extract links: %w", err)
	}

	// Convert the result to a string slice
	var results []string
	if linksArr, ok := links.([]interface{}); ok {
		baseURLObj, _ := url.Parse(pageURL)
		baseHost := baseURLObj.Host

		for _, link := range linksArr {
			if linkStr, ok := link.(string); ok {
				// Skip empty or javascript links
				if linkStr == "" || strings.HasPrefix(linkStr, "javascript:") {
					continue
				}

				// Resolve relative URLs if needed
				if !strings.HasPrefix(linkStr, "http") {
					base, err := url.Parse(pageURL)
					if err != nil {
						continue
					}
					relative, err := url.Parse(linkStr)
					if err != nil {
						continue
					}
					absolute := base.ResolveReference(relative)
					linkStr = absolute.String()
				}

				// Skip social media, print, email links, etc.
				if utils.IsDomainToSkip(utils.ExtractDomain(linkStr)) {
					continue
				}

				// Prioritize links from the same domain
				linkURLObj, err := url.Parse(linkStr)
				if err == nil && linkURLObj.Host == baseHost {
					results = append(results, linkStr)
				} else {
					// Add external links too, but they'll be processed later
					results = append(results, linkStr)
				}
			}
		}
	}

	// Sort links by relevance (prioritizing registration-related links)
	relevantLinks := SortLinksByRelevance(results)

	return relevantLinks, nil
}

// SortLinksByRelevance sorts links by their relevance to registration forms
func SortLinksByRelevance(links []string) []string {
	// First use the PrioritizeLinksByRelevance function to get a relevance-sorted list
	return PrioritizeLinksByRelevance(links)
}
