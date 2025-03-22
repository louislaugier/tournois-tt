package pdf

import (
	"fmt"
	"testing"
	"time"
)

// LogTextExtraction formats and logs the extracted text with information about the extraction
// It provides detailed logs for test output when the -v flag is used
func LogTextExtraction(t *testing.T, source string, text string, duration time.Duration, libraryInfo string) {
	// Detailed logs for test output when using -v flag
	t.Logf("----------------------------------------")
	t.Logf("PDF Text Extraction from %s%s", source, libraryInfo)
	t.Logf("Processing time: %v", duration)
	t.Logf("Text length: %d characters", len(text))
	t.Logf("----------------------------------------")

	// Simplified console output - only show library and time
	libraryName := "PDF"
	if libraryInfo != "" {
		libraryName = libraryInfo
	}
	fmt.Printf("PDF Extraction completed - Library: %s | Source: %s | Time: %v | Characters: %d\n",
		libraryName, source, duration, len(text))
}
