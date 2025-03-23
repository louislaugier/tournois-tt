package page

import (
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	pw "github.com/playwright-community/playwright-go"
)

// -----------------------------------------------------------------------------
// Types and Constants
// -----------------------------------------------------------------------------

// Default timeout values
const (
	DefaultNavigationTimeout = 60 * time.Second
	DefaultWaitTimeout       = 30 * time.Second
	DefaultRetryAttempts     = 5
	DefaultBackoffMs         = 1000
)

// Handler manages page-specific operations
type Handler struct {
	page pw.Page
}

// Config holds the configuration for page operations
type Config struct {
	// SearchURL is the base URL for search operations
	SearchURL string
	// EmptyStateSelector is the CSS selector for empty state indicators
	EmptyStateSelector string
	// ResultsSelector is the CSS selector for search results
	ResultsSelector string
	// NavigationTimeout is the timeout for page navigation operations
	NavigationTimeout time.Duration
	// WaitTimeout is the timeout for page waits
	WaitTimeout time.Duration
	// RetryAttempts is the number of retry attempts for operations
	RetryAttempts int
}

// DefaultConfig returns default configuration values
func DefaultConfig() Config {
	return Config{
		NavigationTimeout: DefaultNavigationTimeout,
		WaitTimeout:       DefaultWaitTimeout,
		RetryAttempts:     DefaultRetryAttempts,
	}
}

// -----------------------------------------------------------------------------
// Constructor and Basic Operations
// -----------------------------------------------------------------------------

// New creates a new page handler
func New(page pw.Page) *Handler {
	if page == nil {
		log.Println("Warning: Creating Handler with nil page")
	}
	return &Handler{page: page}
}

// GetPage returns the underlying Playwright page
func (h *Handler) GetPage() pw.Page {
	return h.page
}

// Close safely closes the page
func (h *Handler) Close() error {
	if h.page == nil {
		return nil
	}

	if err := h.page.Close(); err != nil {
		return fmt.Errorf("failed to close page: %w", err)
	}
	return nil
}

// SetDefaultTimeouts sets the default timeouts for the page
func (h *Handler) SetDefaultTimeouts(navigationTimeout, operationTimeout time.Duration) error {
	if h.page == nil {
		return fmt.Errorf("page is nil")
	}

	if navigationTimeout > 0 {
		h.page.SetDefaultNavigationTimeout(float64(navigationTimeout / time.Millisecond))
	}

	if operationTimeout > 0 {
		h.page.SetDefaultTimeout(float64(operationTimeout / time.Millisecond))
	}

	return nil
}

// -----------------------------------------------------------------------------
// Navigation and Waiting
// -----------------------------------------------------------------------------

