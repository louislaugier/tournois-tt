package browser

import (
	"fmt"
	"log"
	"sync"
	"time"

	pw "github.com/playwright-community/playwright-go"
)

// -----------------------------------------------------------------------------
// Constants and Configuration Types
// -----------------------------------------------------------------------------

// DefaultUserAgent is the user agent string to use for browser requests
const DefaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36"

// Default timeout values
const (
	DefaultNavigationTimeout = 30 * time.Second
	DefaultOperationTimeout  = 15 * time.Second
	HealthCheckTimeout       = 5 * time.Second
)

// Config holds the configuration for browser setup
type Config struct {
	// Headless determines if the browser runs without a visible UI
	Headless bool
	// UserAgent is the user agent string for HTTP requests
	UserAgent string
	// Proxy is the proxy server URL in format http://myproxy.com:3128 or socks5://myproxy.com:3128
	Proxy string
	// IgnoreHTTPSError determines if HTTPS errors should be ignored
	IgnoreHTTPSError bool
	// ViewportWidth is the initial viewport width
	ViewportWidth int
	// ViewportHeight is the initial viewport height
	ViewportHeight int
	// Locale is the browser's locale like 'en-US', 'fr-FR'
	Locale string
	// TimeZoneID is the timezone ID like 'Europe/Paris'
	TimeZoneID string
	// NavigationTimeout is the timeout for page navigation operations
	NavigationTimeout time.Duration
	// OperationTimeout is the timeout for general browser operations
	OperationTimeout time.Duration
}

// DefaultConfig returns the default browser configuration
func DefaultConfig() Config {
	return Config{
		Headless:          true,
		UserAgent:         DefaultUserAgent,
		IgnoreHTTPSError:  true,
		ViewportWidth:     1280,
		ViewportHeight:    800,
		Locale:            "fr-FR",
		TimeZoneID:        "Europe/Paris",
		NavigationTimeout: DefaultNavigationTimeout,
		OperationTimeout:  DefaultOperationTimeout,
	}
}

// -----------------------------------------------------------------------------
// Browser Arguments
// -----------------------------------------------------------------------------

// Args returns the default browser launch arguments
func Args() []string {
	return []string{
		"--disable-dev-shm-usage",
		"--no-sandbox",
		"--disable-setuid-sandbox",
		"--disable-gpu",
		"--no-first-run",
		"--no-zygote",
		"--single-process",
		"--disable-extensions",
	}
}

// -----------------------------------------------------------------------------
// Singleton Management
// -----------------------------------------------------------------------------

var (
	// singletonBrowser is the shared browser instance
	singletonBrowser pw.Browser

	// singletonPlaywright is the shared playwright instance
	singletonPlaywright *pw.Playwright

	// initOnce ensures the browser is initialized only once
	initOnce sync.Once

	// initErr stores any error that occurred during initialization
	initErr error
)

// Init initializes and configures a new browser instance
// The first call initializes the singleton browser, subsequent calls return the same instance
func Init(cfg Config) (pw.Browser, *pw.Playwright, error) {
	initOnce.Do(func() {
		log.Println("Initializing Playwright browser singleton")
		initializePlaywright(cfg)
	})

	if initErr != nil {
		return nil, nil, fmt.Errorf("browser initialization failed: %w", initErr)
	}

	return singletonBrowser, singletonPlaywright, nil
}

// initializePlaywright is an internal function to initialize the Playwright instance
func initializePlaywright(cfg Config) {
	// Install Playwright if not already installed
	if err := pw.Install(); err != nil {
		initErr = fmt.Errorf("playwright installation failed: %w", err)
		return
	}

	// Start Playwright
	var err error
	singletonPlaywright, err = pw.Run()
	if err != nil {
		initErr = fmt.Errorf("playwright start failed: %w", err)
		return
	}

	// Prepare launch options
	launchOptions := pw.BrowserTypeLaunchOptions{
		Headless: pw.Bool(cfg.Headless),
		Args:     Args(),
	}

	// Launch browser
	singletonBrowser, err = singletonPlaywright.Chromium.Launch(launchOptions)
	if err != nil {
		singletonPlaywright.Stop()
		singletonPlaywright = nil
		initErr = fmt.Errorf("browser launch failed: %w", err)
		return
	}

	log.Println("Browser singleton initialized successfully")
}

