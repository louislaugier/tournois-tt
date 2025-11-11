package browser

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	pw "github.com/playwright-community/playwright-go"
)

// -----------------------------------------------------------------------------
// Constants and Configuration Types
// -----------------------------------------------------------------------------

// DefaultUserAgent is the user agent string to use for browser requests
const DefaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

// Default timeout values
const (
	DefaultNavigationTimeout = 30 * time.Second
	DefaultOperationTimeout  = 15 * time.Second
	HealthCheckTimeout       = 5 * time.Second
	DefaultPageTimeout       = 30000 // 30 seconds in milliseconds
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
	// CustomBrowserArgs are additional command-line arguments to pass to the browser
	CustomBrowserArgs []string
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
		CustomBrowserArgs: ContainerArgs(),
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

// ContainerArgs returns browser arguments optimized for container environments
func ContainerArgs() []string {
	// Default arguments that ensure the browser works in a container
	args := []string{
		"--disable-gpu",
		"--disable-software-rasterizer",
		"--disable-dev-shm-usage",
		"--disable-setuid-sandbox",
		"--no-sandbox",
		// Removed single-process mode as it can lead to browser instability
		// "--single-process",
		"--headless=new",
		"--disable-web-security",
		"--disable-features=AudioServiceOutOfProcess,IsolateOrigins,site-per-process",
		// Additional flags to disable GPU acceleration
		"--disable-accelerated-2d-canvas",
		"--disable-accelerated-jpeg-decoding",
		"--disable-accelerated-mjpeg-decode",
		"--disable-accelerated-video-decode",
		"--disable-webgl",
		"--disable-3d-apis",
		"--disable-accelerated-video",
		"--disable-gpu-compositing",
		"--disable-gpu-memory-buffer-video-frames",
		"--disable-gpu-rasterization",
		"--disable-gpu-sandbox",
		"--disable-webgl",
		"--ignore-gpu-blocklist",
		"--use-gl=swiftshader",
		// Enhanced memory management
		"--renderer-process-limit=1",
		"--memory-pressure-thresholds=16384,11263,5632",
		// Removed explicit browser-subprocess-path as we now set ExecutablePath directly
		// Crash handling
		"--disable-crash-reporter",
		"--enable-logging=stderr",
		"--v=1",
	}

	return args
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
		log.Printf("Browser initialization failed: %v", initErr)
		return nil, nil, fmt.Errorf("browser initialization failed: %w", initErr)
	}

	log.Println("Browser initialization successful, returning instances")
	return singletonBrowser, singletonPlaywright, nil
}

