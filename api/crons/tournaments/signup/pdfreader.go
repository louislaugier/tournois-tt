package signup

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/pdf"

	pw "github.com/playwright-community/playwright-go"
)

// PDF processing constants
const (
	maxURLsToProcess = 30 // Limit number of URLs to process to avoid excessive validation
)

// ExtractSignupURLFromPDFFile extracts and validates signup URLs from PDF content
func ExtractSignupURLFromPDFFile(tournament cache.TournamentCache, tournamentDate time.Time, rulesURL string, browserContext pw.BrowserContext) (string, error) {
	debugLog("Checking rules PDF for signup URL: %s", rulesURL)

	// Verify PDF file validity
	if !isPDFFile(rulesURL) {
		return "", fmt.Errorf("rules file is not a PDF: %s", rulesURL)
	}

	// Extract text from PDF using the pkg/pdf implementation
	debugLog("Extracting text from PDF: %s", rulesURL)
	result := pdf.ProcessURLWithExtractor(rulesURL, pdf.ExtractTextFromBytes)
	if result.Error != nil {
		return "", fmt.Errorf("failed to extract text from PDF: %w", result.Error)
	}

	debugLog("PDF processing took %v (fetch: %v, extraction: %v)",
		result.TotalDuration.Round(time.Millisecond),
		result.FetchDuration.Round(time.Millisecond),
		result.Duration.Round(time.Millisecond))

	// Process the PDF text to extract URLs using the URL extraction logic
	return processExtractedPDFText(result.Text, tournament, tournamentDate, browserContext)
}

// isPDFFile checks if a URL points to a PDF file based on extension
func isPDFFile(urlStr string) bool {
	ext := strings.ToLower(filepath.Ext(urlStr))
	return ext == ".pdf"
}

// processExtractedPDFText processes extracted PDF text to find and validate signup URLs
func processExtractedPDFText(pdfText string, tournament cache.TournamentCache, tournamentDate time.Time, browserContext pw.BrowserContext) (string, error) {
	// Check for HelloAsso URLs first
	helloAssoURLs := FindURLsByPattern(pdfText, helloAssoURLRegex)
	if len(helloAssoURLs) > 0 {
		debugLog("Found %d HelloAsso URLs in PDF", len(helloAssoURLs))

		// Limit the number of URLs to validate
		urlsToValidate := limitURLs(helloAssoURLs, maxURLsToProcess)

		// Try to validate the HelloAsso URLs
		validURL, found := tryValidateURLs(urlsToValidate, tournament, tournamentDate, browserContext)
		if found {
			return validURL, nil
		}
	}

	// If no HelloAsso URLs, look for registration-related URLs
	debugLog("Looking for registration-related URLs in PDF")
	registrationURLs := findRegistrationURLsInPDF(pdfText)

	if len(registrationURLs) > 0 {
		debugLog("Found %d potential registration URLs in PDF", len(registrationURLs))

		// Limit the number of URLs to validate
		urlsToValidate := limitURLs(registrationURLs, maxURLsToProcess)

		// Try to validate the registration URLs
		validURL, found := tryValidateURLs(urlsToValidate, tournament, tournamentDate, browserContext)
		if found {
			return validURL, nil
		}
	}

	// No valid signup URL found
	debugLog("No valid signup URL found in PDF")
	return "", nil
}

// limitURLs limits the number of URLs to process
func limitURLs(urls []string, maxURLs int) []string {
	if len(urls) <= maxURLs {
		return urls
	}
	debugLog("Limiting URL validation to %d out of %d URLs", maxURLs, len(urls))
	return urls[:maxURLs]
}

// tryValidateURLs tries to validate a list of URLs and returns the first valid one
func tryValidateURLs(urls []string, tournament cache.TournamentCache, tournamentDate time.Time, browserContext pw.BrowserContext) (string, bool) {
	for _, url := range urls {
		debugLog("Validating URL from PDF: %s", url)

		// Validate the URL using the validator function from validators.go
		validURL, err := ValidateSignupURL(url, tournament, tournamentDate, browserContext)
		if err != nil {
			log.Printf("Warning: Failed to validate URL from PDF: %v", err)
			continue
		}

		if validURL != "" {
			log.Printf("Found valid signup URL in PDF: %s", validURL)
			return validURL, true
		}
	}

	return "", false
}

// findRegistrationURLsInPDF finds URLs in text that might be related to registration
func findRegistrationURLsInPDF(text string) []string {
	// Find all URLs in text using the shared utility function
	allURLs := urlRegex.FindAllString(text, -1)
	var registrationURLs []string

	// Look for URLs near registration keywords
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		line = strings.ToLower(line)

		// Check if line contains registration keywords
		containsKeyword := false
		for _, keyword := range registrationKeywords {
			if strings.Contains(line, keyword) {
				containsKeyword = true
				break
			}
		}

		if containsKeyword {
			// Check current line and surrounding lines for URLs
			startIdx := Max(0, i-2)
			endIdx := Min(len(lines)-1, i+2)

			for j := startIdx; j <= endIdx; j++ {
				urlsInLine := urlRegex.FindAllString(lines[j], -1)
				for _, url := range urlsInLine {
					// Only add unique URLs
					if !Contains(registrationURLs, url) {
						registrationURLs = append(registrationURLs, url)
					}
				}
			}
		}
	}

	// Also include URLs from domains commonly used for registration
	registrationDomains := []string{
		"inscription", "register", "signup", "helloasso", "billetweb", "weezevent",
		"eventbrite", "form", "formulaire",
	}

	for _, url := range allURLs {
		urlLower := strings.ToLower(url)
		for _, domain := range registrationDomains {
			if strings.Contains(urlLower, domain) && !Contains(registrationURLs, url) {
				registrationURLs = append(registrationURLs, url)
				break
			}
		}
	}

	return registrationURLs
}
