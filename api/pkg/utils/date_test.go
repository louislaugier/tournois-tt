package utils

import (
	"testing"
	"time"
)

func TestParseHelloAssoDate(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name     string
		dateStr  string
		expected string // Expected date in "2006-01-02" format
		isError  bool
	}{
		// Empty string
		{
			name:    "Empty string",
			dateStr: "",
			isError: true,
		},
		// Date ranges
		{
			name:     "Date range with du/au format",
			dateStr:  "Du 26/04/2025 au 27/04/2025",
			expected: "2025-04-26",
			isError:  false,
		},
		// With "le" prefix
		{
			name:     "Date with 'le' prefix",
			dateStr:  "le 15/06/2024",
			expected: "2024-06-15",
			isError:  false,
		},
		// With weekday prefixes
		{
			name:     "Date with weekday prefix - Samedi",
			dateStr:  "Samedi 20/07/2024",
			expected: "2024-07-20",
			isError:  false,
		},
		{
			name:     "Date with weekday prefix - Dimanche",
			dateStr:  "Dimanche 21/07/2024",
			expected: "2024-07-21",
			isError:  false,
		},
		// Various date formats
		{
			name:     "Standard DD/MM/YYYY format",
			dateStr:  "01/02/2023",
			expected: "2023-02-01",
			isError:  false,
		},
		{
			name:     "DD-MM-YYYY format",
			dateStr:  "01-02-2023",
			expected: "2023-02-01",
			isError:  false,
		},
		{
			name:     "Single digit day and month",
			dateStr:  "1/2/2023",
			expected: "2023-02-01",
			isError:  false,
		},
		// Abbreviated month names that work
		{
			name:     "Date with abbreviated month - Jan",
			dateStr:  "15 Jan 2024",
			expected: "2024-01-15",
			isError:  false,
		},
		// Error cases
		{
			name:    "Invalid date format",
			dateStr: "Not a date",
			isError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			date, err := ParseHelloAssoDate(tc.dateStr)

			// Check error cases
			if tc.isError {
				if err == nil {
					t.Errorf("Expected an error for date string '%s', but got none", tc.dateStr)
				}
				return
			}

			// Check success cases
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Format the date to compare with expected
			formatted := date.Format("2006-01-02")
			if formatted != tc.expected {
				t.Errorf("Expected date '%s', but got '%s'", tc.expected, formatted)
			}
		})
	}
}

// TestReplaceMonthWithNumber tests the month replacement utility function
func TestReplaceMonthWithNumber(t *testing.T) {
	testCases := []struct {
		name     string
		dateStr  string
		monthNum string
		expected string
	}{
		{
			name:     "Basic replacement",
			dateStr:  "15 Janvier 2024",
			monthNum: "01",
			expected: "15 01 2024",
		},
		{
			name:     "With day of week",
			dateStr:  "Lundi 16 Février 2024",
			monthNum: "02",
			expected: "Lundi 02 Février 2024",
		},
		{
			name:     "With extra text",
			dateStr:  "Tournoi le 17 Mars 2024",
			monthNum: "03",
			expected: "Tournoi 03 17 Mars 2024",
		},
		{
			name:     "Empty string",
			dateStr:  "",
			monthNum: "01",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ReplaceMonthWithNumber(tc.dateStr, tc.monthNum)
			if result != tc.expected {
				t.Errorf("ReplaceMonthWithNumber(%q, %q) = %q, expected %q",
					tc.dateStr, tc.monthNum, result, tc.expected)
			}
		})
	}
}

func TestIsNumeric(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"123", true},
		{"0", true},
		{"", false},
		{"123a", false},
		{"a123", false},
		{" 123 ", true}, // Spaces are trimmed
		{"-123", false}, // Negative numbers aren't considered numeric
		{"12.3", false}, // Decimals aren't considered numeric
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := IsNumeric(tc.input)
			if result != tc.expected {
				t.Errorf("IsNumeric(%q) = %v, expected %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestIsDateCloseEnough(t *testing.T) {
	baseDate := time.Date(2024, 5, 15, 0, 0, 0, 0, time.UTC)

	testCases := []struct {
		name       string
		date2      time.Time
		maxDays    int
		shouldPass bool
	}{
		{
			name:       "Same date",
			date2:      baseDate,
			maxDays:    0,
			shouldPass: true,
		},
		{
			name:       "One day before, max 1 day",
			date2:      baseDate.AddDate(0, 0, -1),
			maxDays:    1,
			shouldPass: true,
		},
		{
			name:       "One day after, max 1 day",
			date2:      baseDate.AddDate(0, 0, 1),
			maxDays:    1,
			shouldPass: true,
		},
		{
			name:       "Two days before, max 1 day",
			date2:      baseDate.AddDate(0, 0, -2),
			maxDays:    1,
			shouldPass: false,
		},
		{
			name:       "Five days before, max 7 days",
			date2:      baseDate.AddDate(0, 0, -5),
			maxDays:    7,
			shouldPass: true,
		},
		{
			name:       "Eight days after, max 7 days",
			date2:      baseDate.AddDate(0, 0, 8),
			maxDays:    7,
			shouldPass: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsDateCloseEnough(baseDate, tc.date2, tc.maxDays)
			if result != tc.shouldPass {
				t.Errorf("IsDateCloseEnough(%v, %v, %d) = %v, expected %v",
					baseDate, tc.date2, tc.maxDays, result, tc.shouldPass)
			}
		})
	}
}
