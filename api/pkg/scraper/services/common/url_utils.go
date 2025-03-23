package common

import (
	"path/filepath"
	"regexp"
	"strings"
)

// -----------------------------------------------------------------------------
// URL Pattern Constants
// -----------------------------------------------------------------------------

var (
	// URLRegex matches any URL
	URLRegex = regexp.MustCompile(`https?://[^\s"']+`)

	// TournoiSubdomainRegex matches URLs with "tournoi" subdomain
	TournoiSubdomainRegex = regexp.MustCompile(`\b(?:https?://)?tournoi\.([a-zA-Z0-9][-a-zA-Z0-9]*\.)+[a-zA-Z]{2,}\b`)

	// PaymentURLRegex matches text about payment on a tournament website
	PaymentURLRegex = regexp.MustCompile(`(?i)paiement[^\n]*(?:sur|en ligne)[^\n]*(?:https?://)?(?:tournoi\.)?([a-zA-Z0-9][-a-zA-Z0-9]*\.)+[a-zA-Z]{2,}`)

	// SignupURLRegex matches text about signup on a tournament website - prioritized over payment
	SignupURLRegex = regexp.MustCompile(`(?i)(?:inscription|s'inscrire|créer un compte|engagement|engagements|etape suivante|étape suivante)[^\n]*(?:sur|en ligne)[^\n]*(?:https?://)?(?:tournoi\.)?([a-zA-Z0-9][-a-zA-Z0-9]*\.)+[a-zA-Z]{2,}`)

	// RegistrationKeywords is a list of keywords related to registration
	RegistrationKeywords = []string{
		"inscription", "inscriptions", "inscrire", "s'inscrire",
		"registre", "enregistrer", "s'enregistrer",
		"tarif", "tarifs", "paiement", "payer",
		"formulaire", "form", "registration", "register", "signup",
		"engagement", "engagements",
		"etape suivante", "étape suivante", "suivant", "continuer",
	}
)

// -----------------------------------------------------------------------------
// URL Extraction and Validation Functions
// -----------------------------------------------------------------------------

// FindURLsByPattern extracts URLs from text matching a specific regex pattern
func FindURLsByPattern(text string, pattern *regexp.Regexp) []string {
	return pattern.FindAllString(text, -1)
}

// FindURLsInText extracts all URLs from text content
func FindURLsInText(text string) []string {
	return FindURLsByPattern(text, URLRegex)
}

// IsPDFFile checks if a URL points to a PDF file
func IsPDFFile(urlStr string) bool {
	// Check file extension
	ext := strings.ToLower(filepath.Ext(urlStr))
	isPDF := ext == ".pdf"

	// Also check for URLs with "pdf" in the path or query parameters
	if !isPDF {
		isPDF = regexp.MustCompile(`[\/?]pdf[\/?]`).MatchString(strings.ToLower(urlStr)) ||
			strings.Contains(strings.ToLower(urlStr), "format=pdf") ||
			strings.Contains(strings.ToLower(urlStr), "type=pdf")
	}

	return isPDF
}

// IsExtractionRetryableError determines if a PDF extraction error can be retried
func IsExtractionRetryableError(errStr string) bool {
	// Check for typical temporary PDF extraction error patterns
	retryablePatterns := []string{
		"timeout",
		"connection reset",
		"connection refused",
		"temporary",
		"network",
		"stream error",
		"EOF",
		"unexpected EOF",
		"HTTP status",
		"TLS handshake",
		"download",
		"i/o timeout",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(strings.ToLower(errStr), strings.ToLower(pattern)) {
			return true
		}
	}

	return false
}

// FindDomainOnlyReferences finds domain references for tournament websites mentioned in text
func FindDomainOnlyReferences(text string) []string {
	// Match patterns like:
	// - "inscriptions sur cctt.fr"
	// - "s'inscrire en ligne sur www.example.com"
	domainRegexes := []*regexp.Regexp{
		// Pattern 1: inscription(s) sur domain.tld
		regexp.MustCompile(`(?i)inscription(?:s)?[^\n.]{1,30}(?:sur|en ligne)[^\n.]{1,30}((?:https?://)?(?:www\.)?(?:[a-zA-Z0-9][-a-zA-Z0-9]*\.)+[a-zA-Z]{2,})`),

		// Pattern 2: A PRIVILEGIER mentions
		regexp.MustCompile(`(?i)(?:A PRIVILEGIER|À PRIVILÉGIER)[^\n.]{1,50}((?:https?://)?(?:www\.)?(?:[a-zA-Z0-9][-a-zA-Z0-9]*\.)+[a-zA-Z]{2,})`),

		// Pattern 3: engagement(s) sur domain.tld
		regexp.MustCompile(`(?i)engagement(?:s)?[^\n.]{1,30}(?:sur|en ligne)[^\n.]{1,30}((?:https?://)?(?:www\.)?(?:[a-zA-Z0-9][-a-zA-Z0-9]*\.)+[a-zA-Z]{2,})`),

		// Pattern 4: simple domain.tld (with inscription nearby)
		regexp.MustCompile(`(?i)inscription(?:s)?[^\n]{1,100}((?:https?://)?(?:www\.)?(?:[a-zA-Z0-9][-a-zA-Z0-9]*\.)+[a-zA-Z]{2,})`),
	}

	var domains []string
	seen := make(map[string]bool)

	// Apply each regex pattern and merge results
	for _, regex := range domainRegexes {
		matches := regex.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			if len(match) > 1 && match[1] != "" {
				domain := match[1]
				// Ensure domain has a protocol
				if !strings.HasPrefix(domain, "http") {
					domain = "https://" + domain
				}
				if !seen[domain] {
					seen[domain] = true
					domains = append(domains, domain)
				}
			}
		}
	}

	return domains
}

// EnsureURLProtocol ensures that a URL has a protocol, defaulting to https
func EnsureURLProtocol(url string) string {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "https://" + url
	}
	return url
}

// GenerateCommonTournamentSubdomains generates common tournament subdomains from a base domain
func GenerateCommonTournamentSubdomains(domain string) []string {
	// Clean the input domain - strip protocol and www
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "www.")

	// If the domain already has a path component, don't try to create subdomains
	if strings.Contains(domain, "/") {
		return []string{"https://" + domain}
	}

	// Create variations of the domain with the "tournoi" subdomain
	variations := []string{
		"https://" + domain,
		"https://tournoi." + domain,
		"https://inscription." + domain,
		"https://inscriptions." + domain,
		"https://register." + domain,
		"https://signup." + domain,
		"https://www." + domain,
	}

	return variations
}

// LimitURLs limits the number of URLs to a maximum count
func LimitURLs(urls []string, maxCount int) []string {
	if len(urls) <= maxCount {
		return urls
	}
	return urls[:maxCount]
}

// IsNavigationError determines if an error string represents a navigation error
// that can be retried rather than a critical browser error
func IsNavigationError(errStr string) bool {
	// Check for typical navigation error patterns
	navigationErrorPatterns := []string{
		"timeout",
		"navigation",
		"navigate",
		"Frame.Goto",
		"Page.Goto",
		"could not navigate",
		"network error",
		"net::ERR",
		"page crashed",
		"connection reset",
		"connection refused",
	}

	for _, pattern := range navigationErrorPatterns {
		if strings.Contains(strings.ToLower(errStr), strings.ToLower(pattern)) {
			return true
		}
	}

	return false
}
