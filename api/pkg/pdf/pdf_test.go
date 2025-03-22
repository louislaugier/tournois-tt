package pdf

import (
	"os"
	"testing"
)

const testPDFURL = "https://apiv2.fftt.com/api/files/412339/Tournoi%20National%20B%20de%20Maine%20C%C5%93ur%20de%20Sarthe%20TT%20VD.docx.pdf"

func TestExtractText(t *testing.T) {
	// This test will be skipped if no test PDF file is provided
	testPDFPath := "mock/test.pdf"

	// Skip if file doesn't exist - this is just an example
	if _, err := os.Stat(testPDFPath); os.IsNotExist(err) {
		t.Skip("Test PDF file not found. Skipping test.")
	}

	text, duration, err := ExtractText(testPDFPath)
	if err != nil {
		t.Fatalf("Failed to extract text from PDF: %v", err)
	}

	LogTextExtraction(t, "File", text, duration, "Original PDF (ledongthuc/pdf)")

	if len(text) == 0 {
		t.Error("Expected non-empty text from PDF extraction")
	}
}

func TestExtractTextFromBytes(t *testing.T) {
	// This test will be skipped if no test PDF file is provided
	testPDFPath := "mock/test.pdf"

	// Skip if file doesn't exist - this is just an example
	if _, err := os.Stat(testPDFPath); os.IsNotExist(err) {
		t.Skip("Test PDF file not found. Skipping test.")
	}

	// Read the file content into memory
	content, err := ReadPDFFile(testPDFPath)
	if err != nil {
		t.Skip("Failed to read test PDF file: ", err)
	}

	text, duration, err := ExtractTextFromBytes(content)
	if err != nil {
		t.Fatalf("Failed to extract text from PDF bytes: %v", err)
	}

	LogTextExtraction(t, "Bytes", text, duration, "Original PDF (ledongthuc/pdf)")

	if len(text) == 0 {
		t.Error("Expected non-empty text from PDF extraction")
	}
}

func TestExtractTextFromSpecificURL(t *testing.T) {
	// Using the specific URL for testing
	text, duration, err := ExtractTextFromURL(testPDFURL)
	if err != nil {
		t.Fatalf("Failed to extract text from the specific PDF URL: %v", err)
	}

	LogTextExtraction(t, "Specific URL", text, duration, "Original PDF (ledongthuc/pdf)")

	if len(text) == 0 {
		t.Error("Expected non-empty text from PDF extraction")
	}
}

func TestExtractTextFromURL(t *testing.T) {
	text, duration, err := ExtractTextFromURL(testPDFURL)
	if err != nil {
		t.Fatalf("Failed to extract text from PDF URL: %v", err)
	}

	LogTextExtraction(t, "URL from env", text, duration, "Original PDF (ledongthuc/pdf)")

	if len(text) == 0 {
		t.Error("Expected non-empty text from PDF extraction")
	}
}