// CleanupSingleton properly cleans up the singleton browser and playwright instance
// This should be called before the application exits
func CleanupSingleton() {
	log.Println("Cleaning up browser resources")

	if singletonBrowser != nil {
		if err := singletonBrowser.Close(); err != nil {
			log.Printf("Warning: Failed to close browser: %v", err)
		}
		singletonBrowser = nil
	}

	if singletonPlaywright != nil {
		singletonPlaywright.Stop()
		singletonPlaywright = nil
	}

	// Reset the init once so it can be initialized again if needed
	initOnce = sync.Once{}
	initErr = nil

	log.Println("Browser resources cleaned up")
}

// -----------------------------------------------------------------------------
// Browser Health Checks
// -----------------------------------------------------------------------------

// IsHealthy checks if the browser instance is healthy and responsive
func IsHealthy() bool {
	if singletonBrowser == nil || singletonPlaywright == nil {
		return false
	}

	// Try to create a context and page as a health check
	config := DefaultConfig()

	browserContext, err := singletonBrowser.NewContext(pw.BrowserNewContextOptions{
		UserAgent: pw.String(config.UserAgent),
	})
	if err != nil {
		log.Printf("Health check failed: context creation error: %v", err)
		return false
	}
	defer browserContext.Close()

	// Try to create a page
	page, err := browserContext.NewPage()
	if err != nil {
		log.Printf("Health check failed: page creation error: %v", err)
		return false
	}
	defer page.Close()

	// Try to navigate to about:blank with a timeout
	options := pw.PageGotoOptions{
		Timeout: pw.Float(float64(HealthCheckTimeout / time.Millisecond)),
	}
	if _, err := page.Goto("about:blank", options); err != nil {
		log.Printf("Health check failed: navigation error: %v", err)
		return false
	}

	// If we get here, the browser is responsive
	return true
}

// RestartIfUnhealthy checks the browser health and restarts it if needed
// Returns true if restart was needed and successful
func RestartIfUnhealthy() (bool, error) {
	if IsHealthy() {
		return false, nil
	}

	log.Println("Browser is unhealthy, attempting restart")

	// Clean up existing resources
	CleanupSingleton()

	// Re-initialize with default configuration
	config := DefaultConfig()
	_, _, err := Init(config)
	if err != nil {
		return false, fmt.Errorf("browser restart failed: %w", err)
	}

	// Verify the browser is now healthy
	if !IsHealthy() {
		return false, fmt.Errorf("browser still unhealthy after restart")
	}

	log.Println("Browser successfully restarted")
	return true, nil
}

// -----------------------------------------------------------------------------
// Context and Page Management
// -----------------------------------------------------------------------------

// NewContext creates a new browser context with the specified configuration
func NewContext(browser pw.Browser, cfg Config) (pw.BrowserContext, error) {
	if browser == nil {
		return nil, fmt.Errorf("invalid browser: nil")
	}

	// Create context options
	contextOptions := pw.BrowserNewContextOptions{
		UserAgent:         pw.String(cfg.UserAgent),
		IgnoreHttpsErrors: pw.Bool(cfg.IgnoreHTTPSError),
		Locale:            pw.String(cfg.Locale),
		TimezoneId:        pw.String(cfg.TimeZoneID),
	}

	// Add proxy if specified
	if cfg.Proxy != "" {
		contextOptions.Proxy = &pw.Proxy{
			Server: cfg.Proxy,
		}
	}

	// Create the context
	context, err := browser.NewContext(contextOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create browser context: %w", err)
	}

	// Set default timeouts if specified
	if cfg.NavigationTimeout > 0 {
		context.SetDefaultNavigationTimeout(float64(cfg.NavigationTimeout / time.Millisecond))
	}

	if cfg.OperationTimeout > 0 {
		context.SetDefaultTimeout(float64(cfg.OperationTimeout / time.Millisecond))
	}

	return context, nil
}

