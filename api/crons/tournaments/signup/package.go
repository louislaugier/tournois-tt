// Package signup provides functionality for finding and validating signup URLs for tournaments
package signup

import (
	"regexp"
	"strings"

	"tournois-tt/api/pkg/scraper/services/common"
	"tournois-tt/api/pkg/scraper/services/helloasso"
)

// Debug enables verbose logging within the signup package
// Set to true to enable additional debug output
var Debug bool

// Constants for signup processing
const (
	numWorkers = 3 // Number of concurrent workers for signup URL refresh
)

// Create a regex pattern that uses the constant
var (
	// urlRegex uses the common URLRegex
	urlRegex = common.URLRegex

	// helloAssoURLRegex is a regex to find HelloAsso URLs
	helloAssoURLRegex = regexp.MustCompile(`https?://(?:www\.)?` +
		strings.TrimPrefix(strings.ReplaceAll(helloasso.BaseURL, ".", "\\."), "https://") +
		`/[^\s"']+`)

	// Use common regex patterns
	tournoiSubdomainRegex = common.TournoiSubdomainRegex
	paymentURLRegex       = common.PaymentURLRegex
	signupURLRegex        = common.SignupURLRegex

	// RegistrationKeywords is the exported version of registrationKeywords for tests
	RegistrationKeywords = common.RegistrationKeywords
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
	return common.FindDomainOnlyReferences(text)
}

// GenerateCommonTournamentSubdomains is the exported version of generateCommonTournamentSubdomains for tests
func GenerateCommonTournamentSubdomains(domain string) []string {
	return common.GenerateCommonTournamentSubdomains(domain)
}

// FindURLsByPattern extracts URLs from text matching a specific regex pattern
func FindURLsByPattern(text string, pattern *regexp.Regexp) []string {
	return common.FindURLsByPattern(text, pattern)
}
