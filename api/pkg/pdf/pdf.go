package pdf

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/ledongthuc/pdf"
)

// ExtractTextFromBytes reads PDF content from a byte slice and returns its text content.
// It also returns the time it took to process the content.
// This is useful when you have the entire PDF in memory.
func ExtractTextFromBytes(pdfContent []byte) (string, time.Duration, error) {
	startTime := time.Now()

	// Create a bytes.Reader which implements io.ReaderAt
	reader := bytes.NewReader(pdfContent)

	// Call the PDF reader with the correct interface type
	pdfReader, err := pdf.NewReader(reader, int64(len(pdfContent)))
	if err != nil {
		return "", time.Since(startTime), fmt.Errorf("error creating PDF reader: %w", err)
	}

	var buf bytes.Buffer
	b, err := pdfReader.GetPlainText()
	if err != nil {
		return "", time.Since(startTime), fmt.Errorf("error extracting text from PDF: %w", err)
	}

	_, err = io.Copy(&buf, b)
	if err != nil {
		return "", time.Since(startTime), fmt.Errorf("error reading PDF text: %w", err)
	}

	processingTime := time.Since(startTime)
	return buf.String(), processingTime, nil
}

// ExtractTextFromURL fetches a PDF from a URL and returns its text content.
// It also returns the time it took to process the content.
func ExtractTextFromURL(url string) (string, time.Duration, error) {
	// Use the utility method to process the URL with our ExtractTextFromBytes function
	result := ProcessURLWithExtractor(url, ExtractTextFromBytes)
	if result.Error != nil {
		return "", result.TotalDuration, result.Error
	}

	return result.Text, result.TotalDuration, nil
}

// ExtractText reads a PDF file and returns its text content.
// It also returns the time it took to process the file.
// This is kept for backward compatibility and local file access.
func ExtractText(pdfPath string) (string, time.Duration, error) {
	startTime := time.Now()

	// Use the utility method to read the PDF file
	pdfContent, err := ReadPDFFile(pdfPath)
	if err != nil {
		return "", time.Since(startTime), err
	}

	// Process the PDF bytes
	return ExtractTextFromBytes(pdfContent)
}
