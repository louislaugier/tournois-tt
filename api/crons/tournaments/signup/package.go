// Package signup provides functionality for finding and validating signup URLs for tournaments
package signup

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"tournois-tt/api/pkg/scraper/browser"
	"tournois-tt/api/pkg/scraper/services/helloasso"

	pw "github.com/playwright-community/playwright-go"
)

// Constants for signup processing
const (
	numWorkers = 3 // Number of concurrent workers for signup URL refresh
)

// Create a regex pattern that uses the constant
var (
	// urlRegex matches any URL
	urlRegex = regexp.MustCompile(`https?://[^\s"']+`)

	// helloAssoURLRegex is a regex to find HelloAsso URLs
	helloAssoURLRegex = regexp.MustCompile(`https?://(?:www\.)?` + 
		strings.TrimPrefix(strings.ReplaceAll(helloasso.BaseURL, ".", "\\."), "https://") + 
		`/[^\s"']+`)

	// tournoiSubdomainRegex matches URLs with "tournoi" subdomain
	tournoiSubdomainRegex = regexp.MustCompile(`\b(?:https?://)?tournoi\.([a-zA-Z0-9][-a-zA-Z0-9]*\.)+[a-zA-Z]{2,}\b`)

	// paymentURLRegex matches text about payment on a tournament website
	paymentURLRegex = regexp.MustCompile(`(?i)paiement[^\n]*(?:sur|en ligne)[^\n]*(?:https?://)?(?:tournoi\.)?([a-zA-Z0-9][-a-zA-Z0-9]*\.)+[a-zA-Z]{2,}`)

	// signupURLRegex matches text about signup on a tournament website - prioritized over payment
	signupURLRegex = regexp.MustCompile(`(?i)(?:inscription|s'inscrire|créer un compte|engagement|engagements|etape suivante|étape suivante)[^\n]*(?:sur|en ligne)[^\n]*(?:https?://)?(?:tournoi\.)?([a-zA-Z0-9][-a-zA-Z0-9]*\.)+[a-zA-Z]{2,}`)

	// registrationKeywords is a list of keywords related to registration
	registrationKeywords = []string{
		"inscription", "inscriptions", "inscrire", "s'inscrire",
		"registre", "enregistrer", "s'enregistrer",
		"tarif", "tarifs", "paiement", "payer",
		"formulaire", "form", "registration", "register", "signup",
		"engagement", "engagements",
		"etape suivante", "étape suivante", "suivant", "continuer",
	}
	
	// RegistrationKeywords is the exported version of registrationKeywords for tests
	RegistrationKeywords = registrationKeywords
)

// GetURLRegex returns the URL regex pattern for use in tests
func GetURLRegex() *regexp.Regexp {
	return urlRegex
}

// GetHelloAssoURLRegex returns the HelloAsso URL regex pattern for use in tests
func GetHelloAssoURLRegex() *regexp.Regexp {
	return helloAssoURLRegex
}

// GetTournoiSubdomainRegex returns the regex for finding tournoi subdomain URLs
func GetTournoiSubdomainRegex() *regexp.Regexp {
	return tournoiSubdomainRegex
}

// GetPaymentURLRegex returns the regex for finding payment URL references
func GetPaymentURLRegex() *regexp.Regexp {
	return paymentURLRegex
}

// GetSignupURLRegex returns the regex for finding signup URL references
func GetSignupURLRegex() *regexp.Regexp {
	return signupURLRegex
}

// FindDomainOnlyReferences is the exported version of findDomainOnlyReferences for tests
func FindDomainOnlyReferences(text string) []string {
	return findDomainOnlyReferences(text)
}

// GenerateCommonTournamentSubdomains is the exported version of generateCommonTournamentSubdomains for tests
func GenerateCommonTournamentSubdomains(domain string) []string {
	return generateCommonTournamentSubdomains(domain)
}

// browserSetup initializes a shared browser instance for the signup URL refresh process
func browserSetup() (pw.Browser, *pw.Playwright, pw.BrowserContext, error) {
	log.Println("Setting up browser for refreshing signup URLs")
	// Initialize a shared browser instance
	cfg := browser.DefaultConfig()
	browserInstance, pwInstance, err := browser.Init(cfg)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize browser: %w", err)
	}

	// Create a browser context that will be shared among all workers
	browserContext, err := browser.NewContext(browserInstance, cfg)
	if err != nil {
		pwInstance.Stop()
		browserInstance.Close()
		return nil, nil, nil, fmt.Errorf("failed to create browser context: %w", err)
	}

	return browserInstance, pwInstance, browserContext, nil
}

// parseTournamentDate attempts to parse tournament dates from various formats
func parseTournamentDate(dateStr string) (time.Time, error) {
	// First try standard format
	date, err := time.Parse("2006-01-02", dateStr)
	if err == nil {
		return date, nil
	}

	// Try ISO 8601 format with time component (like "2024-06-08T00:00:00")
	date, err = time.Parse("2006-01-02T15:04:05", dateStr)
	if err == nil {
		return date, nil
	}

	// Try other common formats
	formats := []string{
		"02/01/2006",
		"2006/01/02",
		"01/02/2006",
		"Jan 2, 2006",
		"2 Jan 2006",
	}

	for _, format := range formats {
		date, err := time.Parse(format, dateStr)
		if err == nil {
			return date, nil
		}
	}

	return time.Time{}, fmt.Errorf("failed to parse date: %s", dateStr)
}

// Helper functions to avoid redeclarations

// ExtractSignificantWords extracts significant words from text (similar to the utils package function)
func ExtractSignificantWords(text string) []string {
	// Clean up the text
	text = strings.ToLower(text)
	words := strings.Fields(text)

	// Filter out common words and keep significant ones
	var significantWords []string
	stopWords := map[string]bool{
		"de": true, "du": true, "la": true, "le": true, "les": true,
		"des": true, "un": true, "une": true, "et": true, "à": true,
		"au": true, "aux": true, "en": true, "dans": true, "par": true,
		"pour": true, "sur": true, "avec": true, "sans": true,
	}

	minWordLength := 3
	for _, word := range words {
		if len(word) >= minWordLength && !stopWords[word] {
			significantWords = append(significantWords, word)
		}
	}

	return significantWords
}

// Max returns the maximum of two integers
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Min returns the minimum of two integers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Contains checks if a string is in a slice
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// FindURLsByPattern extracts URLs from text matching a specific regex pattern
func FindURLsByPattern(text string, pattern *regexp.Regexp) []string {
	return pattern.FindAllString(text, -1)
}

// IsHelloAssoURL checks if a URL is a HelloAsso URL
func IsHelloAssoURL(url string) bool {
	// Use the HelloAsso BaseURL constant to avoid hardcoded URLs
	baseURL := strings.TrimPrefix(helloasso.BaseURL, "https://")
	return strings.Contains(strings.ToLower(url), strings.ToLower(baseURL))
}