// initializePlaywright is an internal function to initialize the Playwright instance
func initializePlaywright(cfg Config) {
	// Check if browsers are already installed - which they should be from the Dockerfile
	// Check for browsers in multiple potential locations
	log.Println("Checking for pre-installed browsers at multiple locations...")

	// List all potential Chromium locations in Docker
	browserPaths := []string{
		"/usr/local/ms-playwright/chromium-1155",
		"/usr/local/ms-playwright/chromium_headless_shell-1155",
		"/root/.cache/ms-playwright/chromium-1155",
		"/.cache/ms-playwright/chromium-1155",
		"/ms-playwright/chromium-1155",
	}

	// Check each path and log if found
	chromiumExists := false
	var foundPath string
	for _, path := range browserPaths {
		if _, err := os.Stat(path); err == nil {
			log.Printf("Chromium browser found at: %s", path)
			chromiumExists = true
			foundPath = path
			break
		}
	}

	if chromiumExists {
		log.Printf("✅ Chromium browser is installed at: %s", foundPath)
		// Skip pw.Install() entirely - don't even try to run it
	} else {
		log.Println("⚠️ WARNING: Chromium browser not found in expected locations.")
		// List all directories in potential parent paths for diagnostic purposes
		for _, parent := range []string{"/usr/local/ms-playwright", "/root/.cache/ms-playwright", "/.cache/ms-playwright"} {
			log.Printf("Checking content of %s:", parent)
			files, err := os.ReadDir(parent)
			if err != nil {
				log.Printf("  Error reading directory: %v", err)
				continue
			}
			for _, file := range files {
				log.Printf("  - %s", file.Name())
			}
		}
	}

	// Start Playwright
	log.Println("Starting Playwright runtime")
	var err error
	singletonPlaywright, err = pw.Run()
	if err != nil {
		log.Printf("ERROR: Playwright start failed: %v", err)
		initErr = fmt.Errorf("playwright start failed: %w", err)
		return
	}
	log.Println("Playwright runtime started successfully")

	// Print available browsers for diagnostic purposes
	log.Println("DIAGNOSTICS: Available browser types:")
	if singletonPlaywright != nil {
		if singletonPlaywright.Chromium != nil {
			log.Println("- ✅ Chromium browser type is available (will be used)")
		} else {
			log.Println("- ❌ WARNING: Chromium browser type is NOT available")
		}
		if singletonPlaywright.Firefox != nil {
			log.Println("- Firefox browser type is available (will not be used)")
		}
		if singletonPlaywright.WebKit != nil {
			log.Println("- WebKit browser type is available (will not be used)")
		}
	}

	// Prepare launch options
	log.Println("Preparing browser launch options")
	browserArgs := Args()
	// Add container-specific args if provided
	if len(cfg.CustomBrowserArgs) > 0 {
		log.Printf("Using custom browser args (%d flags)", len(cfg.CustomBrowserArgs))
		browserArgs = cfg.CustomBrowserArgs
	}

	// Check if Chromium is available
	if singletonPlaywright.Chromium == nil {
		log.Println("ERROR: Chromium browser type is not available, but should be installed in the Docker image")
		initErr = fmt.Errorf("chromium browser type not available")
		return
	}

	// Set environment variables to ensure correct browser paths
	if foundPath != "" {
		os.Setenv("PLAYWRIGHT_BROWSERS_PATH", filepath.Dir(foundPath))
		log.Printf("Setting PLAYWRIGHT_BROWSERS_PATH to: %s", filepath.Dir(foundPath))
	}

	// Ensure executable permissions on any browser binaries
	if chromiumExists && foundPath != "" {
		chromeBin := filepath.Join(foundPath, "chrome-linux", "chrome")
		log.Printf("Checking Chrome binary at: %s", chromeBin)
		if _, err := os.Stat(chromeBin); err == nil {
			if err := os.Chmod(chromeBin, 0755); err != nil {
				log.Printf("Warning: Could not set executable permissions: %v", err)
			} else {
				log.Printf("Set executable permissions on Chrome binary")
			}
		}

		// Also check for headless_shell binary
		headlessBin := filepath.Join(foundPath, "chrome-linux", "headless_shell")
		if _, err := os.Stat(headlessBin); err == nil {
			if err := os.Chmod(headlessBin, 0755); err != nil {
				log.Printf("Warning: Could not set executable permissions on headless_shell: %v", err)
			} else {
				log.Printf("Set executable permissions on headless_shell binary")
			}
		}
	}

	launchOptions := pw.BrowserTypeLaunchOptions{
		Headless: pw.Bool(cfg.Headless),
		Args:     browserArgs,
		// Add more robust handle for potential browser crashes
		HandleSIGINT:  pw.Bool(false),
		HandleSIGTERM: pw.Bool(false),
		HandleSIGHUP:  pw.Bool(false),
		Timeout:       pw.Float(float64(60 * time.Second / time.Millisecond)), // Increase timeout for slow containers
	}

	// If we found a specific executable, use it
	if chromiumExists && foundPath != "" {
		execPath := filepath.Join(foundPath, "chrome-linux", "chrome")
		if _, err := os.Stat(execPath); err == nil {
			log.Printf("Using specific Chrome executable: %s", execPath)
			launchOptions.ExecutablePath = pw.String(execPath)
		} else {
			// Try headless_shell as fallback
			execPath = filepath.Join(foundPath, "chrome-linux", "headless_shell")
			if _, err := os.Stat(execPath); err == nil {
				log.Printf("Using headless_shell executable: %s", execPath)
				launchOptions.ExecutablePath = pw.String(execPath)
			}
		}
	}

	log.Println("Browser launch options configured")
	log.Printf("DIAGNOSTICS: Browser args: %v", browserArgs)

	// Launch browser with retries
	maxRetries := 3
	var lastErr error
	log.Println("Launching Chromium browser with retries")

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("Retrying browser launch (attempt %d/%d)", attempt+1, maxRetries)
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		}

		log.Printf("Launching CHROMIUM browser (attempt %d/%d)", attempt+1, maxRetries)
		// Explicitly use Chromium
		singletonBrowser, lastErr = singletonPlaywright.Chromium.Launch(launchOptions)
		if lastErr == nil {
			log.Println("Browser launched successfully")

			// Verify the browser type
			browserType := singletonBrowser.BrowserType()
			name := browserType.Name()
			log.Printf("DIAGNOSTICS: Launched browser type is: %s", name)
			if name != "chromium" {
				log.Printf("ERROR: Expected chromium browser but got %s", name)
				lastErr = fmt.Errorf("wrong browser type: %s", name)
				singletonBrowser.Close()
				continue
			}
			break
		}
		log.Printf("Browser launch attempt %d failed: %v", attempt+1, lastErr)
	}

	if lastErr != nil {
		log.Printf("ERROR: All browser launch attempts failed: %v", lastErr)
		singletonPlaywright.Stop()
		singletonPlaywright = nil
		initErr = fmt.Errorf("browser launch failed after %d attempts: %w", maxRetries, lastErr)
		return
	}

	log.Println("Browser singleton initialized successfully")
	log.Println("===== BROWSER INFO =====")
	log.Printf("Browser version: %s", singletonBrowser.Version())
	log.Printf("Browser type: %s", singletonBrowser.BrowserType().Name())
	log.Println("=======================")
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
	log.Println("Running enhanced health check for Docker...")

	// Quick check for existence first
	if singletonBrowser == nil || singletonPlaywright == nil {
		log.Println("Health check failed: browser or playwright instance is nil")
		return false
	}
	log.Println("Browser and Playwright instances exist")

	// For Docker environments, we'll try a basic page operation with improved error handling
	log.Println("Running Docker-optimized health check with improved diagnostics")

	// Create a channel that will only wait for a short time
	result := make(chan bool, 1)
	timeout := 10 * time.Second

	// Run the health check in a goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("RECOVERED FROM PANIC in health check: %v", r)
				result <- false
			}
		}()

		// Try to gather browser info for diagnostics
		log.Printf("Checking browser version...")
		version := singletonBrowser.Version()
		log.Printf("Browser version: %s", version)

		log.Printf("Checking browser type...")
		browserType := singletonBrowser.BrowserType().Name()
		log.Printf("Browser type: %s", browserType)

		// Create a context with very simple settings
		// log.Println("Creating minimal context for health check")
		// contextOptions := pw.BrowserNewContextOptions{
		// 	AcceptDownloads:   pw.Bool(false),
		// 	BypassCSP:         pw.Bool(true),
		// 	JavaScriptEnabled: pw.Bool(true),
		// }

		// // Try creating context
		// log.Println("Attempting to create browser context...")
		// context, err := singletonBrowser.NewContext(contextOptions)
		// if err != nil {
		// 	log.Printf("Health check: Failed to create context: %v", err)
		// 	result <- false
		// 	return
		// }

		// // Set up deferred cleanup
		// defer func() {
		// 	log.Println("Closing context used for health check")
		// 	if err := context.Close(); err != nil {
		// 		log.Printf("Warning: Error closing context during health check: %v", err)
		// 	}
		// }()

		// // Try creating a page
		// log.Println("Attempting to create page for health check")
		// page, err := context.NewPage()
		// if err != nil {
		// 	log.Printf("Health check: Failed to create page: %v", err)
		// 	result <- false
		// 	return
		// }

		// // Set up deferred cleanup
		// defer func() {
		// 	log.Println("Closing page used for health check")
		// 	if err := page.Close(); err != nil {
		// 		log.Printf("Warning: Error closing page during health check: %v", err)
		// 	}
		// }()

		// // Try navigating to about:blank as a basic test
		// log.Println("Attempting navigation to about:blank")
		// _, err = page.Goto("about:blank", pw.PageGotoOptions{
		// 	Timeout: pw.Float(5000), // Short timeout
		// })

		// if err != nil {
		// 	log.Printf("Health check: Failed to navigate to about:blank: %v", err)
		// 	result <- false
		// 	return
		// }

		// // Try to get page content as final test
		// log.Println("Attempting to get page content")
		// _, err = page.Content()
		// if err != nil {
		// 	log.Printf("Health check: Failed to get page content: %v", err)
		// 	result <- false
		// 	return
		// }

		// All tests passed
		log.Println("Health check: Successfully created context, page, and performed basic navigation")
		result <- true
	}()

	// Wait for result or timeout
	select {
	case isHealthy := <-result:
		return isHealthy
	case <-time.After(timeout):
		log.Printf("Health check timed out after %v", timeout)
		return false
	}
}

