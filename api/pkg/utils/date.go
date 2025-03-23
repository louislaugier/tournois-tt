package utils

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ParseTournamentDate attempts to parse tournament dates from various formats
func ParseTournamentDate(dateStr string) (time.Time, error) {
	// First try standard format
	date, err := time.Parse("2006-01-02", dateStr)
	if err == nil {
		return date, nil
	}

	// Try ISO 8601 format with time component (like "2024-06-08T00:00:00")
	date, err = time.Parse("2006-01-02T15:04:05", dateStr)
	if err == nil {
		return date, nil
	}

	// Try other common formats
	formats := []string{
		"02/01/2006",
		"2006/01/02",
		"01/02/2006",
		"Jan 2, 2006",
		"2 Jan 2006",
	}

	for _, format := range formats {
		date, err := time.Parse(format, dateStr)
		if err == nil {
			return date, nil
		}
	}

	return time.Time{}, fmt.Errorf("failed to parse date: %s", dateStr)
}

// ParseHelloAssoDate attempts to parse a date from HelloAsso
func ParseHelloAssoDate(dateStr string) (time.Time, error) {
	// Check for empty string
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("could not parse date: empty string")
	}

	// Clean up the date string
	dateStr = strings.TrimSpace(dateStr)
	lowerDateStr := strings.ToLower(dateStr)

	// First, handle date ranges like "Du 26/04/2025 au 27/04/2025"
	if strings.Contains(lowerDateStr, "du ") && strings.Contains(lowerDateStr, " au ") {
		// Extract the first date from the range
		parts := strings.Split(lowerDateStr, " au ")
		if len(parts) > 0 {
			firstDatePart := strings.TrimPrefix(parts[0], "du ")
			firstDatePart = strings.TrimSpace(firstDatePart)

			// Extract the first date if it's in DD/MM/YYYY format
			dateComponents := strings.Split(firstDatePart, "/")
			if len(dateComponents) == 3 {
				day := strings.TrimSpace(dateComponents[0])
				month := strings.TrimSpace(dateComponents[1])
				year := strings.TrimSpace(dateComponents[2])

				formattedDate := fmt.Sprintf("%s/%s/%s", day, month, year)
				if date, err := time.Parse("02/01/2006", formattedDate); err == nil {
					return date, nil
				}
			}
		}
	}

	// Handle single dates with "le" prefix (e.g., "le 26/04/2025")
	if strings.HasPrefix(lowerDateStr, "le ") {
		dateStr = strings.TrimPrefix(lowerDateStr, "le ")
		dateStr = strings.TrimSpace(dateStr)
	}

	// Handle dates with weekday prefixes (e.g., "Samedi 26/04/2025")
	frenchDays := []string{"lundi", "mardi", "mercredi", "jeudi", "vendredi", "samedi", "dimanche"}
	for _, day := range frenchDays {
		if strings.HasPrefix(lowerDateStr, day) {
			dateStr = strings.TrimPrefix(lowerDateStr, day)
			dateStr = strings.TrimSpace(dateStr)
			break
		}
	}

	// Try to extract date in DD/MM/YYYY format using regex
	re := regexp.MustCompile(`\b(\d{1,2})[/\-\.](\d{1,2})[/\-\.](\d{4})\b`)
	matches := re.FindStringSubmatch(dateStr)
	if len(matches) == 4 {
		day := matches[1]
		month := matches[2]
		year := matches[3]

		// Ensure day and month are two digits
		if len(day) == 1 {
			day = "0" + day
		}
		if len(month) == 1 {
			month = "0" + month
		}

		formattedDate := fmt.Sprintf("%s/%s/%s", day, month, year)
		if date, err := time.Parse("02/01/2006", formattedDate); err == nil {
			return date, nil
		}
	}

	// HelloAsso dates will likely be in French formats
	formats := []string{
		// Common French formats
		"02/01/2006",
		"02/01/2006 15:04",
		"02-01-2006",
		"02-01-2006 15:04",
		"2 January 2006", // For English months
		"2 January 2006 15:04",
		// French formats with month names
		"2 Janvier 2006",
		"2 Février 2006",
		"2 Mars 2006",
		"2 Avril 2006",
		"2 Mai 2006",
		"2 Juin 2006",
		"2 Juillet 2006",
		"2 Août 2006",
		"2 Septembre 2006",
		"2 Octobre 2006",
		"2 Novembre 2006",
		"2 Décembre 2006",
		// French formats with abbreviated month names
		"2 Jan 2006",
		"2 Fév 2006",
		"2 Mar 2006",
		"2 Avr 2006",
		"2 Mai 2006",
		"2 Juin 2006",
		"2 Juil 2006",
		"2 Août 2006",
		"2 Sept 2006",
		"2 Oct 2006",
		"2 Nov 2006",
		"2 Déc 2006",
	}

	// Try to parse with each format
	for _, format := range formats {
		if date, err := time.Parse(format, dateStr); err == nil {
			return date, nil
		}
	}

	// Try to normalize the date string
	dateStr = strings.ReplaceAll(dateStr, ",", "")

	// Handle French month names that might have different capitalizations or accents
	lowerDateStr = strings.ToLower(dateStr)
	if strings.Contains(lowerDateStr, "janvier") || strings.Contains(lowerDateStr, "jan") {
		dateStr = ReplaceMonthWithNumber(dateStr, "01")
	} else if strings.Contains(lowerDateStr, "février") || strings.Contains(lowerDateStr, "fév") || strings.Contains(lowerDateStr, "fev") {
		dateStr = ReplaceMonthWithNumber(dateStr, "02")
	} else if strings.Contains(lowerDateStr, "mars") || strings.Contains(lowerDateStr, "mar") {
		dateStr = ReplaceMonthWithNumber(dateStr, "03")
	} else if strings.Contains(lowerDateStr, "avril") || strings.Contains(lowerDateStr, "avr") {
		dateStr = ReplaceMonthWithNumber(dateStr, "04")
	} else if strings.Contains(lowerDateStr, "mai") {
		dateStr = ReplaceMonthWithNumber(dateStr, "05")
	} else if strings.Contains(lowerDateStr, "juin") || strings.Contains(lowerDateStr, "jun") {
		dateStr = ReplaceMonthWithNumber(dateStr, "06")
	} else if strings.Contains(lowerDateStr, "juillet") || strings.Contains(lowerDateStr, "juil") {
		dateStr = ReplaceMonthWithNumber(dateStr, "07")
	} else if strings.Contains(lowerDateStr, "août") || strings.Contains(lowerDateStr, "aout") || strings.Contains(lowerDateStr, "aoû") {
		dateStr = ReplaceMonthWithNumber(dateStr, "08")
	} else if strings.Contains(lowerDateStr, "septembre") || strings.Contains(lowerDateStr, "sept") {
		dateStr = ReplaceMonthWithNumber(dateStr, "09")
	} else if strings.Contains(lowerDateStr, "octobre") || strings.Contains(lowerDateStr, "oct") {
		dateStr = ReplaceMonthWithNumber(dateStr, "10")
	} else if strings.Contains(lowerDateStr, "novembre") || strings.Contains(lowerDateStr, "nov") {
		dateStr = ReplaceMonthWithNumber(dateStr, "11")
	} else if strings.Contains(lowerDateStr, "décembre") || strings.Contains(lowerDateStr, "dec") || strings.Contains(lowerDateStr, "déc") {
		dateStr = ReplaceMonthWithNumber(dateStr, "12")
	}

	// Try to parse in DD/MM/YYYY format
	fields := strings.Fields(dateStr)
	if len(fields) >= 3 {
		// Extract day, month (now numeric), and year
		day := ""
		month := ""
		year := ""

		for _, field := range fields {
			// Look for numeric parts
			if IsNumeric(field) {
				if len(field) == 4 { // Likely a year
					year = field
				} else if len(field) <= 2 { // Likely a day
					if day == "" {
						day = field
					} else if month == "" {
						month = field
					}
				}
			}
		}

		// If we have all parts, try to construct a date
		if day != "" && month != "" && year != "" {
			// Ensure day and month are two digits
			if len(day) == 1 {
				day = "0" + day
			}
			if len(month) == 1 {
				month = "0" + month
			}

			// Create date string in format DD/MM/YYYY
			formattedDate := fmt.Sprintf("%s/%s/%s", day, month, year)
			if date, err := time.Parse("02/01/2006", formattedDate); err == nil {
				return date, nil
			}
		}
	}

	return time.Time{}, fmt.Errorf("could not parse date: %s", dateStr)
}

