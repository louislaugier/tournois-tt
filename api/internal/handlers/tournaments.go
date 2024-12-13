package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
	"tournois-tt/api/pkg/geocoding"
	"tournois-tt/api/pkg/fftt"

	"github.com/gin-gonic/gin"
)

// Club represents a table tennis club
type Club struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Code       string `json:"code"`
	Department string `json:"department"`
	Region     string `json:"region"`
}

// Rules represents tournament rules
type Rules struct {
	AgeMin  int `json:"ageMin"`
	AgeMax  int `json:"ageMax"`
	Points  int `json:"points"`
	Ranking int `json:"ranking"`
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

func formatTableInfo(tables []Table) string {
	if len(tables) == 0 {
		return ""
	}

	var result strings.Builder
	for _, t := range tables {
		fee := float64(t.Fee) / 100.0
		endowment := float64(t.Endowment) / 100.0
		result.WriteString(fmt.Sprintf("%s - %s (%s %s) : %.2f€ / %.2f€\n",
			t.Name, t.Description, t.Date[0:10], t.Time, fee, endowment))
	}
	return result.String()
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
	var ffttTournaments []Tournament
	if err := json.Unmarshal(body, &ffttTournaments); err != nil {
		log.Printf("Error parsing tournaments data: %v", err)
		log.Printf("Raw response body: %s", string(body))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse tournaments data"})
		return
	}

	// Convert to our internal type
	tournaments := make([]Tournament, len(ffttTournaments))
	for i, t := range ffttTournaments {
		log.Printf("Tournament %d organization: %+v", t.ID, t.Organization)
		log.Printf("Tournament %d responses: %+v", t.ID, t.Responses)
		tournaments[i] = t
	}

	log.Printf("Found %d tournaments", len(tournaments))

	var skippedGeocoding, failedGeocoding int
	var tournamentsWithCoordinates []Tournament

	// Add coordinates from cache or geocode new addresses
	log.Printf("Adding coordinates to tournaments")
	for i := range tournaments {
		addr := tournaments[i].Address

		// Log detailed address information
		log.Printf("Attempting to geocode tournament: %s", tournaments[i].Name)
		log.Printf("  Street Address: %q", addr.StreetAddress)
		log.Printf("  Postal Code: %q", addr.PostalCode)
		log.Printf("  Locality: %q", addr.AddressLocality)

		// Attempt to get coordinates
		coords, err := geocoding.GetCoordinates(tournaments[i].Address)
		if err != nil {
			log.Printf("Warning: Failed to get coordinates for tournament %s: %v", tournaments[i].Name, err)
			log.Printf("  Detailed address: %+v", tournaments[i].Address)
			failedGeocoding++
			continue
		}

		tournaments[i].Address.Latitude = coords.Lat
		tournaments[i].Address.Longitude = coords.Lon
		tournaments[i].Type = mapTournamentType(tournaments[i].Type)

		tournamentsWithCoordinates = append(tournamentsWithCoordinates, tournaments[i])
	}

	log.Printf("Geocoding summary:")
	log.Printf("  Total tournaments: %d", len(tournaments))
	log.Printf("  Skipped due to previous geocoding failures: %d", skippedGeocoding)
	log.Printf("  Failed geocoding: %d", failedGeocoding)
	log.Printf("  Tournaments with coordinates: %d", len(tournamentsWithCoordinates))

	elapsed := time.Since(start)
	log.Printf("Request processing completed in %v", elapsed)
	c.JSON(http.StatusOK, tournaments)
}
