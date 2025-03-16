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

const (
	// CacheSource identifies the cache source for HelloAsso
	CacheSource = "helloasso"

	// DefaultCacheExpiration is the default expiration duration for cached search results
	DefaultCacheExpiration = 24 * time.Hour
)

// SearchActivities searches for activities on HelloAsso using the provided query
func SearchActivities(ctx context.Context, query string) ([]Activity, error) {
	return SearchActivitiesWithBrowser(ctx, query, nil, nil)
}

// SearchActivitiesWithBrowser searches for activities on HelloAsso using the provided query and browser resources
func SearchActivitiesWithBrowser(ctx context.Context, query string, sharedBrowserContext pw.BrowserContext, pwInstance *pw.Playwright) ([]Activity, error) {
	var browserInstance pw.Browser
	var browserContext pw.BrowserContext
	var ownedPwInstance *pw.Playwright
	var err error

	// If a browser context is provided, use it; otherwise, create a new browser instance
	if sharedBrowserContext != nil {
		browserContext = sharedBrowserContext
	} else {
		cfg := browser.DefaultConfig()

		// Setup browser
		browserInstance, ownedPwInstance, err = browser.Init(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to setup browser: %v", err)
		}
		defer func() {
			if ownedPwInstance != nil {
				ownedPwInstance.Stop()
			}
			if browserInstance != nil {
				browserInstance.Close()
			}
		}()

		// Setup context
		browserContext, err = browser.NewContext(browserInstance, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to setup context: %v", err)
		}
		defer browserContext.Close()
	}

	// Setup page (tab)
	playwrightPage, err := browser.NewPage(browserContext)
	if err != nil {
		return nil, fmt.Errorf("failed to setup page: %v", err)
	}

	// Create page handler
	pageHandler := page.New(playwrightPage)

	// Build search URL
	encodedQuery := url.QueryEscape(query)
	searchURL := fmt.Sprintf(SearchURLTemplate, encodedQuery)

	// Navigate to search page
	if err := pageHandler.NavigateToPage(searchURL); err != nil {
		return nil, fmt.Errorf("failed to navigate to search page: %v", err)
	}

	// Wait for results
	if err := pageHandler.WaitForResults(Config); err != nil {
		return nil, fmt.Errorf("failed to wait for results: %v", err)
	}

	// Check for empty state
	isEmpty, err := pageHandler.HasEmptyState(Config.EmptyStateSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to check empty state: %v", err)
	}
	if isEmpty {
		return []Activity{}, nil
	}

	// Extract activities from the page
	extractConfig := ExtractionConfig{
		BaseURL:            BaseURL,
		EmptyStateSelector: Config.EmptyStateSelector,
		ActivitySelector:   Config.ResultsSelector,
		Selectors:          Selectors,
	}
	activities, err := ExtractActivities(playwrightPage, extractConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to extract activities: %v", err)
	}

	log.Printf("Extracted %d activities from HelloAsso search", len(activities))

	return activities, nil
}
