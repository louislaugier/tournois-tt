package page

import (
	"fmt"

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
	// Navigate to the search page with timeout
	if _, err := h.page.Goto(url, playwright.PageGotoOptions{
		Timeout:   playwright.Float(60000), // Increase timeout to 60 seconds
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	}); err != nil {
		return fmt.Errorf("could not navigate to page: %v", err)
	}

	return nil
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
	emptyState, err := h.page.QuerySelector(emptyStateSelector)
	if err != nil {
		return false, fmt.Errorf("failed to check empty state: %v", err)
	}
	return emptyState != nil, nil
}

// GetPage returns the underlying Playwright page
func (h *Handler) GetPage() playwright.Page {
	return h.page
}
