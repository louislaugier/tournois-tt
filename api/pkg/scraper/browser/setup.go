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
	log.Println("Setting up browser environment for web scraping (Docker-optimized)")

	// Use container-optimized configuration
	cfg := DefaultConfig()
	log.Println("Using container-optimized configuration")

	// Ensure container-optimized arguments are used
	cfg.CustomBrowserArgs = ContainerArgs()
	log.Println("Container arguments set with", len(cfg.CustomBrowserArgs), "flags")

	// Add health check timeouts
	cfg.NavigationTimeout = 60 * time.Second // Increase timeouts for more stability
	cfg.OperationTimeout = 30 * time.Second
	log.Printf("Timeouts set: navigation=%v, operation=%v", cfg.NavigationTimeout, cfg.OperationTimeout)

	// Reset singleton state to avoid potential reuse of corrupted browser
	CleanupSingleton()
	log.Println("Ensured clean state by resetting any existing browser")

	// Initialize browser instance using the singleton pattern
	log.Println("About to initialize browser singleton")
	browserInstance, pwInstance, err := Init(cfg)
	if err != nil {
		log.Printf("ERROR: Browser initialization failed: %v", err)
		return nil, nil, nil, fmt.Errorf("browser initialization failed: %w", err)
	}
	log.Println("Browser instance and Playwright initialized successfully")

	// Verify we got a Chromium browser
	browserType := browserInstance.BrowserType()
	browserName := browserType.Name()
	log.Printf("DIAGNOSTICS: Browser type being used is: %s", browserName)
	if browserName != "chromium" {
		log.Printf("ERROR: Expected Chromium browser but got %s browser", browserName)
		// Clean up resources
		browserInstance.Close()
		pwInstance.Stop()
		return nil, nil, nil, fmt.Errorf("wrong browser type: %s", browserName)
	}

	// Create a browser context that will be shared among all operations
	log.Println("Creating browser context...")
	browserContext, err := NewContext(browserInstance, cfg)
	if err != nil {
		// Clean up resources on error
		log.Printf("ERROR: Failed to create browser context: %v", err)
		log.Println("Cleaning up resources after context creation failure")
		pwInstance.Stop()
		browserInstance.Close()
		return nil, nil, nil, fmt.Errorf("browser context creation failed: %w", err)
	}
	log.Println("Browser context created successfully")

	// Set default timeouts
	log.Println("Setting navigation and operation timeouts on context")
	if cfg.NavigationTimeout > 0 {
		browserContext.SetDefaultNavigationTimeout(float64(cfg.NavigationTimeout / time.Millisecond))
	}
	if cfg.OperationTimeout > 0 {
		browserContext.SetDefaultTimeout(float64(cfg.OperationTimeout / time.Millisecond))
	}
	log.Println("Context timeouts set successfully")

	// Skip the verification page creation and instead rely on IsHealthy
	log.Println("Verifying browser is fully functioning using enhanced health check")
	if !IsHealthy() {
		log.Println("ERROR: Browser verification failed - browser is not healthy")
		log.Println("Cleaning up browser resources after verification failure")
		browserContext.Close()
		browserInstance.Close()
		pwInstance.Stop()
		return nil, nil, nil, fmt.Errorf("browser verification failed: browser is not healthy")
	}
	log.Println("Browser verification completed successfully")

	// Skip the additional test page creation since we already verified it's healthy
	log.Println("Using successfully verified browser setup")
	log.Println("DIAGNOSTICS: Running in Docker environment")

	// List contexts for diagnostics
	contexts := browserInstance.Contexts()
	log.Printf("DIAGNOSTICS: Number of browser contexts: %d", len(contexts))

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

	log.Printf("Setting up browser with max %d retries", maxRetries)
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("Retrying browser setup (attempt %d/%d)", attempt, maxRetries)
			// Add exponential backoff
			backoffTime := time.Duration(attempt*attempt) * 500 * time.Millisecond
			log.Printf("Waiting %v before retry", backoffTime)
			time.Sleep(backoffTime)

			// Ensure we clean up any failed previous attempt
			log.Println("Cleaning up from previous attempt")
			CleanupSingleton()
		}

		log.Printf("Starting browser setup attempt %d/%d", attempt+1, maxRetries)
		browser, playwright, context, err = Setup()
		if err == nil {
			// Successfully set up the browser
			log.Printf("Browser setup successful on attempt %d", attempt+1)
			return browser, playwright, context, nil
		}

		log.Printf("Browser setup failed on attempt %d: %v", attempt+1, err)
	}

	log.Printf("ERROR: Browser setup failed after %d attempts: %v", maxRetries, err)
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
	log.Println("Checking browser health")
	if !IsHealthy() {
		log.Println("Browser is unhealthy, restarting...")
		CleanupSingleton()

		// Reinitialize with retry
		log.Println("Reinitializing browser with retry")
		_, _, _, err := SetupWithRetry(2)
		if err != nil {
			log.Printf("ERROR: Failed to restart browser: %v", err)
			return false, fmt.Errorf("failed to restart browser: %w", err)
		}
		log.Println("Browser successfully restarted")
		return true, nil
	}
	log.Println("Browser is healthy")
	return false, nil
}
