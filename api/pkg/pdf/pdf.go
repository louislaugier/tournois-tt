package pdf

import (
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"tournois-tt/api/pkg/utils"
)

// ExtractFileContentFromURL downloads a PDF and extracts its text
func ExtractFileContentFromURL(urlStr string) (string, error) {
	// Encode URL properly
	encodedURL, err := utils.EncodeURL(urlStr)
	if err != nil {
		return "", err
	}

	// Download PDF
	resp, err := http.Get(encodedURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		return "", err
	}
	defer resp.Body.Close()

	pdfBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Extract text using pdftotext
	if text, err := extractText(pdfBytes); err == nil && text != "" {
		return text, nil
	}

	return "Failed to extract text. PDF may be scanned or use non-standard fonts.", nil
}

// extractText extracts text from PDF using pdftotext
func extractText(pdfBytes []byte) (string, error) {
	// Check if pdftotext is available
	if _, err := exec.LookPath("pdftotext"); err != nil {
		return "", err
	}

	// Setup temp files
	tmpDir, err := os.MkdirTemp("", "pdf")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	pdfPath := filepath.Join(tmpDir, "in.pdf")
	txtPath := filepath.Join(tmpDir, "out.txt")

	// Process PDF
	if err := os.WriteFile(pdfPath, pdfBytes, 0644); err != nil {
		return "", err
	}

	if err := exec.Command("pdftotext", "-layout", pdfPath, txtPath).Run(); err != nil {
		return "", err
	}

	// Read results
	textBytes, err := os.ReadFile(txtPath)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(textBytes)), nil
}
