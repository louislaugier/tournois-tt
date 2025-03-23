package pdf

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"seehuhn.de/go/pdf"
	"seehuhn.de/go/pdf/graphics/matrix"
	"seehuhn.de/go/pdf/pagetree"
	"seehuhn.de/go/pdf/reader"
)

// ExtractTextFromBytes2 reads PDF content from a byte slice and returns its text content
// using the seehuhn.de/go/pdf library.
// It also returns the time it took to process the content.
func ExtractTextFromBytes2(pdfContent []byte) (string, time.Duration, error) {
	startTime := time.Now()

	// Create a reader from the PDF content
	r := bytes.NewReader(pdfContent)

	// Create a new PDF reader
	pdfReader, err := pdf.NewReader(r, nil)
	if err != nil {
		return "", time.Since(startTime), fmt.Errorf("error creating PDF reader: %w", err)
	}

	// Collect the text content
	var textBuffer strings.Builder

	// Create a content reader that collects text
	contents := reader.New(pdfReader, nil)
	contents.Text = func(text string) error {
		textBuffer.WriteString(text)
		return nil
	}

	// Iterate through all pages
	pages := pagetree.NewIterator(pdfReader)
	pageNo := 0
	pages.All()(func(_ pdf.Reference, pageDict pdf.Dict) bool {
		// Add page separator
		if pageNo > 0 {
			textBuffer.WriteString("\n\n")
		}

		// Parse the page content
		err = contents.ParsePage(pageDict, matrix.Identity)
		if err != nil {
			// Don't fail the entire extraction for one bad page
			textBuffer.WriteString(fmt.Sprintf("[Error extracting text from page %d: %v]", pageNo+1, err))
		}

		pageNo++
		return true
	})

	processingTime := time.Since(startTime)
	return textBuffer.String(), processingTime, nil
}

// ExtractTextFromURL2 fetches a PDF from a URL and returns its text content
// using the seehuhn.de/go/pdf library.
// It also returns the time it took to process the content.
func ExtractTextFromURL2(url string) (string, time.Duration, error) {
	// Use the utility method to process the URL with our ExtractTextFromBytes2 function
	result := ProcessURLWithExtractor(url, ExtractTextFromBytes2)
	if result.Error != nil {
		// Calculate total duration by adding fetch and processing durations
		totalDuration := result.FetchDuration + result.Duration
		return "", totalDuration, result.Error
	}

	// Calculate total duration by adding fetch and processing durations
	totalDuration := result.FetchDuration + result.Duration
	return result.Text, totalDuration, nil
}

// ExtractText2 reads a PDF file and returns its text content
// using the seehuhn.de/go/pdf library.
// It also returns the time it took to process the file.
func ExtractText2(pdfPath string) (string, time.Duration, error) {
	startTime := time.Now()

	// Open the PDF file
	file, err := os.Open(pdfPath)
	if err != nil {
		return "", time.Since(startTime), fmt.Errorf("error opening PDF file: %w", err)
	}
	defer file.Close()

	// Create a new PDF reader directly from the file
	pdfReader, err := pdf.NewReader(file, nil)
	if err != nil {
		return "", time.Since(startTime), fmt.Errorf("error creating PDF reader: %w", err)
	}

	// Collect the text content
	var textBuffer strings.Builder

	// Create a content reader that collects text
	contents := reader.New(pdfReader, nil)
	contents.Text = func(text string) error {
		textBuffer.WriteString(text)
		return nil
	}

	// Iterate through all pages
	pages := pagetree.NewIterator(pdfReader)
	pageNo := 0
	pages.All()(func(_ pdf.Reference, pageDict pdf.Dict) bool {
		// Add page separator
		if pageNo > 0 {
			textBuffer.WriteString("\n\n")
		}

		// Parse the page content
		err = contents.ParsePage(pageDict, matrix.Identity)
		if err != nil {
			// Don't fail the entire extraction for one bad page
			textBuffer.WriteString(fmt.Sprintf("[Error extracting text from page %d: %v]", pageNo+1, err))
		}

		pageNo++
		return true
	})

	processingTime := time.Since(startTime)
	return textBuffer.String(), processingTime, nil
}
