// Package pdf_processor provides functionality for extracting and validating signup URLs from PDF files
package pdf_processor

import (
	"fmt"
	"log"
	"time"

	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/pdf"
	"tournois-tt/api/pkg/scraper/services/common"
	"tournois-tt/api/pkg/utils"

	pw "github.com/playwright-community/playwright-go"
)

// Constants for PDF processing
const (
	MaxURLsToProcess = 30 // Limit number of URLs to process to avoid excessive validation
	MaxRedirections  = 5  // Maximum number of recursively followed links to find signup form
)

// Config holds configuration for PDF processing
type Config struct {
	// Maximum retry attempts for PDF extraction errors
	MaxRetries int
	// Delay between retries
	RetryDelay time.Duration
	// Validator function to validate URLs
	Validator func(string, cache.TournamentCache, time.Time, pw.BrowserContext) (string, error)
}

// DefaultConfig returns a default configuration for PDF processing
func DefaultConfig() Config {
	return Config{
		MaxRetries: 3,
		RetryDelay: 5 * time.Second,
	}
}

// ExtractSignupURL extracts and validates signup URLs from PDF content
func ExtractSignupURL(tournament cache.TournamentCache, tournamentDate time.Time, rulesURL string, browserContext pw.BrowserContext, validator func(string, cache.TournamentCache, time.Time, pw.BrowserContext) (string, error)) (string, error) {
	config := DefaultConfig()
	config.Validator = validator
	return ExtractSignupURLWithConfig(tournament, tournamentDate, rulesURL, browserContext, config)
}

// ExtractSignupURLWithConfig extracts and validates signup URLs from PDF content with custom configuration
func ExtractSignupURLWithConfig(tournament cache.TournamentCache, tournamentDate time.Time, rulesURL string, browserContext pw.BrowserContext, config Config) (string, error) {
	utils.DebugLog("Checking rules PDF for signup URL: %s", rulesURL)

	// Verify PDF file validity
	if !common.IsPDFFile(rulesURL) {
		return "", fmt.Errorf("rules file is not a PDF: %s", rulesURL)
	}

	// Extract text from PDF using the pkg/pdf implementation with retries
	utils.DebugLog("Extracting text from PDF: %s", rulesURL)

	pdfText, err := extractPDFTextWithRetries(rulesURL, config.MaxRetries, config.RetryDelay)
	if err != nil {
		return "", err
	}

	// Create regex patterns for URL extraction
	patterns := pdf.NewRegexPatterns()

	// Process the PDF text to extract URLs
	return processExtractedPDFText(pdfText, patterns, tournament, tournamentDate, browserContext, config.Validator)
}

// extractPDFTextWithRetries extracts text from a PDF with retries for temporary errors
func extractPDFTextWithRetries(rulesURL string, maxRetries int, retryDelay time.Duration) (string, error) {
	var pdfText string
	var fetchDuration, processDuration time.Duration
	var attemptsMade int

	// First, try to read the text from a pre-extracted file (useful for testing)
	tempFilePath := "/tmp/debug_pdf.txt"
	if fileContent, err := pdf.ReadFileContent(tempFilePath); err == nil && fileContent != "" {
		utils.DebugLog("Using pre-extracted PDF text from %s (%d characters)", tempFilePath, len(fileContent))
		return fileContent, nil
	}

	// If no pre-extracted file exists, extract the text from the PDF
	for attemptsMade = 0; attemptsMade < maxRetries; attemptsMade++ {
		// Exponential backoff on retries
		if attemptsMade > 0 {
			currentRetryDelay := retryDelay * time.Duration(attemptsMade)
			log.Printf("PDF extraction error, retrying in %v (attempt %d/%d)",
				currentRetryDelay, attemptsMade+1, maxRetries)
			time.Sleep(currentRetryDelay)
		}

		result := pdf.ProcessURLWithExtractor(rulesURL, pdf.ExtractTextFromBytes)
		fetchDuration = result.FetchDuration
		processDuration = result.Duration

		// If no error or not a temporary error, break the retry loop
		if result.Error == nil {
			pdfText = result.Text
			break
		}

		// Check if this is a temporary error that can be retried
		if !common.IsExtractionRetryableError(result.Error.Error()) {
			return "", fmt.Errorf("failed to extract text from PDF: %w", result.Error)
		}
	}

	// If we still don't have any text after all retries
	if pdfText == "" {
		return "", fmt.Errorf("failed to extract text from PDF after %d attempts", attemptsMade+1)
	}

	utils.DebugLog("PDF processing took %v (fetch: %v, extraction: %v)",
		(fetchDuration + processDuration).Round(time.Millisecond),
		fetchDuration.Round(time.Millisecond),
		processDuration.Round(time.Millisecond))

	return pdfText, nil
}
