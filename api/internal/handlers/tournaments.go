package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
	"tournois-tt/api/pkg/fftt"
	"tournois-tt/api/pkg/geocoding"
	"tournois-tt/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

// Modify TournamentsHandler to use geocoding cache
func TournamentsHandler(c *gin.Context) {
	// Get all query parameters
	queryParams := c.Request.URL.Query()

	// Call FFTT API
	resp, err := fftt.GetClient().GetTournaments(queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data from FFTT API"})
		return
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"error": "FFTT API returned an error"})
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body"})
		return
	}

	// Debug log first tournament raw data
	var rawTournaments []map[string]interface{}
	json.Unmarshal(body, &rawTournaments)

	if len(rawTournaments) > 0 {
		json.MarshalIndent(rawTournaments[0], "", "  ")
	}

	var ffttTournaments []fftt.Tournament
	if err := json.Unmarshal(body, &ffttTournaments); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse tournaments data"})
		return
	}

	// Convert to our internal type
	tournaments := make([]fftt.Tournament, len(ffttTournaments))

	// Create a collection for newly geocoded results
	newResults := make([]geocoding.GeocodeResult, 0)

	for i, t := range ffttTournaments {
		tournament := t
		// Map the tournament type to full form
		tournament.Type = utils.MapTournamentType(t.Type)

		// Append a dot to the postal code
		tournament.Address.PostalCode = tournament.Address.PostalCode + "\u200e"

		// Append 23:59 to endDate if it doesn't already have a time
		if !strings.Contains(tournament.EndDate, ":") {
			tournament.EndDate = tournament.EndDate + " 23:59"
		}

		// Check if address is in geocoding cache
		if cachedResult, exists := geocoding.GetCachedGeocodeResult(t.Address); exists && !cachedResult.Failed {
			// Use cached coordinates or failed status
			tournament.Address.Latitude = cachedResult.Latitude
			tournament.Address.Longitude = cachedResult.Longitude

			// Skip further geocoding if already processed
			tournaments[i] = tournament
		} else {
			// Use the geocoding package's GetCoordinates function that tries both Nominatim and Google
			location, err := geocoding.GetCoordinates(t.Address)

			// Create a geocode result
			result := geocoding.GeocodeResult{
				Address:   t.Address,
				Timestamp: time.Now(),
			}

			if err != nil || location.Failed {
				result.Failed = true
			} else {
				// Update the tournament with coordinates
				tournament.Address.Latitude = location.Lat
				tournament.Address.Longitude = location.Lon

				// Set geocode result values
				result.Latitude = location.Lat
				result.Longitude = location.Lon
				result.Failed = false

				// Log successful geocoding for debugging
				log.Printf("Geocoded new address: %s -> (%f, %f)",
					geocoding.ConstructFullAddress(t.Address), location.Lat, location.Lon)
			}

			// Cache the result in memory
			geocoding.SetCachedGeocodeResult(result)

			// Add to collection of newly geocoded results
			newResults = append(newResults, result)

			tournaments[i] = tournament
		}
	}

	// Only save if we have new results to persist
	if len(newResults) > 0 {
		log.Printf("Saving %d newly geocoded results to cache", len(newResults))
		if err := geocoding.SaveGeocodeResultsToCache(newResults); err != nil {
			log.Printf("Warning: Failed to save geocoding cache: %v", err)
		}
	}

	c.JSON(http.StatusOK, tournaments)
}
