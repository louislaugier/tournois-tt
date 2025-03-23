// Package extraction provides specialized PDF extraction functionality for tournament data
package extraction

import (
	"fmt"
	"log"
	"time"

	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/navigation"
	pdfutil "tournois-tt/api/pkg/pdf"
	"tournois-tt/api/pkg/utils"

	pw "github.com/playwright-community/playwright-go"
)

// ExtractSignupURLFromPDF processes a PDF file to extract potential signup URLs
// for a tournament and validates them against tournament information
func ExtractSignupURLFromPDF(tournament cache.TournamentCache, tournamentDate time.Time,
	pdfURL string, browserContext pw.BrowserContext) (string, error) {

	utils.DebugLog("Checking rules PDF for signup URL: %s", pdfURL)

	// Verify PDF file validity
	if !utils.IsPDFFile(pdfURL) {
		return "", fmt.Errorf("rules file is not a PDF: %s", pdfURL)
	}

	// Extract text from PDF
	utils.DebugLog("Extracting text from PDF: %s", pdfURL)

	pdfText, err := extractPDFTextWithRetries(pdfURL, 3, 5*time.Second)
	if err != nil {
		return "", err
	}

	// Look for URLs in the PDF text
	urls := utils.FindURLsInText(pdfText)
	if len(urls) == 0 {
		return "", fmt.Errorf("no URLs found in PDF")
	}

	// Filter out URLs from common domains that aren't useful
	var filteredURLs []string
	for _, url := range urls {
		if !utils.IsURLToSkip(url) {
			filteredURLs = append(filteredURLs, url)
		}
	}

	// Try to find a signup form on any of the URLs
	for _, url := range filteredURLs {
		signupURL, err := navigation.ValidateSignupURL(url, tournament, tournamentDate, browserContext)
		if err == nil && signupURL != "" {
			log.Printf("Found signup URL in PDF: %s", signupURL)
			return signupURL, nil
		}
	}

	// If direct validation failed, try recursive navigation on the URLs
	for _, url := range filteredURLs {
		signupURL, err := navigation.RecursivelyFindRegistrationForm(url, tournament, tournamentDate, browserContext)
		if err != nil {
			log.Printf("Error during recursive navigation: %v", err)
		} else if signupURL != "" {
			log.Printf("Found signup URL via recursive navigation: %s", signupURL)
			return signupURL, nil
		}
	}

	return "", fmt.Errorf("no signup URL found in PDF")
}

// extractPDFTextWithRetries attempts to extract text from a PDF with retry logic
func extractPDFTextWithRetries(pdfURL string, maxRetries int, retryDelay time.Duration) (string, error) {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Add delay between retries (except for first attempt)
		if attempt > 0 {
			time.Sleep(retryDelay)
			utils.DebugLog("Retrying PDF extraction (attempt %d/%d): %s", attempt+1, maxRetries, pdfURL)
		}

		// Extract text from PDF using the pkg/pdf utility
		text, _, err := pdfutil.FetchPDFAndExtractText(pdfURL)
		if err == nil {
			return text, nil
		}

		lastErr = err

		// Only retry if the error is one we can recover from
		if !utils.IsExtractionRetryableError(err.Error()) {
			break
		}
	}

	return "", fmt.Errorf("failed to extract text from PDF after %d attempts: %w", maxRetries, lastErr)
}
