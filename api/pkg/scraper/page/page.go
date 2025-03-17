package page

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/playwright-community/playwright-go"
)

// Handler manages page-specific operations
type Handler struct {
	page playwright.Page
}

// Config holds the configuration for page operations
type Config struct {
	SearchURL          string
	EmptyStateSelector string
	ResultsSelector    string
}

// New creates a new page handler
func New(page playwright.Page) *Handler {
	return &Handler{page: page}
}

// NavigateToPage navigates to the page with the given URL
func (h *Handler) NavigateToPage(url string) error {
	const maxRetries = 5
	const initialBackoffMs = 1000

	var lastError error

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Navigate to the search page with timeout
		if _, err := h.page.Goto(url, playwright.PageGotoOptions{
			Timeout:   playwright.Float(60000), // 60 seconds timeout
			WaitUntil: playwright.WaitUntilStateNetworkidle,
		}); err == nil {
			// Success - no error
			return nil
		} else {
			lastError = err
			backoffMs := initialBackoffMs * math.Pow(2, float64(attempt))
			log.Printf("Navigation to %s failed (attempt %d/%d): %v. Retrying in %.1f seconds...",
				url, attempt+1, maxRetries, err, backoffMs/1000)
			time.Sleep(time.Duration(backoffMs) * time.Millisecond)
		}
	}

	return fmt.Errorf("could not navigate to page after %d attempts: %v", maxRetries, lastError)
}

// WaitForResults waits for either search results or empty state to appear
func (h *Handler) WaitForResults(cfg Config) error {
	_, err := h.page.WaitForFunction(fmt.Sprintf(`
		() => {
			const results = document.querySelector('%s');
			const emptyState = document.querySelector('%s');
			return results !== null || emptyState !== null;
		}
	`, cfg.ResultsSelector, cfg.EmptyStateSelector), playwright.FrameWaitForFunctionOptions{
		Timeout: playwright.Float(30000),
	})
	if err != nil {
		return fmt.Errorf("timeout waiting for results: %v", err)
	}

	return nil
}

// HasEmptyState checks if the page shows an empty state
func (h *Handler) HasEmptyState(emptyStateSelector string) (bool, error) {
	emptyState := h.page.Locator(emptyStateSelector)
	count, err := emptyState.Count()
	if err != nil {
		return false, fmt.Errorf("failed to check empty state: %v", err)
	}
	return count > 0, nil
}

// GetPage returns the underlying Playwright page
func (h *Handler) GetPage() playwright.Page {
	return h.page
}
