package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"tournois-tt/api/pkg/fftt"
	"tournois-tt/api/pkg/geocoding"

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

// Add a function to load geocoding cache
func loadGeocodeCache() (map[string]geocoding.GeocodeResult, error) {
	cacheFilePath := filepath.Join("cache", "geocoding_cache.json")

	// Check if cache file exists
	if _, err := os.Stat(cacheFilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("geocoding cache file not found")
	}

	// Read cache file
	data, err := os.ReadFile(cacheFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read geocoding cache: %v", err)
	}

	var cachedResults []geocoding.GeocodeResult
	if err := json.Unmarshal(data, &cachedResults); err != nil {
		return nil, fmt.Errorf("failed to parse geocoding cache: %v", err)
	}

	cacheMap := make(map[string]geocoding.GeocodeResult)
	for _, result := range cachedResults {
		key := generateCacheKey(result.Address)
		cacheMap[key] = result
	}

	return cacheMap, nil
}

// generateCacheKey creates a unique key for an address
func generateCacheKey(addr geocoding.Address) string {
	return fmt.Sprintf("%s|%s|%s",
		strings.TrimSpace(addr.StreetAddress),
		strings.TrimSpace(addr.PostalCode),
		strings.TrimSpace(addr.AddressLocality))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Modify TournamentsHandler to use geocoding cache
func TournamentsHandler(c *gin.Context) {
	start := time.Now()
	log.Printf("Starting tournament request processing")

	// Load geocoding cache
	geocodingCache, err := loadGeocodeCache()
	if err != nil {
		log.Printf("Failed to load geocoding cache: %v", err)
		log.Printf("Initializing empty geocoding cache")
		geocodingCache = make(map[string]geocoding.GeocodeResult)
	} else {
		log.Printf("Successfully loaded %d cached geocoding results", len(geocodingCache))
	}

	// Get all query parameters
	queryParams := c.Request.URL.Query()
	log.Printf("Query params: %v", queryParams)

	// Log specific postcode parameter if present
	if postcode := queryParams.Get("address.postalCode"); postcode != "" {
		log.Printf("Filtering by postcode: %s", postcode)
	}

	// Call FFTT API
	log.Printf("Calling FFTT API with URL-encoded params: %s", queryParams.Encode())
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

	// Debug log for raw response
	log.Printf("Raw response sample (first 500 chars): %s", string(body)[:min(len(string(body)), 500)])

	// Debug log first tournament raw data
	var rawTournaments []map[string]interface{}
	if err := json.Unmarshal(body, &rawTournaments); err != nil {
		log.Printf("Error parsing raw tournaments: %v", err)
	} else if len(rawTournaments) > 0 {
		rawJSON, _ := json.MarshalIndent(rawTournaments[0], "", "  ")
		log.Printf("First tournament raw data: %s", string(rawJSON))
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
	log.Printf("Processing %d tournaments", len(ffttTournaments))

	for i, t := range ffttTournaments {
		tournaments[i] = t

		// Log postcode for each tournament if we're filtering by postcode
		if queryParams.Get("address.postalCode") != "" {
			log.Printf("Tournament %s postcode: %s", t.Name, t.Address.PostalCode)
		}

		// Map the tournament type to full form
		tournaments[i].Type = mapTournamentType(t.Type)

		// Append a dot to the postal code
		tournaments[i].Address.PostalCode = tournaments[i].Address.PostalCode + "\u200e"

		// Append 23:59 to endDate if it doesn't already have a time
		if !strings.Contains(tournaments[i].EndDate, ":") {
			tournaments[i].EndDate = tournaments[i].EndDate + " 23:59"
		}

		// Debug logging for rules URL
		if tournaments[i].Rules != nil {
			log.Printf("Tournament %s - Rules URL: %s", t.Name, tournaments[i].Rules.URL)
		}

		// Check if address is in geocoding cache
		cacheKey := generateCacheKey(t.Address)
		if cachedResult, exists := geocodingCache[cacheKey]; exists {
			log.Printf("Using cached geocoding result for tournament: %s", t.Name)

			// Use cached coordinates or failed status
			tournaments[i].Address.Latitude = cachedResult.Latitude
			tournaments[i].Address.Longitude = cachedResult.Longitude
			tournaments[i].Address.Failed = cachedResult.Failed

			// Skip further geocoding if already processed
			continue
		} else {
			// Attempt to geocode if not in cache
			coords, err := geocoding.GetCoordinates(t.Address)
			if err != nil {
				log.Printf("Warning: Failed to get coordinates for tournament %s: %v", t.Name, err)
				continue
			}

			if !coords.Failed {
				tournaments[i].Address.Latitude = coords.Lat
				tournaments[i].Address.Longitude = coords.Lon
			} else {
				tournaments[i].Address.Failed = true
			}
		}
	}

	elapsed := time.Since(start)
	log.Printf("Request processing completed in %v", elapsed)
	c.JSON(http.StatusOK, tournaments)
}
