package refresh

import (
	"fmt"
	"log"
	"time"

	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/fftt"
	"tournois-tt/api/pkg/utils"

	pw "github.com/playwright-community/playwright-go"
)

// ProcessTournament processes a tournament by updating its Page field from FFTT API
// This function is used by the Worker function in worker.go
func ProcessTournament(workerID int, tournament cache.TournamentCache,
	browserContext pw.BrowserContext, pwInstance *pw.Playwright) (cache.TournamentCache, bool, error) {

	// Check and update Page field from FFTT API if missing
	if tournament.Page == "" {
		updatedTournament, pageUpdated, err := updatePageFromFFTT(workerID, tournament)
		if err != nil {
			log.Printf("Worker %d: Error updating Page field for tournament %s: %v", workerID, tournament.Name, err)
			return tournament, false, err
		}
		if pageUpdated {
			log.Printf("Worker %d: Updated Page field for tournament %s: %s", workerID, tournament.Name, updatedTournament.Page)
			return updatedTournament, true, nil
		}
	}

	return tournament, false, nil
}

// updatePageFromFFTT fetches the latest tournament data from FFTT API and updates the Page field
func updatePageFromFFTT(workerID int, tournament cache.TournamentCache) (cache.TournamentCache, bool, error) {
	// Parse the tournament date to set up a narrow query range
	tournamentDate, err := utils.ParseTournamentDate(tournament.StartDate)
	if err != nil {
		return tournament, false, fmt.Errorf("failed to parse tournament date: %w", err)
	}

	// Create a narrow date range (same day)
	startOfDay := time.Date(tournamentDate.Year(), tournamentDate.Month(), tournamentDate.Day(), 0, 0, 0, 0, tournamentDate.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	utils.DebugLog("Worker %d: Fetching tournament %d from FFTT API to update Page field", workerID, tournament.ID)

	// Fetch tournaments from FFTT API for this specific date
	tournaments, err := fftt.GetFutureTournaments(startOfDay, &endOfDay)
	if err != nil {
		return tournament, false, fmt.Errorf("failed to fetch tournament from FFTT API: %w", err)
	}

	// Find the matching tournament by ID
	for _, t := range tournaments {
		if t.ID == tournament.ID {
			// Found the tournament - update Page field if it exists
			if t.Page != "" && t.Page != tournament.Page {
				tournament.Page = t.Page
				return tournament, true, nil
			}
			// Tournament found but Page is still empty
			return tournament, false, nil
		}
	}

	// Tournament not found in FFTT API
	utils.DebugLog("Worker %d: Tournament %d not found in FFTT API", workerID, tournament.ID)
	return tournament, false, nil
}
