package pdf

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"tournois-tt/api/pkg/utils"
)

// PDF processing constants
const (
	MaxURLsToProcess = 30 // Limit number of URLs to process to avoid excessive validation
)

// IsPDFFile checks if a URL points to a PDF file
func IsPDFFile(url string) bool {
	return utils.IsPDFFile(url)
}

// ReadFileContent reads content from a file
func ReadFileContent(filePath string) (string, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// FindURLsInText finds all URLs in text
func FindURLsInText(text string) []string {
	return utils.FindURLsInText(text)
}

// ProcessWithTimings processes a PDF URL and returns the extraction result with timing information
func ProcessWithTimings(url string) (string, time.Duration, time.Duration, error) {
	startTime := time.Now()

	// Fetch PDF content
	pdfBytes, err := fetchPDFContent(url)
	if err != nil {
		return "", 0, 0, fmt.Errorf("error fetching PDF content: %w", err)
	}
	fetchDuration := time.Since(startTime)

	// Extract text from PDF
	processingStartTime := time.Now()
	extractedText, err := processBytes(pdfBytes)
	processDuration := time.Since(processingStartTime)

	if err != nil {
		return "", fetchDuration, processDuration, fmt.Errorf("error extracting text from PDF: %w", err)
	}

	totalTime := time.Since(startTime)
	log.Printf("PDF processed in %v (fetch: %v, extract: %v)",
		totalTime.Round(time.Millisecond),
		fetchDuration.Round(time.Millisecond),
		processDuration.Round(time.Millisecond))

	return extractedText, fetchDuration, processDuration, nil
}

// fetchPDFContent fetches PDF content from a URL
func fetchPDFContent(url string) ([]byte, error) {
	// Implementation goes here
	return nil, fmt.Errorf("not implemented")
}

// processBytes processes PDF bytes to extract text
func processBytes(pdfBytes []byte) (string, error) {
	// Implementation goes here
	return "", fmt.Errorf("not implemented")
}

// ReadBytesAndExtractText reads PDF content from a byte slice
func ReadBytesAndExtractText(pdfBytes []byte) (string, time.Duration, error) {
	startTime := time.Now()

	// Check if this is a valid PDF file
	if !utils.IsPDFFile(string(pdfBytes[:min(len(pdfBytes), 1024)])) {
		return "", time.Since(startTime), fmt.Errorf("not a valid PDF file")
	}

	// Create a temporary file to store the PDF
	tempFile, err := os.CreateTemp("", "pdf-*.pdf")
	if err != nil {
		return "", time.Since(startTime), fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	if _, err := tempFile.Write(pdfBytes); err != nil {
		return "", time.Since(startTime), fmt.Errorf("failed to write to temporary file: %w", err)
	}

	// Read text from the temporary file
	text, err := ReadFileAndExtractText(tempFile.Name())
	if err != nil {
		text = extractURLsFromRawPDF(string(pdfBytes))
		urls := utils.FindURLsByPattern(text, utils.URLRegex)
		if len(urls) > 0 {
			text = strings.Join(urls, "\n")
		} else {
			urls := utils.FindURLsInText(string(pdfBytes))
			if len(urls) > 0 {
				text = urls[0]
			}
		}
	}

	return text, time.Since(startTime), nil
}

// FetchPDFAndExtractText fetches a PDF from a URL and extracts text
func FetchPDFAndExtractText(url string) (string, time.Duration, error) {
	startTime := time.Now()

	// Fetch the PDF
	resp, err := http.Get(url)
	if err != nil {
		return "", time.Since(startTime), fmt.Errorf("failed to download PDF: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", time.Since(startTime), fmt.Errorf("failed to download PDF: %s", resp.Status)
	}

	// Read the response body
	pdfBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", time.Since(startTime), fmt.Errorf("failed to read PDF: %w", err)
	}

	// Extract text from the PDF bytes
	text, _, err := ReadBytesAndExtractText(pdfBytes)
	return text, time.Since(startTime), err
}

// ReadFileAndExtractText reads a PDF file from a specified path and extracts text
func ReadFileAndExtractText(path string) (string, error) {
	// Read the entire file
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF file: %w", err)
	}
	defer file.Close()

	pdfBytes, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read PDF file: %w", err)
	}

	// Create a temporary implementation using raw PDF text
	text := extractURLsFromRawPDF(string(pdfBytes))

	// If nothing found, log an error but don't fail
	if text == "" {
		log.Printf("Warning: Could not extract text from PDF file, this needs a proper PDF extraction library")
	}

	return text, nil
}

// extractURLsFromRawPDF tries to extract URLs from the raw PDF content as a fallback
func extractURLsFromRawPDF(pdfContent string) string {
	// Simple heuristic approach to find HTTP/HTTPS URLs in the raw PDF content
	urls := utils.FindURLsInText(pdfContent)
	return strings.Join(urls, "\n")
}

// min returns the minimum of a and b
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
