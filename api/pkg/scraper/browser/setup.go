package browser

import (
	"fmt"
	"log"
	"time"

	pw "github.com/playwright-community/playwright-go"
)

// -----------------------------------------------------------------------------
// Setup Functions
// -----------------------------------------------------------------------------

// Setup initializes a shared browser instance for web scraping operations.
// It returns the browser, playwright instance, browser context, and any error encountered.
// The caller is responsible for calling CleanupSingleton() when the browser is no longer needed.
func Setup() (pw.Browser, *pw.Playwright, pw.BrowserContext, error) {
	log.Println("Setting up browser environment for web scraping")

	// Use default configuration
	cfg := DefaultConfig()

	// Add health check timeouts
	if cfg.NavigationTimeout == 0 {
		cfg.NavigationTimeout = DefaultNavigationTimeout
	}
	if cfg.OperationTimeout == 0 {
		cfg.OperationTimeout = DefaultOperationTimeout
	}

	// Initialize browser instance using the singleton pattern
	browserInstance, pwInstance, err := Init(cfg)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("browser initialization failed: %w", err)
	}

	// Perform initial health check
	if !IsHealthy() {
		pwInstance.Stop()
		browserInstance.Close()
		return nil, nil, nil, fmt.Errorf("browser failed initial health check")
	}

	// Create a browser context that will be shared among all operations
	browserContext, err := NewContext(browserInstance, cfg)
	if err != nil {
		// Clean up resources on error
		log.Println("Failed to create browser context, cleaning up resources")
		pwInstance.Stop()
		browserInstance.Close()
		return nil, nil, nil, fmt.Errorf("browser context creation failed: %w", err)
	}

	// Set default timeouts
	if cfg.NavigationTimeout > 0 {
		browserContext.SetDefaultNavigationTimeout(float64(cfg.NavigationTimeout / time.Millisecond))
	}
	if cfg.OperationTimeout > 0 {
		browserContext.SetDefaultTimeout(float64(cfg.OperationTimeout / time.Millisecond))
	}

	log.Println("Browser setup completed successfully")
	return browserInstance, pwInstance, browserContext, nil
}

// SetupWithRetry attempts to set up the browser with retries on failure
// This is useful for environments where the browser might fail to initialize on first attempt
func SetupWithRetry(maxRetries int) (pw.Browser, *pw.Playwright, pw.BrowserContext, error) {
	var browser pw.Browser
	var playwright *pw.Playwright
	var context pw.BrowserContext
	var err error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("Retrying browser setup (attempt %d/%d)", attempt, maxRetries)
			// Add exponential backoff
			backoffTime := time.Duration(attempt*attempt) * 500 * time.Millisecond
			time.Sleep(backoffTime)
		}

		browser, playwright, context, err = Setup()
		if err == nil {
			return browser, playwright, context, nil
		}

		log.Printf("Browser setup failed: %v", err)
	}

	return nil, nil, nil, fmt.Errorf("browser setup failed after %d attempts: %w", maxRetries, err)
}

// ShutdownBrowser properly cleans up all browser resources.
// This is a convenience wrapper around CleanupSingleton that should be called
// before the application exits.
func ShutdownBrowser() {
	log.Println("Shutting down browser environment")
	CleanupSingleton()
}

// EnsureHealthyBrowser verifies the browser is healthy and restarts it if needed
// Returns true if the browser was restarted
func EnsureHealthyBrowser() (bool, error) {
	if !IsHealthy() {
		log.Println("Browser is unhealthy, restarting...")
		CleanupSingleton()

		// Reinitialize with retry
		_, _, _, err := SetupWithRetry(2)
		if err != nil {
			return false, fmt.Errorf("failed to restart browser: %w", err)
		}
		return true, nil
	}
	return false, nil
}
