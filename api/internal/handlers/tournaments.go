package handlers

import (
	"log"
	"net/http"
	"strings"
	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/fftt"
	"tournois-tt/api/pkg/geocoding"
	"tournois-tt/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

// TournamentsHandler handles tournament requests by retrieving data from the cache
func TournamentsHandler(c *gin.Context) {
	// Load tournaments from cache
	cachedTournaments, err := cache.LoadTournaments()
	if err != nil {
		log.Printf("Error loading tournament cache: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load tournaments from cache"})
		return
	}

	// Convert cache map to slice of tournaments
	tournaments := make([]fftt.Tournament, 0, len(cachedTournaments))

	for _, cachedTournament := range cachedTournaments {
		tournament := fftt.Tournament{
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
			Endowment:              cachedTournament.Endowment,
			IsRulesPdfChecked:      cachedTournament.IsRulesPdfChecked,
			IsSiteExistenceChecked: cachedTournament.IsSiteExistenceChecked,
			SiteURL:                cachedTournament.SiteUrl,
			SignupURL:              cachedTournament.SignupUrl,
		}

		// Map the tournament type to full form
		tournament.Type = utils.MapTournamentType(tournament.Type)

		// Append a dot to the postal code
		tournament.Address.PostalCode = tournament.Address.PostalCode + "\u200e"

		// Append 23:59 to endDate if it doesn't already have a time
		if !strings.Contains(tournament.EndDate, ":") {
			tournament.EndDate = tournament.EndDate + " 23:59"
		}

		// Add Rules if available
		if cachedTournament.Rules != nil {
			tournament.Rules = &fftt.Rules{
				AgeMin:  cachedTournament.Rules.AgeMin,
				AgeMax:  cachedTournament.Rules.AgeMax,
				Points:  cachedTournament.Rules.Points,
				Ranking: cachedTournament.Rules.Ranking,
				URL:     cachedTournament.Rules.URL,
			}
		}

		tournaments = append(tournaments, tournament)
	}

	c.JSON(http.StatusOK, tournaments)
}
