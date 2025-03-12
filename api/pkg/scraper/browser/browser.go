package browser

import (
	"fmt"

	pw "github.com/playwright-community/playwright-go"
)

// Config holds the configuration for browser setup
type Config struct {
	Headless  bool
	UserAgent string
}

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
func Init(cfg Config) (pw.Browser, *pw.Playwright, error) {
	playwright, err := pw.Run()
	if err != nil {
		return nil, nil, fmt.Errorf("could not start playwright: %v", err)
	}

	browser, err := playwright.Chromium.Launch(pw.BrowserTypeLaunchOptions{
		Headless: pw.Bool(cfg.Headless),
		Args:     Args(),
	})
	if err != nil {
		playwright.Stop()
		return nil, nil, fmt.Errorf("could not launch browser: %v", err)
	}

	return browser, playwright, nil
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
