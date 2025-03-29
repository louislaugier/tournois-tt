// Package utils provides utility functions
// used throughout the application
package utils

import (
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
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

	// SignupURLRegex matches text about signup on a tournament website
	SignupURLRegex = regexp.MustCompile(`(?i)(?:inscription|s'inscrire|créer un compte|engagement|engagements|étape suivante)[^\n]*(?:sur|en ligne)[^\n]*(?:https?://)?(?:tournoi\.)?([a-zA-Z0-9][-a-zA-Z0-9]*\.)+[a-zA-Z]{2,}`)

	// Mutex for checked URLs
	checkedURLsMutex sync.RWMutex
	checkedURLs      = make(map[string]bool)
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
	// Handle URLs without protocol
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		urlStr = "https://" + urlStr
	}

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

// MarkURLAsChecked adds a URL to the list of checked URLs
func MarkURLAsChecked(url string) {
	checkedURLsMutex.Lock()
	checkedURLs[url] = true
	checkedURLsMutex.Unlock()
}

// HasURLBeenChecked checks if a URL has already been processed
func HasURLBeenChecked(url string) bool {
	checkedURLsMutex.RLock()
	checked := checkedURLs[url]
	checkedURLsMutex.RUnlock()
	return checked
}

// CheckAndMarkURL checks if a URL should be skipped based on predefined rules
// and marks it as checked
func CheckAndMarkURL(urlStr string) bool {
	// Skip empty URLs
	if urlStr == "" {
		return true
	}

	// Skip already checked URLs
	if HasURLBeenChecked(urlStr) {
		return true
	}

	// Mark this URL as checked
	MarkURLAsChecked(urlStr)

	return IsURLToSkip(urlStr)
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

// StringSliceContains checks if a string slice contains a specific value
func StringSliceContains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// IsCommonWord checks if a word is a common word that should be filtered
func IsCommonWord(word string) bool {
	commonWords := []string{
		"the", "and", "for", "are", "but", "not", "you", "all", "any",
		"can", "had", "her", "was", "one", "our", "out", "day", "get",
		"has", "him", "his", "how", "man", "new", "now", "old", "see",
		"two", "way", "who", "boy", "did", "its", "let", "put", "say",
		"she", "too", "use", "les", "des", "est", "sur", "par", "avec",
	}

	return StringSliceContains(commonWords, strings.ToLower(word))
}

// ExtractSignificantWords extracts significant words from a text string,
// filtering out common words and short terms
func ExtractSignificantWordsFromText(text string) []string {
	// Map of common words to filter out
	commonWords := map[string]bool{
		"the": true, "and": true, "for": true, "with": true, "this": true,
		"that": true, "are": true, "from": true, "your": true, "have": true,
		"more": true, "will": true, "home": true, "can": true, "page": true,
		"you": true, "was": true, "all": true, "has": true, "but": true,
		"one": true, "what": true, "out": true, "when": true, "use": true,
		"les": true, "des": true, "une": true, "est": true, "pas": true,
		"par": true, "sur": true, "sous": true, "avec": true, "sans": true, "club": true,
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

// EnsureURLProtocol ensures that a URL has a protocol, defaulting to https
func EnsureURLProtocol(url string) string {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "https://" + url
	}
	return url
}

// FindDomainOnlyReferences finds domain references for tournament websites mentioned in text
func FindDomainOnlyReferences(text string) []string {
	// This is a simplified version without constants dependency
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

// EncodeURL properly encodes a URL string using the standard library
func EncodeURL(urlStr string) (string, error) {
	// Parse the URL
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	// Let the standard library handle the encoding through String()
	return u.String(), nil
}
