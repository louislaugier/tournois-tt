package browser

import (
	"fmt"

	"tournois-tt/api/pkg/scraper/config"

	pw "github.com/playwright-community/playwright-go"
)

// Init initializes and configures a new browser instance
func Init(cfg config.BrowserConfig) (pw.Browser, *pw.Playwright, error) {
	playwright, err := pw.Run()
	if err != nil {
		return nil, nil, fmt.Errorf("could not start playwright: %v", err)
	}

	browser, err := playwright.Chromium.Launch(pw.BrowserTypeLaunchOptions{
		Headless: pw.Bool(cfg.Headless),
		Args:     config.BrowserArgs(),
	})
	if err != nil {
		playwright.Stop()
		return nil, nil, fmt.Errorf("could not launch browser: %v", err)
	}

	return browser, playwright, nil
}

// NewContext creates a new browser context with the specified configuration
func NewContext(browser pw.Browser, cfg config.BrowserConfig) (pw.BrowserContext, error) {
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
