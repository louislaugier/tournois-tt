package pdf

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Error returned when pdftotext binary is not available
var ErrPDFToTextNotInstalled = fmt.Errorf("pdftotext binary not found - install poppler-utils package")

// isPDFToTextAvailable checks if the pdftotext command is available
func isPDFToTextAvailable() bool {
	// First check if it's in PATH
	if _, err := exec.LookPath("pdftotext"); err == nil {
		return true
	}

	// Try to run it with -v option
	cmd := exec.Command("pdftotext", "-v")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Run the command
	err := cmd.Run()

	// Some versions of pdftotext output version info to stderr with exit code 0
	// Others output to stderr with a non-zero exit code
	if err == nil {
		return true
	}

	// If we have output in stderr and it mentions pdftotext, it's probably installed
	if stderr.Len() > 0 && strings.Contains(stderr.String(), "pdftotext") {
		return true
	}

	return false
}

// ExtractTextFromBytes3 reads PDF content from a byte slice and returns its text content
// using the pdftotext binary from Poppler utilities.
// It also returns the time it took to process the content.
func ExtractTextFromBytes3(pdfContent []byte) (string, time.Duration, error) {
	startTime := time.Now()

	// Check if pdftotext is available
	if !isPDFToTextAvailable() {
		return "", time.Since(startTime), ErrPDFToTextNotInstalled
	}

	// Create a temporary file to store the PDF content
	tmpFile, err := os.CreateTemp("", "pdf-*.pdf")
	if err != nil {
		return "", time.Since(startTime), fmt.Errorf("error creating temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name()) // Clean up

	// Write the PDF content to the temporary file
	if _, err := tmpFile.Write(pdfContent); err != nil {
		return "", time.Since(startTime), fmt.Errorf("error writing to temporary file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return "", time.Since(startTime), fmt.Errorf("error closing temporary file: %w", err)
	}

	// Create a temporary file for the text output
	txtFilePath := tmpFile.Name() + ".txt"
	defer os.Remove(txtFilePath) // Clean up

	// Run pdftotext command to extract text
	cmd := exec.Command("pdftotext", "-layout", "-enc", "UTF-8", tmpFile.Name(), txtFilePath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", time.Since(startTime), fmt.Errorf("error running pdftotext: %w, stderr: %s", err, stderr.String())
	}

	// Read the extracted text
	textContent, err := os.ReadFile(txtFilePath)
	if err != nil {
		return "", time.Since(startTime), fmt.Errorf("error reading text output: %w", err)
	}

	processingTime := time.Since(startTime)
	return string(textContent), processingTime, nil
}

// ExtractTextFromURL3 fetches a PDF from a URL and returns its text content
// using the pdftotext binary from Poppler utilities.
// It also returns the time it took to process the content.
func ExtractTextFromURL3(url string) (string, time.Duration, error) {
	// Use the utility method to process the URL with our ExtractTextFromBytes3 function
	result := ProcessURLWithExtractor(url, ExtractTextFromBytes3)
	if result.Error != nil {
		// Calculate total duration by adding fetch and processing durations
		totalDuration := result.FetchDuration + result.Duration
		return "", totalDuration, result.Error
	}

	// Calculate total duration by adding fetch and processing durations
	totalDuration := result.FetchDuration + result.Duration
	return result.Text, totalDuration, nil
}

// ExtractText3 reads a PDF file and returns its text content
// using the pdftotext binary from Poppler utilities.
// It also returns the time it took to process the file.
func ExtractText3(pdfPath string) (string, time.Duration, error) {
	startTime := time.Now()

	// Check if pdftotext is available
	if !isPDFToTextAvailable() {
		return "", time.Since(startTime), ErrPDFToTextNotInstalled
	}

	// Ensure the input file exists
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		return "", time.Since(startTime), fmt.Errorf("pdf file not found: %s", err)
	}

	// Get the absolute path
	absPath, err := filepath.Abs(pdfPath)
	if err != nil {
		return "", time.Since(startTime), fmt.Errorf("error getting absolute path: %w", err)
	}

	// Create a temporary file for the text output
	tmpDir := os.TempDir()
	txtFilePath := filepath.Join(tmpDir, "output-"+filepath.Base(pdfPath)+".txt")
	defer os.Remove(txtFilePath) // Clean up

	// Run pdftotext command to extract text
	cmd := exec.Command("pdftotext", "-layout", "-enc", "UTF-8", absPath, txtFilePath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", time.Since(startTime), fmt.Errorf("error running pdftotext: %w, stderr: %s", err, stderr.String())
	}

	// Read the extracted text
	textContent, err := os.ReadFile(txtFilePath)
	if err != nil {
		return "", time.Since(startTime), fmt.Errorf("error reading text output: %w", err)
	}

	processingTime := time.Since(startTime)
	return string(textContent), processingTime, nil
}
