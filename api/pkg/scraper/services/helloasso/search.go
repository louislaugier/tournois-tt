package helloasso

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"
	"tournois-tt/api/pkg/scraper/browser"
	"tournois-tt/api/pkg/scraper/page"
)

const (
	// CacheSource identifies the cache source for HelloAsso
	CacheSource = "helloasso"

	// DefaultCacheExpiration is the default expiration duration for cached search results
	DefaultCacheExpiration = 24 * time.Hour
)

// SearchActivities searches for activities on HelloAsso using the provided query
func SearchActivities(ctx context.Context, query string) ([]Activity, error) {
	cfg := browser.DefaultConfig()

	// Setup browser
	browserInstance, pwInstance, err := browser.Init(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to setup browser: %v", err)
	}
	defer pwInstance.Stop()
	defer browserInstance.Close()

	// Setup context
	browserContext, err := browser.NewContext(browserInstance, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to setup context: %v", err)
	}
	defer browserContext.Close()

	// Setup page
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
