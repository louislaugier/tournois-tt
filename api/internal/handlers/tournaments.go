package handlers

import (
	"log"
	"net/http"
	"strings"
	"time"
	"tournois-tt/api/internal/crons/tournaments/signup/refresh"
	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/fftt"
	"tournois-tt/api/pkg/geocoding"
	"tournois-tt/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

// TournamentResponse represents the data to return to API clients
type TournamentResponse struct {
	ID        int               `json:"id"`
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	StartDate string            `json:"startDate"`
	EndDate   string            `json:"endDate"`
	Address   geocoding.Address `json:"address"`
	Club      fftt.Club         `json:"club"`
	Rules     *fftt.Rules       `json:"rules,omitempty"`
	SignupURL string            `json:"signupUrl,omitempty"`
	Endowment int               `json:"endowment"`
}

// TournamentsHandler handles tournament requests by retrieving data from the cache
func TournamentsHandler(c *gin.Context) {
	cachedTournaments, err := cache.LoadTournaments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load tournaments from cache"})
		return
	}

	// Convert to response format with only needed fields
	var tournamentsResponse []TournamentResponse
	for _, cachedTournament := range cachedTournaments {
		tournamentsResponse = append(tournamentsResponse, TournamentResponse{
			ID:        cachedTournament.ID,
			Name:      cachedTournament.Name,
			Type:      cachedTournament.Type,
			StartDate: cachedTournament.StartDate,
			EndDate:   cachedTournament.EndDate,
			Address: geocoding.Address{
				StreetAddress:             cachedTournament.Address.StreetAddress,
				PostalCode:                cachedTournament.Address.PostalCode,
				AddressLocality:           cachedTournament.Address.AddressLocality,
				DisambiguatingDescription: cachedTournament.Address.DisambiguatingDescription,
				Latitude:                  cachedTournament.Address.Latitude,
				Longitude:                 cachedTournament.Address.Longitude,
				Failed:                    cachedTournament.Address.Failed,
			},
			Club: fftt.Club{
				ID:         cachedTournament.Club.ID,
				Name:       cachedTournament.Club.Name,
				Code:       cachedTournament.Club.Code,
				Department: cachedTournament.Club.Department,
				Region:     cachedTournament.Club.Region,
				Identifier: cachedTournament.Club.Identifier,
			},
			SignupURL: cachedTournament.SignupUrl,
			Endowment: cachedTournament.Endowment,
		})

		// Add rules if available
		if cachedTournament.Rules != nil {
			tournamentsResponse[len(tournamentsResponse)-1].Rules = &fftt.Rules{
				AgeMin:  cachedTournament.Rules.AgeMin,
				AgeMax:  cachedTournament.Rules.AgeMax,
				Points:  cachedTournament.Rules.Points,
				Ranking: cachedTournament.Rules.Ranking,
				URL:     cachedTournament.Rules.URL,
			}
		}
	}

	// Filter by postal code if provided
	postalCode := c.Query("postalCode")
	if postalCode != "" {
		var filteredTournaments []TournamentResponse
		for _, t := range tournamentsResponse {
			if strings.HasPrefix(t.Address.PostalCode, postalCode) {
				filteredTournaments = append(filteredTournaments, t)
			}
		}
		tournamentsResponse = filteredTournaments
	}

	// Log response data
	log.Printf("Returned %d tournaments (filtered by postal code: %s)",
		len(tournamentsResponse), postalCode)

	for i, t := range tournamentsResponse {
		date, err := utils.ParseTournamentDate(t.StartDate)
		if err != nil {
			log.Printf("Failed to parse tournament date: %v", err)
			continue
		}

		// check also if date is after now
		if refresh.IsCurrentSeasonQuery(&date, nil) && date.After(time.Now()) {
			tournamentsResponse[i].Address.PostalCode += "test1"
			break
		}
	}

	c.JSON(http.StatusOK, tournamentsResponse)
}