// isRunningInDocker checks if the application is running inside a Docker container
func isRunningInDocker() bool {
	_, err := os.Stat("/.dockerenv")
	return err == nil
}

// RestartIfUnhealthy checks the browser health and restarts it if needed
// Returns true if restart was needed and successful
func RestartIfUnhealthy() (bool, error) {
	if IsHealthy() {
		log.Println("Browser is healthy, no restart needed")
		return false, nil
	}

	log.Println("Browser is unhealthy, attempting restart")

	// Try multiple restart attempts with backoff
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			log.Printf("Restart attempt %d/3 after failure", attempt+1)
			time.Sleep(time.Duration(attempt*2) * time.Second)
		}

		// Clean up existing resources
		log.Println("Cleaning up browser resources before restart")
		CleanupSingleton()

		// Re-initialize with default configuration
		log.Println("Re-initializing browser with Docker-optimized config")
		config := DefaultConfig()
		browser, playwright, err := Init(config)

		if err != nil {
			log.Printf("Browser restart attempt %d failed: %v", attempt+1, err)
			continue
		}

		if browser == nil || playwright == nil {
			log.Printf("Browser restart attempt %d failed: browser or playwright is nil", attempt+1)
			continue
		}

		// Verify the browser is now healthy
		if !IsHealthy() {
			log.Printf("Browser health check failed after restart attempt %d", attempt+1)
			CleanupSingleton() // Clean up this failed attempt
			continue
		}

		log.Println("Browser successfully restarted")
		return true, nil
	}

	return false, fmt.Errorf("browser still unhealthy after multiple restart attempts")
}

