package helloasso

import (
	"fmt"
	"log"
	"strings"
	"tournois-tt/api/pkg/scraper/page"

	pw "github.com/playwright-community/playwright-go"
)

const (
	BaseURL           = "https://www.helloasso.com"
	SearchURLTemplate = "https://www.helloasso.com/e/recherche?query=%s"
)

// Selectors contains all the CSS selectors used for HelloAsso scraping
var Selectors = ActivitySelectors{
	Title:        ".Thumbnail--Name",
	Date:         ".Thumbnail--Date",
	URL:          "a",
	Price:        ".Thumbnail--ImagePill",
	Organization: ".Thumbnail--OrganizationName",
	Category:     ".Thumbnail--MetadataTag",
	Location:     ".Thumbnail--MetadataLocation",
}

// Config contains all the configuration for HelloAsso scraping
var Config = page.Config{
	EmptyStateSelector: `[data-testid="empty-state"]`,
	ResultsSelector:    ".Hits-Activity",
}

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

// ExtractActivities extracts activities from the search results page
func ExtractActivities(page pw.Page, cfg ExtractionConfig) ([]Activity, error) {
	// Check if we have an empty state
	emptyState, err := page.QuerySelector(cfg.EmptyStateSelector)
	if err == nil && emptyState != nil {
		log.Printf("Empty state found - no results available")
		return []Activity{}, nil
	}

	// Find activities using the provided selector
	elements := page.Locator(cfg.ActivitySelector)
	count, err := elements.Count()
	if err != nil {
		return nil, fmt.Errorf("could not count activity elements: %v", err)
	}
	log.Printf("Found %d activity elements", count)

	activities := []Activity{}

	for i := 0; i < count; i++ {
		element := elements.Nth(i)
		activity := Activity{}

		// Extract title
		if title, err := element.Locator(cfg.Selectors.Title).TextContent(); err == nil && title != "" {
			activity.Title = strings.TrimSpace(title)
		}

		// Extract date
		if date, err := element.Locator(cfg.Selectors.Date).TextContent(); err == nil && date != "" {
			activity.Date = strings.TrimSpace(date)
		}

		// Extract URL
		if href, err := element.Locator(cfg.Selectors.URL).GetAttribute("href"); err == nil && href != "" {
			if strings.HasPrefix(href, "/") {
				activity.URL = cfg.BaseURL + href
			} else {
				activity.URL = href
			}
		}

		// Extract price
		if price, err := element.Locator(cfg.Selectors.Price).TextContent(); err == nil && price != "" {
			activity.Price = strings.TrimSpace(price)
		}

		// Extract organization name
		if org, err := element.Locator(cfg.Selectors.Organization).TextContent(); err == nil && org != "" {
			activity.Organization = strings.TrimSpace(org)
		}

		// Extract category
		if category, err := element.Locator(cfg.Selectors.Category).TextContent(); err == nil && category != "" {
			activity.Category = strings.TrimSpace(category)
		}

		// Extract location
		if location, err := element.Locator(cfg.Selectors.Location).TextContent(); err == nil && location != "" {
			activity.Location = strings.TrimSpace(location)
		}

		if activity.Title != "" { // Only add activities that at least have a title
			activities = append(activities, activity)
			log.Printf("Found activity: %s at %s", activity.Title, activity.Location)
		}
	}

	return activities, nil
}
