package helloasso

import (
	"fmt"
	"log"
	"strings"
	"time"
	"tournois-tt/api/pkg/scraper/page"
	"tournois-tt/api/pkg/utils"

	pw "github.com/playwright-community/playwright-go"
)

// -----------------------------------------------------------------------------
// Constants and Configuration
// -----------------------------------------------------------------------------

const BaseURL = "https://www.helloasso.com"

var SearchURLTemplate = fmt.Sprintf("%s/e/recherche?query=%%s", BaseURL)

// Selectors contains all the CSS selectors used for HelloAsso scraping
var Selectors = ActivitySelectors{
	Title:        ".Thumbnail--Name, .h-hWXvSR4d4s2KhyqM1HZ, .activity-card-name",
	Date:         ".Thumbnail--Date, .h-13fYiT0HkOyGPRHF5JRe, .activity-card-date",
	URL:          "a",
	Price:        ".Thumbnail--ImagePill, .h-6iSMgI2kjIKY4skQH4QN, .activity-card-price",
	Organization: ".Thumbnail--OrganizationName, .h-ibr9SVVUL6_n4JGkqaSo, .activity-card-organization",
	Category:     ".Thumbnail--MetadataTag, .h-4YUZwOrnMEbhZDfnPRLv, .activity-card-category",
	Location:     ".Thumbnail--MetadataLocation, .h-hVDkACpKtEyaNE5Gx6BH, .activity-card-location",
}

// PageConfig contains all the configuration for HelloAsso scraping
var PageConfig = struct {
	EmptyStateSelector string
	ResultsSelector    string
}{
	EmptyStateSelector: `[data-testid="empty-state"], .h-39a8TrXCLKJGJxr4DEp6, .no-results`,
	ResultsSelector:    ".Hits-Activity, .h-k2MJThJUO3PScbPTfyXD, .activity-card",
}

// -----------------------------------------------------------------------------
// Types and Structures
// -----------------------------------------------------------------------------

// ExtractionConfig holds the configuration for extracting activities
type ExtractionConfig struct {
	BaseURL            string
	EmptyStateSelector string
	ActivitySelector   string
	Selectors          ActivitySelectors
}

// ActivitySelectors holds the CSS selectors for extracting activity data
type ActivitySelectors struct {
	Title        string
	Date         string
	URL          string
	Price        string
	Organization string
	Category     string
	Location     string
}

// HelloAssoConfig stores configuration for HelloAsso extractors
type HelloAssoConfig struct {
	// Browser timeout settings
	NavigationTimeout time.Duration
	WaitTimeout       time.Duration

	// Search settings
	MaxSearchResults int
}

// DefaultConfig returns the default configuration for HelloAsso extractors
func DefaultConfig() HelloAssoConfig {
	return HelloAssoConfig{
		NavigationTimeout: 30 * time.Second,
		WaitTimeout:       15 * time.Second,
		MaxSearchResults:  10,
	}
}

// -----------------------------------------------------------------------------
// Activity Extraction Functions
// -----------------------------------------------------------------------------

