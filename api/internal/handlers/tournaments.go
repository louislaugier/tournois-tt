package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
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
			go func(addr geocoding.Address) {
				// Attempt to geocode if not in cache
				var coords geocoding.GeocodeResult
				var err error

				// First try Nominatim
				coords, err = geocoding.GetCoordinatesNominatim(addr)
				if err != nil || coords.Failed {
					// If Nominatim fails, try Google
					coords, err = geocoding.GetCoordinatesGoogle(addr)
					if err != nil || coords.Failed {
						return
					}
				}

				// Cache the result
				geocoding.SetCachedGeocodeResult(coords)
			}(t.Address)
		}
	}

	c.JSON(http.StatusOK, tournaments)
}