// ReplaceMonthWithNumber replaces a month name with its numeric representation
func ReplaceMonthWithNumber(dateStr, monthNum string) string {
	// Split into fields
	fields := strings.Fields(dateStr)
	result := make([]string, 0, len(fields))

	for i, field := range fields {
		// Skip the month name, add the month number instead
		if i == 1 { // Month is typically the second field in "DD Month YYYY"
			result = append(result, monthNum)
		} else {
			result = append(result, field)
		}
	}

	return strings.Join(result, " ")
}

// IsNumeric checks if a string contains only numbers
func IsNumeric(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}

	for _, char := range s {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}

// IsDateCloseEnough checks if two dates are within a specific number of days of each other
func IsDateCloseEnough(date1, date2 time.Time, maxDaysDiff int) bool {
	diff := date1.Sub(date2)
	days := int(diff.Hours() / 24)
	if days < 0 {
		days = -days
	}
	return days <= maxDaysDiff
}

// GetMonthNameFrench returns the French name of the month for the given month number (1-12)
// or time.Month value
func GetMonthNameFrench(month interface{}) string {
	months := []string{
		"janvier", "février", "mars", "avril", "mai", "juin",
		"juillet", "août", "septembre", "octobre", "novembre", "décembre",
	}

	// Handle both int and time.Month types
	var monthNum int
	switch m := month.(type) {
	case int:
		monthNum = m
	case time.Month:
		monthNum = int(m)
	default:
		return ""
	}

	if monthNum < 1 || monthNum > 12 {
		return ""
	}

	return months[monthNum-1]
}
