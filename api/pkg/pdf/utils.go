package pdf

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Common constants
const (
	// DefaultFetchTimeout is the default timeout for HTTP requests
	DefaultFetchTimeout = 30 * time.Second
)

// PDFExtractor represents any text extraction function that processes PDF data
type PDFExtractor func([]byte) (string, time.Duration, error)

// PDFResult holds the result of a PDF text extraction operation
type PDFResult struct {
	Text          string
	Duration      time.Duration
	FetchDuration time.Duration
	TotalDuration time.Duration
	Error         error
}

// FetchPDFFromURL downloads PDF content from a URL and returns the content as bytes
func FetchPDFFromURL(url string) ([]byte, time.Duration, error) {
	startTime := time.Now()

	// Create client with timeout
	client := &http.Client{
		Timeout: DefaultFetchTimeout,
	}

	// Fetch the PDF from the URL
	resp, err := client.Get(url)
	if err != nil {
		return nil, time.Since(startTime), fmt.Errorf("error fetching PDF from URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, time.Since(startTime), fmt.Errorf("error fetching PDF: HTTP status %d", resp.StatusCode)
	}

	// Read the entire response body
	pdfContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, time.Since(startTime), fmt.Errorf("error reading PDF response: %w", err)
	}

	fetchDuration := time.Since(startTime)
	return pdfContent, fetchDuration, nil
}

// ReadPDFFile reads a PDF file and returns its content as bytes
func ReadPDFFile(pdfPath string) ([]byte, error) {
	// Check if file exists
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("PDF file not found: %s", pdfPath)
	}

	// Read file content into memory
	pdfContent, err := os.ReadFile(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("error reading PDF file content: %w", err)
	}

	return pdfContent, nil
}

// ProcessURLWithExtractor fetches a PDF from a URL and processes it with the provided extractor
func ProcessURLWithExtractor(url string, extractor PDFExtractor) PDFResult {
	startTime := time.Now()

	// Fetch the PDF from the URL
	pdfContent, fetchDuration, err := FetchPDFFromURL(url)
	if err != nil {
		return PDFResult{
			Error:         err,
			TotalDuration: time.Since(startTime),
			FetchDuration: fetchDuration,
		}
	}

	// Process the PDF bytes with the provided extractor
	text, processingDuration, err := extractor(pdfContent)
	if err != nil {
		return PDFResult{
			Error:         err,
			TotalDuration: time.Since(startTime),
			FetchDuration: fetchDuration,
		}
	}

	totalDuration := time.Since(startTime)

	// Log the timing information in a concise format
	fmt.Printf("PDF URL: fetch=%v, process=%v, total=%v\n",
		fetchDuration.Round(time.Millisecond),
		processingDuration.Round(time.Millisecond),
		totalDuration.Round(time.Millisecond))

	return PDFResult{
		Text:          text,
		Duration:      processingDuration,
		FetchDuration: fetchDuration,
		TotalDuration: totalDuration,
		Error:         nil,
	}
}
