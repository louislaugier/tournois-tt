// Package page provides utilities for handling Playwright pages
package page

import (
	"fmt"
	"time"
	"tournois-tt/api/pkg/utils"

	pw "github.com/playwright-community/playwright-go"
)

// Config holds configuration for page handlers
type Config struct {
	// Selectors
	EmptyStateSelector string
	ResultsSelector    string

	// Timeouts
	NavigationTimeout time.Duration
	OperationTimeout  time.Duration
}

// PageHandler wraps a playwright page with additional functionality
type PageHandler struct {
	Page pw.Page
}

// New creates a new page handler (alias for NewPageHandler)
func New(page pw.Page) *PageHandler {
	return NewPageHandler(page)
}

// NewPageHandler creates a new page handler
func NewPageHandler(page pw.Page) *PageHandler {
	return &PageHandler{
		Page: page,
	}
}

// Close safely closes the page
func (h *PageHandler) Close() {
	if h.Page != nil {
		h.Page.Close()
	}
}

// GetContent gets the page content
func (h *PageHandler) GetContent() (string, error) {
	content, err := h.Page.Content()
	if err != nil {
		return "", err
	}
	return content, nil
}

// SetDefaultTimeouts sets default timeouts for navigation and operations
func (h *PageHandler) SetDefaultTimeouts(navigationTimeout, operationTimeout time.Duration) {
	h.Page.SetDefaultNavigationTimeout(float64(navigationTimeout / time.Millisecond))
	h.Page.SetDefaultTimeout(float64(operationTimeout / time.Millisecond))
}

// TakeScreenshot takes a screenshot of the page
func (h *PageHandler) TakeScreenshot(path string) error {
	_, err := h.Page.Screenshot(pw.PageScreenshotOptions{
		Path:     pw.String(path),
		FullPage: pw.Bool(true),
	})
	return err
}

// HasEmptyState checks if the page has an empty state element
func (h *PageHandler) HasEmptyState(selector string) (bool, error) {
	locator := h.Page.Locator(selector)
	count, err := locator.Count()
	return count > 0, err
}

// WaitForResults waits for results to load
func (h *PageHandler) WaitForResults(config Config) error {
	if config.ResultsSelector == "" {
		return fmt.Errorf("results selector is empty")
	}

	_, err := h.Page.WaitForSelector(config.ResultsSelector, pw.PageWaitForSelectorOptions{
		State:   pw.WaitForSelectorStateAttached,
		Timeout: pw.Float(30000),
	})
	return err
}

// SafeOperation executes an operation with error handling
func (h *PageHandler) SafeOperation(name string, operation func() error) error {
	utils.DebugLog("Starting operation: %s", name)
	err := operation()
	if err != nil {
		return fmt.Errorf("operation %s failed: %w", name, err)
	}
	utils.DebugLog("Completed operation: %s", name)
	return nil
}

// SafeNavigation navigates to a URL with improved error handling
func (h *PageHandler) SafeNavigation(url string, maxRetries int, errorCallback func(error) error) error {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			utils.DebugLog("Retrying navigation to %s (attempt %d/%d)", url, i+1, maxRetries)
			time.Sleep(time.Duration(i) * time.Second)
		}

		_, err := h.Page.Goto(url, pw.PageGotoOptions{
			WaitUntil: pw.WaitUntilStateNetworkidle,
			Timeout:   pw.Float(30000),
		})

		if err == nil {
			return nil
		}

		lastErr = err

		if errorCallback != nil {
			if cbErr := errorCallback(err); cbErr != nil {
				return cbErr
			}
		}
	}

	return fmt.Errorf("failed to navigate to %s after %d attempts: %w", url, maxRetries, lastErr)
}

// GetText gets text content from the page
func (h *PageHandler) GetText(selector string) (string, error) {
	element, err := h.Page.QuerySelector(selector)
	if err != nil {
		return "", err
	}

	if element == nil {
		return "", nil
	}

	return element.TextContent()
}

// WaitAndClick waits for an element and clicks it
func (h *PageHandler) WaitAndClick(selector string, timeout float64) error {
	// Wait for the selector to be visible
	_, err := h.Page.WaitForSelector(selector, pw.PageWaitForSelectorOptions{
		Timeout: pw.Float(timeout),
	})

	if err != nil {
		return err
	}

	// Click the element
	return h.Page.Click(selector)
}

// EvaluateBool evaluates a JavaScript expression and returns a boolean result
func (h *PageHandler) EvaluateBool(expression string) (bool, error) {
	result, err := h.Page.Evaluate(expression)
	if err != nil {
		return false, err
	}

	boolResult, ok := result.(bool)
	if !ok {
		return false, fmt.Errorf("expected boolean result, got %T", result)
	}

	return boolResult, nil
}