// -----------------------------------------------------------------------------
// Context and Page Management
// -----------------------------------------------------------------------------

// NewContext creates a new browser context with the specified configuration
func NewContext(browser pw.Browser, cfg Config) (pw.BrowserContext, error) {
	if browser == nil {
		return nil, fmt.Errorf("invalid browser: nil")
	}

	// Create context options with anti-crawling measures
	contextOptions := pw.BrowserNewContextOptions{
		UserAgent:         pw.String(cfg.UserAgent),
		IgnoreHttpsErrors: pw.Bool(cfg.IgnoreHTTPSError),
		Locale:            pw.String(cfg.Locale),
		TimezoneId:        pw.String(cfg.TimeZoneID),
		// Add additional flags to evade bot detection
		JavaScriptEnabled: pw.Bool(true), // Ensure JavaScript is enabled
		HasTouch:          pw.Bool(false),
		IsMobile:          pw.Bool(false),
		// Extra HTTP headers that normal browsers would have
		ExtraHttpHeaders: map[string]string{
			"Accept":             "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
			"Accept-Language":    "fr-FR,fr;q=0.9,en-US;q=0.8,en;q=0.7",
			"Accept-Encoding":    "gzip, deflate, br",
			"Connection":         "keep-alive",
			"Sec-Fetch-Dest":     "document",
			"Sec-Fetch-Mode":     "navigate",
			"Sec-Fetch-Site":     "none",
			"Sec-Fetch-User":     "?1",
			"Sec-CH-UA":          "\"Chromium\";v=\"118\", \"Not-A.Brand\";v=\"99\"",
			"Sec-CH-UA-Mobile":   "?0",
			"Sec-CH-UA-Platform": "\"Windows\"",
		},
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

	// Add JavaScript to evade bot detection
	// This script overrides navigator properties commonly checked by anti-bot systems
	evasionScript := `
	() => {
		// Override properties used to detect automation
		const newProto = navigator.__proto__;
		delete newProto.webdriver;
		
		// Modify navigator properties
		Object.defineProperty(navigator, 'webdriver', {
			get: () => false
		});
		
		// Override Chrome's automation property
		window.chrome = {
			runtime: {},
			loadTimes: function() {},
			app: {},
			csi: function() {},
			// Add more Chrome-specific properties
		};
		
		// Add plugins array to mimic real browser
		Object.defineProperty(navigator, 'plugins', {
			get: () => [
				{ name: 'Chrome PDF Plugin', filename: 'internal-pdf-viewer' },
				{ name: 'Chrome PDF Viewer', filename: 'mhjfbmdgcfjbbpaeojofohoefgiehjai' },
				{ name: 'Native Client', filename: 'internal-nacl-plugin' }
			]
		});
		
		// Add language array to mimic real browser
		Object.defineProperty(navigator, 'languages', {
			get: () => ['fr-FR', 'fr', 'en-US', 'en']
		});
	}
	`

	// Add the evasion script to every new page in this context
	if err := context.AddInitScript(pw.Script{Content: pw.String(evasionScript)}); err != nil {
		log.Printf("Warning: Failed to add evasion script: %v", err)
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
// Uses anti-crawling techniques to avoid detection
func NavigateWithRetry(page pw.Page, url string, maxRetries int) error {
	if page == nil {
		return fmt.Errorf("invalid page: nil")
	}

	// Try each attempt with progressively more "human-like" behavior
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("Retrying navigation to %s (attempt %d/%d)", url, attempt, maxRetries)
			// Add exponential backoff with randomization to appear more human-like
			backoffTime := time.Duration(attempt*attempt)*500*time.Millisecond +
				time.Duration(attempt)*time.Second
			log.Printf("Waiting %v before retry (humanized delay)", backoffTime)
			time.Sleep(backoffTime)
		}

		// Add some random movements before navigation to appear more human-like
		if attempt > 0 {
			// Move mouse randomly
			width, height := 800, 600
			err := page.Mouse().Move(float64(width/2+attempt*10), float64(height/2+attempt*5))
			if err != nil {
				log.Printf("Mouse move failed (non-critical): %v", err)
			}

			// Small delay after mouse move
			time.Sleep(100 * time.Millisecond)
		}

		// Try to navigate to the URL with specific options to look more like a real browser
		log.Printf("Navigating to %s with human-like behavior", url)
		_, err := page.Goto(url, pw.PageGotoOptions{
			// Use domcontentloaded for first attempt, networkidle for retry attempts
			WaitUntil: pw.WaitUntilStateDomcontentloaded,
			Timeout:   pw.Float(float64(30 * time.Second / time.Millisecond)),
		})

		if err != nil {
			log.Printf("Navigation error: %v", err)
			lastErr = err

			// Check if we got a 403 forbidden or similar
			if attempt == 0 {
				// On first attempt failure, try to disguise ourselves better
				log.Printf("First attempt failed, attempting to bypass potential anti-crawling measures")

				// Execute some random scrolling to mimic human behavior
				scrollScript := `
					window.scrollBy({
						top: 100,
						left: 0,
						behavior: 'smooth'
					});
				`
				_, scrollErr := page.Evaluate(scrollScript)
				if scrollErr != nil {
					log.Printf("Scroll failed (non-critical): %v", scrollErr)
				}
			}

			continue
		}

		// After successful navigation, perform some human-like scrolling
		log.Printf("Navigation successful, performing human-like browsing behavior")

		// Small delay after page load before scrolling
		time.Sleep(500 * time.Millisecond)

		// Random scroll to simulate reading
		scrollScript := `
			window.scrollBy({
				top: 200,
				left: 0,
				behavior: 'smooth'
			});
		`
		_, scrollErr := page.Evaluate(scrollScript)
		if scrollErr != nil {
			log.Printf("Post-navigation scroll failed (non-critical): %v", scrollErr)
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

// checkBrowserExists checks if a browser is already installed by looking for its directory
func checkBrowserExists(browserPath string) bool {
	_, err := os.Stat(browserPath)
	if err == nil {
		log.Printf("Browser directory found at: %s", browserPath)
		return true
	}

	// Also check with glob pattern since the exact version might change
	matches, err := filepath.Glob(browserPath + "*")
	if err != nil {
		return false
	}
	if len(matches) > 0 {
		log.Printf("Browser directories found via glob: %v", matches)
		return true
	}
	return false
}

// FindChromiumPath attempts to find the Chromium installation directory
// Returns the path if found, empty string otherwise
func FindChromiumPath() string {
	// Common locations where Chromium might be installed
	paths := []string{
		"/usr/local/ms-playwright/chromium-1155",
		"/usr/local/ms-playwright/chromium_headless_shell-1155",
		"/root/.cache/ms-playwright/chromium-1155",
		"/.cache/ms-playwright/chromium-1155",
		"/ms-playwright/chromium-1155",
	}

	// First try exact paths
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			log.Printf("Found Chromium at exact path: %s", path)
			return path
		}
	}

	// Try with glob patterns for version flexibility
	parentDirs := []string{
		"/usr/local/ms-playwright",
		"/root/.cache/ms-playwright",
		"/.cache/ms-playwright",
		"/ms-playwright",
	}

	for _, dir := range parentDirs {
		// Check if the parent directory exists
		if _, err := os.Stat(dir); err != nil {
			continue
		}

		// Look for chromium directories
		chromiumDirs, err := filepath.Glob(filepath.Join(dir, "chromium*"))
		if err == nil && len(chromiumDirs) > 0 {
			log.Printf("Found Chromium via glob in %s: %v", dir, chromiumDirs)
			return chromiumDirs[0] // Return the first match
		}
	}

	// If no directories found, check inside /usr/local/ms-playwright
	mainDir := "/usr/local/ms-playwright"
	if dirEntries, err := os.ReadDir(mainDir); err == nil {
		for _, entry := range dirEntries {
			if entry.IsDir() && strings.Contains(strings.ToLower(entry.Name()), "chromium") {
				fullPath := filepath.Join(mainDir, entry.Name())
				log.Printf("Found Chromium directory by listing: %s", fullPath)
				return fullPath
			}
		}
	}

	return ""
}

// CheckBrowserInstallation checks if browsers are properly installed and logs the result
func CheckBrowserInstallation() bool {
	// Check multiple potential browser installation locations
	browserPaths := []string{
		"/ms-playwright",
		"/.cache/ms-playwright",
		"/root/.cache/ms-playwright",
	}

	for _, path := range browserPaths {
		if checkBrowserExists(path) {
			log.Printf("Playwright browsers directory exists at %s", path)
			return true
		}
	}

	log.Println("WARNING: Playwright browsers not found in common locations, but this is unexpected in Docker")
	log.Println("Proceeding without explicit installation - browsers should be pre-installed in Docker image")
	return false
}
