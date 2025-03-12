package scraper

import (
	"context"
	"fmt"
	"net/url"

	"tournois-tt/api/pkg/scraper/browser"
	"tournois-tt/api/pkg/scraper/config"
	"tournois-tt/api/pkg/scraper/extractor"
	"tournois-tt/api/pkg/scraper/logger"
	"tournois-tt/api/pkg/scraper/models"
	"tournois-tt/api/pkg/scraper/page"
)

const (
	helloAssoBaseURL           = "https://www.helloasso.com"
	helloAssoSearchURLTemplate = "https://www.helloasso.com/e/recherche?query=%s"
)

// HelloAssoSelectors contains all the CSS selectors used for HelloAsso scraping
var helloAssoSelectors = extractor.ActivitySelectors{
	Title:        ".Thumbnail--Name",
	Date:         ".Thumbnail--Date",
	URL:          "a",
	Price:        ".Thumbnail--ImagePill",
	Organization: ".Thumbnail--OrganizationName",
	Category:     ".Thumbnail--MetadataTag",
	Location:     ".Thumbnail--MetadataLocation",
}

// HelloAssoConfig contains all the configuration for HelloAsso scraping
var helloAssoConfig = page.Config{
	EmptyStateSelector: `[data-testid="empty-state"]`,
	ResultsSelector:    ".Hits-Activity",
}

// SearchHelloAssoActivities searches for activities on HelloAsso using the provided query
func SearchHelloAssoActivities(ctx context.Context, query string) ([]models.Activity, error) {
	cfg := config.DefaultConfig()

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

	// Setup logging
	// logger.SetupPageLogging(playwrightPage)

	// Create page handler
	pageHandler := page.New(playwrightPage)

	// Build search URL
	encodedQuery := url.QueryEscape(query)
	searchURL := fmt.Sprintf(helloAssoSearchURLTemplate, encodedQuery)

	// Navigate to search page
	if err := pageHandler.NavigateToPage(searchURL); err != nil {
		return nil, fmt.Errorf("failed to navigate to search page: %v", err)
	}

	// Wait for results
	if err := pageHandler.WaitForResults(helloAssoConfig); err != nil {
		return nil, fmt.Errorf("failed to wait for results: %v", err)
	}

	// Check for empty state
	isEmpty, err := pageHandler.HasEmptyState(helloAssoConfig.EmptyStateSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to check empty state: %v", err)
	}
	if isEmpty {
		logger.Info("No results found")
		return []models.Activity{}, nil
	}

	// Extract activities from the page
	extractConfig := extractor.ExtractionConfig{
		BaseURL:            helloAssoBaseURL,
		EmptyStateSelector: helloAssoConfig.EmptyStateSelector,
		ActivitySelector:   ".Thumbnail.Thumbnail-Activity",
		Selectors:          helloAssoSelectors,
	}
	return extractor.ExtractActivities(playwrightPage, extractConfig)
}
