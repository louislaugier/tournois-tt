package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
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

func extractTournamentID(id string) string {
	parts := strings.Split(id, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

type FFTTTournament struct {
	ID        int           `json:"id"`
	Name      string        `json:"name"`
	Type      string        `json:"type"`
	StartDate string        `json:"startDate"`
	EndDate   string        `json:"endDate"`
	Address   types.Address `json:"address"`
	Club      types.Club    `json:"club"`
	Rules     *types.Rules  `json:"rules"`
	Status    int           `json:"status"`
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
	var ffttTournaments []FFTTTournament
	if err := json.Unmarshal(body, &ffttTournaments); err != nil {
		log.Printf("Error parsing tournaments data: %v", err)
		log.Printf("Raw response body: %s", string(body))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse tournaments data"})
		return
	}

	// Convert to our internal type
	tournaments := make([]types.Tournament, len(ffttTournaments))
	for i, t := range ffttTournaments {
		tournaments[i] = types.Tournament{
			ID:        t.ID,
			Name:      t.Name,
			Type:      t.Type,
			StartDate: t.StartDate,
			EndDate:   t.EndDate,
			Address:   t.Address,
			Club:      t.Club,
			Rules:     t.Rules,
			Status:    t.Status,
		}
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
