package browser

import (
	"fmt"
	"sync"

	pw "github.com/playwright-community/playwright-go"
)

// Config holds the configuration for browser setup
type Config struct {
	Headless  bool
	UserAgent string
}

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

// DefaultConfig returns the default browser configuration
func DefaultConfig() Config {
	return Config{
		Headless:  true,
		UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
	}
}

// BrowserArgs returns the default browser launch arguments
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

// Init initializes and configures a new browser instance
// The first call initializes the singleton browser, subsequent calls return the same instance
func Init(cfg Config) (pw.Browser, *pw.Playwright, error) {
	initOnce.Do(func() {
		var err error
		singletonPlaywright, err = pw.Run()
		if err != nil {
			initErr = fmt.Errorf("could not start playwright: %v", err)
			return
		}

		singletonBrowser, err = singletonPlaywright.Chromium.Launch(pw.BrowserTypeLaunchOptions{
			Headless: pw.Bool(cfg.Headless),
			Args:     Args(),
		})
		if err != nil {
			singletonPlaywright.Stop()
			initErr = fmt.Errorf("could not launch browser: %v", err)
			return
		}
	})

	if initErr != nil {
		return nil, nil, initErr
	}

	return singletonBrowser, singletonPlaywright, nil
}

// NewContext creates a new browser context with the specified configuration
func NewContext(browser pw.Browser, cfg Config) (pw.BrowserContext, error) {
	context, err := browser.NewContext(pw.BrowserNewContextOptions{
		UserAgent: pw.String(cfg.UserAgent),
	})
	if err != nil {
		return nil, fmt.Errorf("could not create context: %v", err)
	}

	return context, nil
}

// NewPage creates a new page
func NewPage(context pw.BrowserContext) (pw.Page, error) {
	page, err := context.NewPage()
	if err != nil {
		return nil, fmt.Errorf("could not create page: %v", err)
	}

	return page, nil
}

// CleanupSingleton properly cleans up the singleton browser and playwright instance
// This should be called before the application exits
func CleanupSingleton() {
	if singletonBrowser != nil {
		singletonBrowser.Close()
		singletonBrowser = nil
	}

	if singletonPlaywright != nil {
		singletonPlaywright.Stop()
		singletonPlaywright = nil
	}

	// Reset the init once so it can be initialized again if needed
	initOnce = sync.Once{}
}
