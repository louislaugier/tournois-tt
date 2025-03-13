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

// Club represents a table tennis club
type Club struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Code       string `json:"code"`
	Department string `json:"department"`
	Region     string `json:"region"`
	Identifier string `json:"identifier"`
}

// Rules represents tournament rules
type Rules struct {
	AgeMin  int    `json:"ageMin"`
	AgeMax  int    `json:"ageMax"`
	Points  int    `json:"points"`
	Ranking int    `json:"ranking"`
	URL     string `json:"url,omitempty"`
}

// Table represents a tournament table
type Table struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Date        string `json:"date"`
	Time        string `json:"time"`
	Fee         int    `json:"fee"`
	Endowment   int    `json:"endowment"`
}

// Organization represents tournament organization details
type Organization struct {
	Name    string `json:"name"`
	Contact string `json:"contact"`
}

// Response represents tournament responses
type Response struct {
	PlayerID  int    `json:"playerId"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// Tournament represents a complete tournament
type Tournament struct {
	ID           int               `json:"id"`
	Name         string            `json:"name"`
	Type         string            `json:"type"`
	StartDate    string            `json:"startDate"`
	EndDate      string            `json:"endDate"`
	Address      geocoding.Address `json:"address"`
	Club         Club              `json:"club"`
	Rules        *Rules            `json:"rules"`
	Tables       []Table           `json:"tables"`
	Status       int               `json:"status"`
	Endowment    int               `json:"endowment"`
	Organization *Organization     `json:"organization,omitempty"`
	Responses    []Response        `json:"responses,omitempty"`
}

// Modify TournamentsHandler to use geocoding cache
func TournamentsHandler(c *gin.Context) {
	// Load geocoding cache
	geocodingCache, err := geocoding.LoadGeocodeCache()
	if err != nil {
		geocodingCache = make(map[string]geocoding.GeocodeResult)
	}

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

	var ffttTournaments []Tournament
	if err := json.Unmarshal(body, &ffttTournaments); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse tournaments data"})
		return
	}

	// Convert to our internal type
	tournaments := make([]Tournament, len(ffttTournaments))

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
		cacheKey := geocoding.GenerateCacheKey(t.Address)
		if cachedResult, exists := geocodingCache[cacheKey]; exists && !cachedResult.Failed {
			// Use cached coordinates or failed status
			tournament.Address.Latitude = cachedResult.Latitude
			tournament.Address.Longitude = cachedResult.Longitude

			// Skip further geocoding if already processed
			tournaments[i] = tournament
		} else {
			go func() {
				// Attempt to geocode if not in cache
				var coords geocoding.GeocodeResult
				var err error

				// First try Nominatim
				coords, err = geocoding.GetCoordinatesNominatim(t.Address)
				if err != nil || coords.Failed {
					// If Nominatim fails, try Google
					coords, err = geocoding.GetCoordinatesGoogle(t.Address)
					if err != nil || coords.Failed {
						return
					}
				}

				// Cache the result
				cacheKey := geocoding.GenerateCacheKey(t.Address)
				geocodingCache[cacheKey] = coords
			}()
		}
	}

	c.JSON(http.StatusOK, tournaments)
}
