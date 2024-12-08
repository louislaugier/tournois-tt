package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
	"tournois-tt/api/internal/geocoding"
	"tournois-tt/api/pkg/fftt"

	"github.com/gin-gonic/gin"
)

type Tournament struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	StartDate string  `json:"startDate"`
	EndDate   string  `json:"endDate"`
	Address   Address `json:"address"`
	Club      Club    `json:"club"`
}

type Address struct {
	PostalCode      string  `json:"postalCode"`
	StreetAddress   string  `json:"streetAddress"`
	AddressLocality string  `json:"addressLocality"`
	AddressCountry  *string `json:"addressCountry"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
}

type Club struct {
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
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
	var tournaments []Tournament
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
		addr := geocoding.Address{
			StreetAddress:   tournaments[i].Address.StreetAddress,
			PostalCode:      tournaments[i].Address.PostalCode,
			AddressLocality: tournaments[i].Address.AddressLocality,
			AddressCountry:  tournaments[i].Address.AddressCountry,
		}

		coords, err := geocoding.GetCoordinates(addr)
		if err != nil {
			log.Printf("Warning: Failed to get coordinates for tournament %s: %v", tournaments[i].Name, err)
			continue
		}

		tournaments[i].Address.Latitude = coords.Lat
		tournaments[i].Address.Longitude = coords.Lon
	}

	elapsed := time.Since(start)
	log.Printf("Request processing completed in %v", elapsed)
	c.JSON(http.StatusOK, tournaments)
}
