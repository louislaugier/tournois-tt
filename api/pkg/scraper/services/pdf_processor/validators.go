package pdf_processor

import (
	"strings"
	"time"

	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/pdf"
	"tournois-tt/api/pkg/utils"

	pw "github.com/playwright-community/playwright-go"
)

// tryValidateURLs tries to validate a list of URLs as signup URLs
func tryValidateURLs(urls []string, tournament cache.TournamentCache, tournamentDate time.Time, browserContext pw.BrowserContext,
	validator func(string, cache.TournamentCache, time.Time, pw.BrowserContext) (string, error)) (string, bool) {
	// Limit the number of URLs to process to avoid excessive validation
	urls = pdf.LimitURLs(urls, MaxURLsToProcess)

	for _, url := range urls {
		// Clean up the URL
		cleanURL := pdf.EnsureURLProtocol(url)
		utils.DebugLog("Validating URL: %s", cleanURL)

		// Try to validate
		validURL, err := validator(cleanURL, tournament, tournamentDate, browserContext)
		if err != nil {
			utils.DebugLog("Error validating URL %s: %v", cleanURL, err)
			continue
		}

		if validURL != "" {
			return validURL, true
		}
	}

	return "", false
}

// findRegistrationURLsInPDF finds all URLs in PDF text that might be registration related
func findRegistrationURLsInPDF(text string) []string {
	// Extract all URLs from the PDF text
	allURLs := pdf.FindURLsInText(text)

	// Filter out URLs that are clearly not registration related
	var filteredURLs []string
	excludePatterns := []string{
		"google.com", "youtube.com", "facebook.com", "twitter.com",
		"instagram.com", "linkedin.com", "github.com", "apple.com",
		"microsoft.com", "amazon.com", "yahoo.com", "gmail.com",
		"outlook.com", "hotmail.com", "wikipedia.org", "mozilla.org",
		"adobe.com", "pinterest.com", "tiktok.com", "snapchat.com",
		"reddit.com", "tumblr.com", "flickr.com", "whatsapp.com",
		"telegram.org", "discord.com", "zoom.us", "teams.microsoft.com",
		"fftt.com",
	}

	for _, url := range allURLs {
		isExcluded := false
		for _, pattern := range excludePatterns {
			if strings.Contains(strings.ToLower(url), pattern) {
				isExcluded = true
				break
			}
		}

		if !isExcluded {
			filteredURLs = append(filteredURLs, url)
		}
	}

	return filteredURLs
}

// validateDomainURLs validates a list of domain URLs by checking for signup links on each domain
func validateDomainURLs(domains []string, tournament cache.TournamentCache, tournamentDate time.Time, browserContext pw.BrowserContext,
	validator func(string, cache.TournamentCache, time.Time, pw.BrowserContext) (string, error)) (string, error) {
	// Extract domains from the URLs
	var cleanDomains []string
	for _, url := range domains {
		utils.DebugLog("Validating domain URL: %s", url)
		// Try the domain URL directly
		validURL, err := validator(url, tournament, tournamentDate, browserContext)
		if err != nil {
			utils.DebugLog("Error validating domain URL %s: %v", url, err)
		} else if validURL != "" {
			return validURL, nil
		}

		// Clean up the domain for further processing
		domain := url
		if strings.HasPrefix(domain, "https://") {
			domain = strings.TrimPrefix(domain, "https://")
		} else if strings.HasPrefix(domain, "http://") {
			domain = strings.TrimPrefix(domain, "http://")
		}

		if strings.Contains(domain, "/") {
			domain = strings.Split(domain, "/")[0]
		}

		cleanDomains = append(cleanDomains, domain)
	}

	// For each domain, try a few common paths where signup forms might be found
	for _, domain := range cleanDomains {
		// Try common paths
		commonPaths := []string{
			"/inscription", "/inscriptions", "/signup", "/register",
			"/tournament", "/tournoi", "/competition", "/engagement",
			"/engagements", "/participer", "/form", "/formulaire",
		}

		for _, path := range commonPaths {
			pathURL := "https://" + domain + path
			utils.DebugLog("Validating domain URL with path: %s", pathURL)
			validURL, err := validator(pathURL, tournament, tournamentDate, browserContext)
			if err != nil {
				utils.DebugLog("Error validating domain URL with path %s: %v", pathURL, err)
				continue
			}

			if validURL != "" {
				return validURL, nil
			}
		}
	}

	return "", nil
}
