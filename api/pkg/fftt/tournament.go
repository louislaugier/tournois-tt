package fftt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

// Debug flag to control verbose logging
var Debug = false

// debugLog logs a message only if Debug is true
func debugLog(format string, args ...interface{}) {
	if Debug {
		log.Printf(format, args...)
	}
}

// truncateString truncates a string to maxLen and adds "..." if it was truncated
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// FetchTournaments fetches tournaments from the FFTT API with the given query parameters
func FetchTournaments(queryParams url.Values) ([]Tournament, error) {
	// Ensure the FFTTClient is initialized if it's nil
	if FFTTClient == nil {
		GetClient() // This will initialize FFTTClient
	}

	// Use the client to make the request
	resp, err := FFTTClient.GetTournaments(queryParams)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tournaments: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("FFTT API returned status %d", resp.StatusCode)
	}

	// Read the response body for inspection
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Log the full response body if in debug mode
	debugLog("FFTT API response: %s", string(bodyBytes))

	// Check if the response is an array
	trimmedBody := bytes.TrimSpace(bodyBytes)
	if len(trimmedBody) == 0 {
		return []Tournament{}, nil // Empty response, return empty slice
	}

	// First try to parse it as a direct array of tournaments
	if trimmedBody[0] == '[' {
		var tournaments []Tournament
		if err := json.Unmarshal(trimmedBody, &tournaments); err != nil {
			return nil, fmt.Errorf("failed to decode tournaments array: %v", err)
		}
		return tournaments, nil
	}

	// If it's not an array, it could be a Hydra Collection format or an error
	var responseObj map[string]json.RawMessage
	if err := json.Unmarshal(trimmedBody, &responseObj); err != nil {
		// If we can't parse it as a JSON object, return a limited part of the response
		truncatedResponse := truncateString(string(trimmedBody), 200)
		return nil, fmt.Errorf("FFTT API returned an invalid response format (truncated): %s", truncatedResponse)
	}

	// Check if it's a Hydra Collection with "hydra:member" array
	if memberBytes, exists := responseObj["hydra:member"]; exists {
		var tournaments []Tournament
		if err := json.Unmarshal(memberBytes, &tournaments); err != nil {
			return nil, fmt.Errorf("failed to decode tournaments from hydra:member: %v", err)
		}
		return tournaments, nil
	}

	// Check for error information in the response
	if description, exists := responseObj["hydra:description"]; exists {
		var errorMsg string
		if err := json.Unmarshal(description, &errorMsg); err == nil {
			return nil, fmt.Errorf("FFTT API returned an error: %s", errorMsg)
		}
	}

	if title, exists := responseObj["hydra:title"]; exists {
		var titleMsg string
		if err := json.Unmarshal(title, &titleMsg); err == nil {
			return nil, fmt.Errorf("FFTT API returned an error: %s", titleMsg)
		}
	}

	// If we can't identify the response format, return a limited part of the object
	truncatedResponse := truncateString(string(trimmedBody), 200)
	return nil, fmt.Errorf("FFTT API returned an unexpected object response (truncated): %s", truncatedResponse)
}

// GetFutureTournaments fetches and returns tournaments that start after the given date
func GetFutureTournaments(startDateAfter time.Time, startDateBefore *time.Time) ([]Tournament, error) {
	// Create query params for future tournaments
	queryParams := url.Values{}
	queryParams.Set("startDate[after]", startDateAfter.Format("2006-01-02T15:04:05"))
	if startDateBefore != nil {
		queryParams.Set("startDate[before]", startDateBefore.Format("2006-01-02T15:04:05"))
	}
	queryParams.Set("itemsPerPage", "999999")
	queryParams.Set("order[startDate]", "asc")

	return FetchTournaments(queryParams)
}
