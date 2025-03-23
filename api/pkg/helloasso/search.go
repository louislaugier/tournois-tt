package helloasso

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"
	"tournois-tt/api/pkg/scraper/browser"
	"tournois-tt/api/pkg/scraper/page"

	pw "github.com/playwright-community/playwright-go"
)

// -----------------------------------------------------------------------------
// Constants and Configuration
// -----------------------------------------------------------------------------

const (
	// CacheSource identifies the cache source for HelloAsso
	CacheSource = "helloasso"

	// DefaultCacheExpiration is the default expiration duration for cached search results
	DefaultCacheExpiration = 24 * time.Hour
)

// -----------------------------------------------------------------------------
// Search Functions
// -----------------------------------------------------------------------------

// SearchActivities searches for activities on HelloAsso using the provided query
// This is a convenience wrapper around SearchActivitiesWithBrowser that handles browser initialization
func SearchActivities(ctx context.Context, query string) ([]Activity, error) {
	return SearchActivitiesWithBrowser(ctx, query, nil, nil)
}

// SearchActivitiesWithBrowser searches for activities on HelloAsso using the provided query and browser resources
// If sharedBrowserContext is provided, it will be used instead of creating a new browser
func SearchActivitiesWithBrowser(ctx context.Context, query string, sharedBrowserContext pw.BrowserContext, pwInstance *pw.Playwright) ([]Activity, error) {
	var browserInstance pw.Browser
	var browserContext pw.BrowserContext
	var ownedPwInstance *pw.Playwright
	var err error

	// If a browser context is provided, use it; otherwise, create a new browser instance
	if sharedBrowserContext != nil {
		browserContext = sharedBrowserContext
	} else {
		// Use our improved setup with a more maintainable config
		cfg := browser.DefaultConfig()

		// Add health check and timeout settings
		cfg.NavigationTimeout = 30 * time.Second
		cfg.OperationTimeout = 15 * time.Second

		// Setup browser
		browserInstance, ownedPwInstance, err = browser.Init(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to setup browser: %w", err)
		}

		// Ensure cleanup for browser resources
		defer func() {
			if ownedPwInstance != nil {
				ownedPwInstance.Stop()
			}
			if browserInstance != nil {
				browserInstance.Close()
			}
		}()

		// Validate browser health
		if !browser.IsHealthy() {
			return nil, fmt.Errorf("browser failed health check, unable to continue")
		}

		// Setup context with proper configuration
		browserContext, err = browser.NewContext(browserInstance, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to setup context: %w", err)
		}
		defer browser.SafeCloseContext(browserContext)
	}

	// Setup page (tab) with enhanced viewport settings
	playwrightPage, err := browser.NewPageWithViewport(browserContext, 1280, 800)
	if err != nil {
		return nil, fmt.Errorf("failed to setup page: %w", err)
	}
	defer browser.SafeClose(playwrightPage)

	// Create page handler
	pageHandler := page.NewPageHandler(playwrightPage)
	defer pageHandler.Close()

	// Set appropriate timeouts for HelloAsso which can be slow
	pageHandler.SetDefaultTimeouts(30*time.Second, 15*time.Second)

	// Build search URL
	encodedQuery := url.QueryEscape(query)
	searchURL := fmt.Sprintf(SearchURLTemplate, encodedQuery)
	log.Printf("Searching HelloAsso for: %s", query)

	// Navigate to search page with safe navigation (includes retries and health checks)
	if err := pageHandler.SafeNavigation(searchURL, 3, nil); err != nil {
		return nil, fmt.Errorf("failed to navigate to search page: %w", err)
	}

	// Wait for results with improved error handling
	pageConfig := page.Config{
		EmptyStateSelector: PageConfig.EmptyStateSelector,
		ResultsSelector:    PageConfig.ResultsSelector,
	}

	if err := pageHandler.WaitForResults(pageConfig); err != nil {
		// Take a screenshot for debugging
		screenshotPath := fmt.Sprintf("helloasso_search_error_%d.png", time.Now().Unix())
		if screenshotErr := pageHandler.TakeScreenshot(screenshotPath); screenshotErr == nil {
			log.Printf("Error screenshot saved to: %s", screenshotPath)
		}

		return nil, fmt.Errorf("failed to wait for results: %w", err)
	}

	// Check for empty state
	isEmpty, err := pageHandler.HasEmptyState(PageConfig.EmptyStateSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to check empty state: %w", err)
	}

	if isEmpty {
		log.Printf("No results found for query: %s", query)
		return []Activity{}, nil
	}

	// Extract activities from the page using SafeOperation for reliability
	var activities []Activity
	err = pageHandler.SafeOperation("extract activities", func() error {
		extractConfig := ExtractionConfig{
			BaseURL:            BaseURL,
			EmptyStateSelector: PageConfig.EmptyStateSelector,
			ActivitySelector:   PageConfig.ResultsSelector,
			Selectors:          Selectors,
		}

		var extractErr error
		activities, extractErr = ExtractActivities(playwrightPage, extractConfig)
		return extractErr
	})

	if err != nil {
		return nil, fmt.Errorf("failed to extract activities: %w", err)
	}

	log.Printf("Extracted %d activities from HelloAsso search for query: %s", len(activities), query)
	return activities, nil
}
