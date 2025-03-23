package pdf

import (
	"regexp"
	"tournois-tt/api/pkg/scraper/services/common"
)

// RegexPatterns for finding URLs in PDF content
type RegexPatterns struct {
	URLRegex              *regexp.Regexp
	TournoiSubdomainRegex *regexp.Regexp
	PaymentURLRegex       *regexp.Regexp
	SignupURLRegex        *regexp.Regexp
}

// NewRegexPatterns creates a new set of regex patterns for PDF URL extraction
func NewRegexPatterns() *RegexPatterns {
	return &RegexPatterns{
		URLRegex:              common.URLRegex,
		TournoiSubdomainRegex: common.TournoiSubdomainRegex,
		PaymentURLRegex:       common.PaymentURLRegex,
		SignupURLRegex:        common.SignupURLRegex,
	}
}

// CheckIsPDFFile checks if a URL points to a PDF file by examining the file extension and URL patterns
// This is a wrapper around common.IsPDFFile for backward compatibility
func CheckIsPDFFile(urlStr string) bool {
	return common.IsPDFFile(urlStr)
}

// IsExtractionRetryableError determines if a PDF extraction error can be retried
// This is a wrapper around common.IsExtractionRetryableError for backward compatibility
func IsExtractionRetryableError(errStr string) bool {
	return common.IsExtractionRetryableError(errStr)
}

// FindDomainOnlyReferences finds domain references for tournament websites mentioned in text
// This is a wrapper around common.FindDomainOnlyReferences for backward compatibility
func FindDomainOnlyReferences(text string) []string {
	return common.FindDomainOnlyReferences(text)
}

// EnsureURLProtocol ensures that a URL has a protocol, defaulting to https
// This is a wrapper around common.EnsureURLProtocol for backward compatibility
func EnsureURLProtocol(url string) string {
	return common.EnsureURLProtocol(url)
}

// GenerateCommonTournamentSubdomains generates common tournament subdomains from a base domain
// This is a wrapper around common.GenerateCommonTournamentSubdomains for backward compatibility
func GenerateCommonTournamentSubdomains(domain string) []string {
	return common.GenerateCommonTournamentSubdomains(domain)
}

// LimitURLs limits the number of URLs to a maximum count
// This is a wrapper around common.LimitURLs for backward compatibility
func LimitURLs(urls []string, maxCount int) []string {
	return common.LimitURLs(urls, maxCount)
}

// This file provides backward compatibility for code that uses these functions.
// For new code, use the common package directly.
