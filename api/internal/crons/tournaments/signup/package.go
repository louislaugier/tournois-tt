// Package signup provides functionality for finding and validating signup URLs for tournaments
package signup

import (
	"regexp"
	"strings"

	"tournois-tt/api/pkg/constants"
	"tournois-tt/api/pkg/helloasso"
	"tournois-tt/api/pkg/utils"
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
	// urlRegex uses the utils URLRegex
	urlRegex = utils.URLRegex

	// helloAssoURLRegex is a regex to find HelloAsso URLs
	helloAssoURLRegex = regexp.MustCompile(`https?://(?:www\.)?` +
		strings.TrimPrefix(strings.ReplaceAll(helloasso.BaseURL, ".", "\\."), "https://") +
		`/[^\s"']+`)

	// Use regex patterns from utils package
	tournoiSubdomainRegex = regexp.MustCompile(`https?://[^/]*tournoi[^/]*\.[^/]+/?`)
	paymentURLRegex       = regexp.MustCompile(`(?i)https?://[^\s"']+(?:pay|payment|paiement|checkout|caisse|transaction|order|commande)`)
	signupURLRegex        = regexp.MustCompile(`(?i)https?://[^\s"']+(?:sign|registration|inscription|register|inscrire|signup|form|formulaire)`)

	// RegistrationKeywords for finding registration-related content
	RegistrationKeywords = constants.RegistrationKeywords
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

// FindDomainOnlyReferences finds domain references in text
func FindDomainOnlyReferences(text string) []string {
	return utils.FindDomainOnlyReferences(text)
}

// GenerateCommonTournamentSubdomains generates common tournament subdomains
func GenerateCommonTournamentSubdomains(domain string) []string {
	return utils.GenerateCommonTournamentSubdomains(domain)
}

// FindURLsByPattern extracts URLs from text matching a specific regex pattern
func FindURLsByPattern(text string, pattern *regexp.Regexp) []string {
	return utils.FindURLsByPattern(text, pattern)
}