// NewPage creates a new page with the default viewport size (1280x800)
func NewPage(context pw.BrowserContext) (pw.Page, error) {
	if context == nil {
		return nil, fmt.Errorf("invalid context: nil")
	}

	// Create the page
	page, err := context.NewPage()
	if err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}

	// Set default viewport size
	if err = page.SetViewportSize(1280, 800); err != nil {
		page.Close()
		return nil, fmt.Errorf("failed to set viewport size: %w", err)
	}

	return page, nil
}

// NewPageWithViewport creates a new page with a specific viewport size
func NewPageWithViewport(context pw.BrowserContext, width, height int) (pw.Page, error) {
	if context == nil {
		return nil, fmt.Errorf("invalid context: nil")
	}

	// Validate dimensions
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid viewport dimensions: width=%d, height=%d", width, height)
	}

	// Create the page
	page, err := context.NewPage()
	if err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}

	// Set viewport size
	if err = page.SetViewportSize(width, height); err != nil {
		page.Close()
		return nil, fmt.Errorf("failed to set viewport size: %w", err)
	}

	return page, nil
}

// -----------------------------------------------------------------------------
// Page Navigation Helpers
// -----------------------------------------------------------------------------

// NavigateWithRetry attempts to navigate to a URL with retries on failure
func NavigateWithRetry(page pw.Page, url string, maxRetries int) error {
	if page == nil {
		return fmt.Errorf("invalid page: nil")
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("Retrying navigation to %s (attempt %d/%d)", url, attempt, maxRetries)
			// Add exponential backoff
			backoffTime := time.Duration(attempt*attempt) * 500 * time.Millisecond
			time.Sleep(backoffTime)
		}

		// Try to navigate to the URL
		_, err := page.Goto(url)
		if err != nil {
			lastErr = err
			continue
		}

		// Success
		return nil
	}

	return fmt.Errorf("navigation failed after %d attempts: %w", maxRetries, lastErr)
}

// SafeClose attempts to safely close a page, handling any errors
func SafeClose(page pw.Page) {
	if page == nil {
		return
	}

	if err := page.Close(); err != nil {
		log.Printf("Warning: Failed to close page: %v", err)
	}
}

// SafeCloseContext attempts to safely close a browser context, handling any errors
func SafeCloseContext(context pw.BrowserContext) {
	if context == nil {
		return
	}

	if err := context.Close(); err != nil {
		log.Printf("Warning: Failed to close browser context: %v", err)
	}
}

// -----------------------------------------------------------------------------
// Usage Examples
// -----------------------------------------------------------------------------

// FetchPageContent safely navigates to a URL and returns the page content
// This demonstrates best practices for using the browser API
func FetchPageContent(url string, retries int) (string, error) {
	// Check if browser needs to be initialized or recovered
	if !IsHealthy() {
		restarted, err := RestartIfUnhealthy()
		if err != nil {
			return "", fmt.Errorf("browser health check failed: %w", err)
		}
		if restarted {
			log.Println("Browser was automatically restarted before navigation")
		}
	}

	// Create a lightweight configuration
	cfg := DefaultConfig()

	// Get the browser instance
	browser, _, err := Init(cfg)
	if err != nil {
		return "", fmt.Errorf("browser initialization failed: %w", err)
	}

	// Create a browser context
	context, err := NewContext(browser, cfg)
	if err != nil {
		return "", fmt.Errorf("context creation failed: %w", err)
	}
	defer SafeCloseContext(context)

	// Create a new page
	page, err := NewPage(context)
	if err != nil {
		return "", fmt.Errorf("page creation failed: %w", err)
	}
	defer SafeClose(page)

	// Navigate to the URL with retry capability
	if err := NavigateWithRetry(page, url, retries); err != nil {
		return "", fmt.Errorf("navigation failed: %w", err)
	}

	// Get the page content
	content, err := page.Content()
	if err != nil {
		return "", fmt.Errorf("failed to get page content: %w", err)
	}

	return content, nil
}
