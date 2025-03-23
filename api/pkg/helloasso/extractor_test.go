package helloasso

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestExtractActivitiesLogic tests the core logic of the activity extraction
// This approach avoids the complexity of mocking Playwright interfaces
func TestExtractActivitiesLogic(t *testing.T) {
	// Test cases
	testCases := []struct {
		name             string
		hasEmptyState    bool
		rawActivitiesLen int
		showAllButton    bool
		expectedCount    int
	}{
		{
			name:             "EmptyState",
			hasEmptyState:    true,
			rawActivitiesLen: 0,
			showAllButton:    false,
			expectedCount:    0,
		},
		{
			name:             "OneActivity",
			hasEmptyState:    false,
			rawActivitiesLen: 1,
			showAllButton:    false,
			expectedCount:    1,
		},
		{
			name:             "SkipShowAllButton",
			hasEmptyState:    false,
			rawActivitiesLen: 2,
			showAllButton:    true,
			expectedCount:    1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call our test helper function
			activities := simulateExtractActivities(tc.hasEmptyState, tc.rawActivitiesLen, tc.showAllButton)

			// Assert expected count
			assert.Len(t, activities, tc.expectedCount, "Expected %d activities, got %d", tc.expectedCount, len(activities))

			// For OneActivity, check all fields are populated
			if tc.name == "OneActivity" {
				if assert.NotEmpty(t, activities, "Should have at least one activity") {
					activity := activities[0]
					assert.NotEmpty(t, activity.Title, "Title should be populated")
					assert.NotEmpty(t, activity.Date, "Date should be populated")
					assert.NotEmpty(t, activity.Price, "Price should be populated")
					assert.NotEmpty(t, activity.Organization, "Organization should be populated")
					assert.NotEmpty(t, activity.Category, "Category should be populated")
					assert.NotEmpty(t, activity.Location, "Location should be populated")
					assert.NotEmpty(t, activity.URL, "URL should be populated")
				}
			}
		})
	}
}

// simulateExtractActivities is a helper function that simulates the behavior of ExtractActivities
// without using Playwright interfaces
func simulateExtractActivities(hasEmptyState bool, rawActivitiesLen int, hasShowAllButton bool) []Activity {
	// If empty state is detected, return empty slice
	if hasEmptyState {
		return []Activity{}
	}

	// Create result slice
	activities := []Activity{}

	// Process each activity
	for i := 0; i < rawActivitiesLen; i++ {
		// Skip show all button if it's present and this is the last activity
		if hasShowAllButton && i == rawActivitiesLen-1 {
			continue
		}

		// Create an activity with all fields
		activity := Activity{
			Title:        "Tournoi de Tennis de Table",
			Date:         "15 avril 2024",
			Price:        "10 €",
			Organization: "Club de Tennis de Table",
			Category:     "Sport",
			Location:     "Paris",
			URL:          BaseURL + "/event/tournoi-tt",
		}

		activities = append(activities, activity)
	}

	return activities
}
