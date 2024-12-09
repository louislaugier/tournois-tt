package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
	"tournois-tt/api/internal/geocoding"
	"tournois-tt/api/internal/types"
	"tournois-tt/api/pkg/fftt"

	"github.com/gin-gonic/gin"
)

// mapTournamentType converts single letter type to full name
func mapTournamentType(t string) string {
	switch t {
	case "I":
		return "International"
	case "A":
		return "National A"
	case "B":
		return "National B"
	case "R":
		return "Régional"
	case "D":
		return "Départemental"
	case "P":
		return "Promotionnel"
	default:
		return t
	}
}

func TournamentsHandler(c *gin.Context) {
	start := time.Now()
	log.Printf("Starting tournament request processing")

	// Get all query parameters
	queryParams := c.Request.URL.Query()
	log.Printf("Query params: %v", queryParams)

	// Call FFTT API
	log.Printf("Calling FFTT API")
	resp, err := fftt.GetClient().GetTournaments(queryParams)
	if err != nil {
		log.Printf("Error fetching from FFTT API: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data from FFTT API"})
		return
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		log.Printf("FFTT API returned non-200 status: %d", resp.StatusCode)
		c.JSON(resp.StatusCode, gin.H{"error": "FFTT API returned an error"})
		return
	}

	// Read and parse the response body
	log.Printf("Reading FFTT API response body")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body"})
		return
	}

	log.Printf("Parsing tournaments data")
	var tournaments []types.Tournament
	if err := json.Unmarshal(body, &tournaments); err != nil {
		log.Printf("Error parsing tournaments data: %v", err)
		log.Printf("Raw response body: %s", string(body))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse tournaments data"})
		return
	}
	log.Printf("Found %d tournaments", len(tournaments))

	// Add coordinates from cache or geocode new addresses
	log.Printf("Adding coordinates to tournaments")
	for i := range tournaments {
		coords, err := geocoding.GetCoordinates(tournaments[i].Address)
		if err != nil {
			log.Printf("Warning: Failed to get coordinates for tournament %s: %v", tournaments[i].Name, err)
			continue
		}

		tournaments[i].Address.Latitude = coords.Lat
		tournaments[i].Address.Longitude = coords.Lon
		tournaments[i].Address.Approximate = coords.Approximate
		tournaments[i].Type = mapTournamentType(tournaments[i].Type)
	}

	elapsed := time.Since(start)
	log.Printf("Request processing completed in %v", elapsed)
	c.JSON(http.StatusOK, tournaments)
}