// ExtractActivities extracts activities from the search results page
func ExtractActivities(page pw.Page, cfg ExtractionConfig) ([]Activity, error) {
	if page == nil {
		return nil, fmt.Errorf("page is nil")
	}

	// Check if we have an empty state
	emptyStateLocator := page.Locator(cfg.EmptyStateSelector)
	count, err := emptyStateLocator.Count()
	if err == nil && count > 0 {
		log.Printf("Empty state found - no results available")
		return []Activity{}, nil
	}

	// Find activities using the provided selector
	elements := page.Locator(cfg.ActivitySelector)
	count, err = elements.Count()
	if err != nil {
		return nil, fmt.Errorf("could not count activity elements: %w", err)
	}
	log.Printf("Found %d activity elements", count)

	activities := []Activity{}

	// For each activity element, extract information using explicit iteration
	for i := 0; i < count; i++ {
		// Skip any elements that are "show all" buttons or pagination
		if i == count-1 {
			// Check if this is the "show all" link that often appears as the last element
			element := elements.Nth(i)
			className, err := element.GetAttribute("class")
			if err == nil && (strings.Contains(className, "ShowAll") || strings.Contains(className, "Pagination")) {
				log.Printf("Skipping element %d: appears to be a 'Show all' or pagination button", i)
				continue
			}
		}

		activity := Activity{}
		element := elements.Nth(i)

		// Extract URL first (from parent 'a' tag)
		urlElement := element.Locator("a").First()
		urlCount, err := urlElement.Count()
		if err != nil || urlCount == 0 {
			log.Printf("Warning: Element %d has no URL, skipping", i)
			continue
		}

		href, err := urlElement.GetAttribute("href")
		if err != nil || href == "" {
			log.Printf("Warning: Element %d has empty href, skipping", i)
			continue
		}

		// Normalize URL
		if strings.HasPrefix(href, "/") {
			activity.URL = cfg.BaseURL + href
		} else {
			activity.URL = href
		}
		log.Printf("Element %d: Found URL: '%s'", i, activity.URL)

		// Extract title
		titleElement := element.Locator(cfg.Selectors.Title)
		titleCount, titleErr := titleElement.Count()
		if titleErr == nil && titleCount > 0 {
			title, err := titleElement.First().TextContent()
			if err == nil && title != "" {
				activity.Title = strings.TrimSpace(title)
				log.Printf("Element %d: Found title: '%s'", i, activity.Title)
			}
		}

		// Extract date
		dateElement := element.Locator(cfg.Selectors.Date)
		dateCount, dateErr := dateElement.Count()
		if dateErr == nil && dateCount > 0 {
			date, err := dateElement.First().TextContent()
			if err == nil && date != "" {
				activity.Date = strings.TrimSpace(date)
				log.Printf("Element %d: Found date: '%s'", i, activity.Date)
			}
		}

		// Extract price
		priceElement := element.Locator(cfg.Selectors.Price)
		priceCount, priceErr := priceElement.Count()
		if priceErr == nil && priceCount > 0 {
			price, err := priceElement.First().TextContent()
			if err == nil && price != "" {
				activity.Price = strings.TrimSpace(price)
				log.Printf("Element %d: Found price: '%s'", i, activity.Price)
			}
		}

		// Extract organization name
		orgElement := element.Locator(cfg.Selectors.Organization)
		orgCount, orgErr := orgElement.Count()
		if orgErr == nil && orgCount > 0 {
			org, err := orgElement.First().TextContent()
			if err == nil && org != "" {
				activity.Organization = strings.TrimSpace(org)
				log.Printf("Element %d: Found organization: '%s'", i, activity.Organization)
			}
		}

		// Extract category
		categoryElement := element.Locator(cfg.Selectors.Category)
		categoryCount, categoryErr := categoryElement.Count()
		if categoryErr == nil && categoryCount > 0 {
			category, err := categoryElement.First().TextContent()
			if err == nil && category != "" {
				activity.Category = strings.TrimSpace(category)
				log.Printf("Element %d: Found category: '%s'", i, activity.Category)
			}
		}

		// Extract location
		locationElement := element.Locator(cfg.Selectors.Location)
		locationCount, locationErr := locationElement.Count()
		if locationErr == nil && locationCount > 0 {
			location, err := locationElement.First().TextContent()
			if err == nil && location != "" {
				activity.Location = strings.TrimSpace(location)
				log.Printf("Element %d: Found location: '%s'", i, activity.Location)
			}
		}

		// Add activity if we have both title and URL
		if activity.Title != "" && activity.URL != "" {
			activities = append(activities, activity)
			log.Printf("Added activity: %s (Date: %s, URL: %s, Location: %s)",
				activity.Title, activity.Date, activity.URL, activity.Location)
		} else {
			log.Printf("Skipping activity - Missing required fields: Title=%t, URL=%t",
				activity.Title != "", activity.URL != "")
		}
	}

	log.Printf("Extracted %d activities from HelloAsso search", len(activities))
	return activities, nil
}

// Extract searches HelloAsso for tournament-related activities
func Extract(query string, browserContext pw.BrowserContext, config HelloAssoConfig) ([]string, error) {
	// Create a new page
	pageObj, err := browserContext.NewPage()
	if err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}
	defer pageObj.Close()

	// Create a page handler
	pageHandler := page.NewPageHandler(pageObj)

	// Set navigation timeout
	pageObj.SetDefaultNavigationTimeout(float64(config.NavigationTimeout / time.Millisecond))
	pageObj.SetDefaultTimeout(float64(config.WaitTimeout / time.Millisecond))

	// Search for the tournament
	encodedQuery := strings.ReplaceAll(query, " ", "+")
	searchURL := fmt.Sprintf("https://www.helloasso.com/search/?q=%s", encodedQuery)

	utils.DebugLog("Searching HelloAsso for: %s", query)
	if err := pageHandler.SafeNavigation(searchURL, 3, nil); err != nil {
		return nil, fmt.Errorf("failed to navigate to HelloAsso search page: %w", err)
	}

	// Wait for results to load
	time.Sleep(2 * time.Second)

	// Extract activity links using JavaScript
	script := `
	() => {
		const links = [];
		const resultElements = document.querySelectorAll('.search-group .search-card-item');
		
		for (const result of resultElements) {
			const linkElement = result.querySelector('a.card-link');
			if (linkElement && linkElement.href) {
				links.push(linkElement.href);
			}
		}
		
		return links;
	}
	`

	result, err := pageObj.Evaluate(script)
	if err != nil {
		return nil, fmt.Errorf("failed to extract activity links: %w", err)
	}

	// Process the results
	var activityLinks []string
	if resultArray, ok := result.([]interface{}); ok {
		for _, item := range resultArray {
			if link, ok := item.(string); ok {
				activityLinks = append(activityLinks, link)
			}
		}
	}

	// Limit the number of results
	if len(activityLinks) > config.MaxSearchResults {
		activityLinks = activityLinks[:config.MaxSearchResults]
	}

	utils.DebugLog("Found %d HelloAsso activities for query: %s", len(activityLinks), query)
	return activityLinks, nil
}