// NavigateToPage navigates to the page with the given URL
func (h *Handler) NavigateToPage(url string) error {
	if h.page == nil {
		return fmt.Errorf("page is nil")
	}

	const maxRetries = DefaultRetryAttempts
	const initialBackoffMs = DefaultBackoffMs

	var lastError error

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Navigate to the search page with timeout
		if _, err := h.page.Goto(url, pw.PageGotoOptions{
			Timeout:   pw.Float(float64(DefaultNavigationTimeout / time.Millisecond)),
			WaitUntil: pw.WaitUntilStateNetworkidle,
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

	return fmt.Errorf("could not navigate to page after %d attempts: %w", maxRetries, lastError)
}

// NavigateWithConfig navigates to a URL with configurable retry and timeout settings
func (h *Handler) NavigateWithConfig(url string, cfg Config) error {
	if h.page == nil {
		return fmt.Errorf("page is nil")
	}

	maxRetries := DefaultRetryAttempts
	if cfg.RetryAttempts > 0 {
		maxRetries = cfg.RetryAttempts
	}

	timeout := DefaultNavigationTimeout
	if cfg.NavigationTimeout > 0 {
		timeout = cfg.NavigationTimeout
	}

	var lastError error

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Navigate to the page with timeout
		if _, err := h.page.Goto(url, pw.PageGotoOptions{
			Timeout:   pw.Float(float64(timeout / time.Millisecond)),
			WaitUntil: pw.WaitUntilStateNetworkidle,
		}); err == nil {
			// Success - no error
			return nil
		} else {
			lastError = err
			backoffMs := DefaultBackoffMs * math.Pow(2, float64(attempt))
			log.Printf("Navigation to %s failed (attempt %d/%d): %v. Retrying in %.1f seconds...",
				url, attempt+1, maxRetries, err, backoffMs/1000)
			time.Sleep(time.Duration(backoffMs) * time.Millisecond)
		}
	}

	return fmt.Errorf("could not navigate to page after %d attempts: %w", maxRetries, lastError)
}

// WaitForResults waits for either search results or empty state to appear
func (h *Handler) WaitForResults(cfg Config) error {
	if h.page == nil {
		return fmt.Errorf("page is nil")
	}

	if cfg.ResultsSelector == "" || cfg.EmptyStateSelector == "" {
		return fmt.Errorf("missing required selectors in config")
	}

	timeout := DefaultWaitTimeout
	if cfg.WaitTimeout > 0 {
		timeout = cfg.WaitTimeout
	}

	_, err := h.page.WaitForFunction(fmt.Sprintf(`
		() => {
			const results = document.querySelector('%s');
			const emptyState = document.querySelector('%s');
			return results !== null || emptyState !== null;
		}
	`, cfg.ResultsSelector, cfg.EmptyStateSelector), pw.FrameWaitForFunctionOptions{
		Timeout: pw.Float(float64(timeout / time.Millisecond)),
	})
	if err != nil {
		return fmt.Errorf("timeout waiting for results: %w", err)
	}

	return nil
}

// WaitForSelector waits for a selector to be visible with a timeout
func (h *Handler) WaitForSelector(selector string, timeout time.Duration) error {
	if h.page == nil {
		return fmt.Errorf("page is nil")
	}

	if timeout <= 0 {
		timeout = DefaultWaitTimeout
	}

	_, err := h.page.WaitForSelector(selector, pw.PageWaitForSelectorOptions{
		State:   pw.WaitForSelectorStateVisible,
		Timeout: pw.Float(float64(timeout / time.Millisecond)),
	})
	if err != nil {
		return fmt.Errorf("timeout waiting for selector %s: %w", selector, err)
	}

	return nil
}

// -----------------------------------------------------------------------------
// Page State and Content
// -----------------------------------------------------------------------------

// HasEmptyState checks if the page shows an empty state
func (h *Handler) HasEmptyState(emptyStateSelector string) (bool, error) {
	if h.page == nil {
		return false, fmt.Errorf("page is nil")
	}

	emptyState := h.page.Locator(emptyStateSelector)
	count, err := emptyState.Count()
	if err != nil {
		return false, fmt.Errorf("failed to check empty state: %w", err)
	}
	return count > 0, nil
}

// GetContent returns the HTML content of the page
func (h *Handler) GetContent() (string, error) {
	if h.page == nil {
		return "", fmt.Errorf("page is nil")
	}

	content, err := h.page.Content()
	if err != nil {
		return "", fmt.Errorf("failed to get page content: %w", err)
	}
	return content, nil
}

// GetTextContent returns the text content of a specified selector
func (h *Handler) GetTextContent(selector string) (string, error) {
	if h.page == nil {
		return "", fmt.Errorf("page is nil")
	}

	element := h.page.Locator(selector)
	if element == nil {
		return "", fmt.Errorf("selector not found: %s", selector)
	}

	isVisible, err := element.IsVisible()
	if err != nil {
		return "", fmt.Errorf("failed to check visibility for selector %s: %w", selector, err)
	}

	if !isVisible {
		return "", fmt.Errorf("selector is not visible: %s", selector)
	}

	text, err := element.TextContent()
	if err != nil {
		return "", fmt.Errorf("failed to get text content for selector %s: %w", selector, err)
	}

	return strings.TrimSpace(text), nil
}

// -----------------------------------------------------------------------------
// Interaction Methods
// -----------------------------------------------------------------------------

// Fill enters text into a form field
func (h *Handler) Fill(selector, text string) error {
	if h.page == nil {
		return fmt.Errorf("page is nil")
	}

	locator := h.page.Locator(selector)
	if err := locator.Fill(text); err != nil {
		return fmt.Errorf("failed to fill text in selector %s: %w", selector, err)
	}
	return nil
}

// Click clicks on an element specified by the selector
func (h *Handler) Click(selector string) error {
	if h.page == nil {
		return fmt.Errorf("page is nil")
	}

	locator := h.page.Locator(selector)
	if err := locator.Click(); err != nil {
		return fmt.Errorf("failed to click selector %s: %w", selector, err)
	}
	return nil
}

// ClickAndWait clicks an element and waits for either a network idle event or a specific selector to appear
func (h *Handler) ClickAndWait(clickSelector, waitSelector string, timeout time.Duration) error {
	if h.page == nil {
		return fmt.Errorf("page is nil")
	}

	if timeout <= 0 {
		timeout = DefaultWaitTimeout
	}

	// Click the element
	if err := h.Click(clickSelector); err != nil {
		return fmt.Errorf("failed to click during ClickAndWait: %w", err)
	}

	// If a wait selector is provided, wait for that element
	if waitSelector != "" {
		return h.WaitForSelector(waitSelector, timeout)
	}

	// Otherwise, wait a bit to let any navigation complete
	time.Sleep(1 * time.Second)

	return nil
}

// -----------------------------------------------------------------------------
// Screenshot and Debugging
// -----------------------------------------------------------------------------

// TakeScreenshot takes a screenshot of the current page state
func (h *Handler) TakeScreenshot(path string) error {
	if h.page == nil {
		return fmt.Errorf("page is nil")
	}

	if _, err := h.page.Screenshot(pw.PageScreenshotOptions{
		Path:     pw.String(path),
		FullPage: pw.Bool(true),
	}); err != nil {
		return fmt.Errorf("failed to take screenshot: %w", err)
	}

	log.Printf("Screenshot saved to: %s", path)
	return nil
}

// LogConsole enables console logging from the browser
func (h *Handler) LogConsole() {
	if h.page == nil {
		log.Println("Cannot enable console logging on nil page")
		return
	}

	h.page.On("console", func(msg pw.ConsoleMessage) {
		log.Printf("Browser console [%s]: %s", msg.Type(), msg.Text())
	})

	log.Println("Browser console logging enabled")
}

// -----------------------------------------------------------------------------
// Integration with Browser Package
// -----------------------------------------------------------------------------

// CheckPageHealth verifies if the page is healthy and can be used for operations
func (h *Handler) CheckPageHealth() bool {
	if h.page == nil {
		return false
	}

	// Try a simple operation that should always work on a healthy page
	_, err := h.page.Title()
	if err != nil {
		log.Printf("Page health check failed: %v", err)
		return false
	}

	return true
}

// SafeNavigation attempts to navigate to a URL with automatic recovery if the page is unhealthy
// It works with the browser.RestartIfUnhealthy mechanism
func (h *Handler) SafeNavigation(url string, retries int, browserHealthCheck func() (bool, error)) error {
	if h.page == nil {
		return fmt.Errorf("page is nil")
	}

	// First try normal navigation
	err := h.NavigateToPage(url)
	if err == nil {
		// Navigation succeeded
		return nil
	}

	// If navigation failed and we have a health check function, try recovery
	if browserHealthCheck != nil {
		log.Printf("Navigation failed, checking browser health: %v", err)

		// Check if browser needs recovery
		recovered, recoveryErr := browserHealthCheck()
		if recoveryErr != nil {
			return fmt.Errorf("browser health check failed: %w", recoveryErr)
		}

		if recovered {
			log.Println("Browser was recovered, page is now invalid")
			return fmt.Errorf("page is no longer valid after browser recovery, create a new page")
		}
	}

	// If browser is healthy but navigation still failed, retry with backoff
	if retries > 0 {
		log.Printf("Retrying navigation to %s with %d attempts", url, retries)
		return h.NavigateWithConfig(url, Config{
			RetryAttempts: retries,
		})
	}

	return fmt.Errorf("navigation failed: %w", err)
}

// SafeOperation executes an operation on the page with health checking
// If the operation fails due to page or browser issues, it will return an appropriate error
func (h *Handler) SafeOperation(operation string, action func() error) error {
	if h.page == nil {
		return fmt.Errorf("page is nil")
	}

	// First check if the page is healthy
	if !h.CheckPageHealth() {
		return fmt.Errorf("cannot perform %s: page is unhealthy", operation)
	}

	// Execute the operation
	if err := action(); err != nil {
		return fmt.Errorf("%s failed: %w", operation, err)
	}

	return nil
}

// -----------------------------------------------------------------------------
// Usage Examples
// -----------------------------------------------------------------------------

/*
Example usage of the page handler with browser integration:

```go
import (
	"log"
	"time"

	"tournois-tt/api/pkg/scraper/browser"
	"tournois-tt/api/pkg/scraper/page"
)

func Example() {
	// Initialize browser
	browserConfig := browser.DefaultConfig()
	browserInstance, pwInstance, err := browser.Init(browserConfig)
	if err != nil {
		log.Fatalf("Failed to initialize browser: %v", err)
	}
	defer browser.CleanupSingleton()

	// Create browser context
	context, err := browser.NewContext(browserInstance, browserConfig)
	if err != nil {
		log.Fatalf("Failed to create browser context: %v", err)
	}
	defer browser.SafeCloseContext(context)

	// Create a page
	pwPage, err := browser.NewPage(context)
	if err != nil {
		log.Fatalf("Failed to create page: %v", err)
	}

	// Create page handler
	pageHandler := page.New(pwPage)
	defer pageHandler.Close()

	// Set timeouts
	pageHandler.SetDefaultTimeouts(30*time.Second, 15*time.Second)

	// Navigate to a page with recovery capability
	err = pageHandler.SafeNavigation("https://example.com", 3, browser.RestartIfUnhealthy)
	if err != nil {
		log.Fatalf("Failed to navigate: %v", err)
	}

	// Use SafeOperation for reliable interactions
	err = pageHandler.SafeOperation("click button", func() error {
		return pageHandler.Click("#submit-button")
	})
	if err != nil {
		log.Printf("Operation failed: %v", err)
	}

	// Get content with error handling
	content, err := pageHandler.GetContent()
	if err != nil {
		log.Printf("Failed to get content: %v", err)
	}

	// Take a screenshot for debugging
	if err := pageHandler.TakeScreenshot("screenshot.png"); err != nil {
		log.Printf("Failed to take screenshot: %v", err)
	}
}
*/
