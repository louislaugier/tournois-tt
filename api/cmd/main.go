package main

import (
	"log"
	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/pdf"
)

func main() {
	tournament := findTournament3026()
	if tournament == nil {
		return
	}

	// _, _, browserContext, _ := browser.Setup()
	// defer browser.ShutdownBrowser()
	// tournamentDate, _ := time.Parse("2006-01-02T15:04:05", tournament.StartDate)

	content, err := pdf.ExtractFileContentFromURL(tournament.Rules.URL)
	if err != nil {
		return
	}
	log.Printf("content: %s", content)

	// activities, err := helloasso.SearchActivities(context.Background(), "tournoi tennis de table courbevoie")
	// if err != nil {
	// 	log.Println(err)
	// }
	// log.Println(activities)

	// test.LogClubEmailAddresses()
	// test.LogCommitteeAndLeagueEmailAddresses()

	////////////////////////////////////////////////////////

	// crons.Schedule()

	// // Run geocoding refresh in a background goroutine
	// go func() {
	// 	tournaments.RefreshTournamentsAndGeocoding()
	// 	tournaments.RefreshSignupURLs()
	// }()

	// r := router.NewRouter()

	// log.Printf("Server starting...")
	// if err := r.Run(":8080"); err != nil {
	// 	log.Fatalf("Error starting server: %v", err)
	// }
}

// Function to find tournament with ID 3026 and log its details
func findTournament3026() *cache.TournamentCache {
	log.Println("Loading tournaments from cache...")

	// Load tournaments from cache
	tournaments, err := cache.LoadTournaments()
	if err != nil {
		log.Fatalf("Failed to load tournaments: %v", err)
	}

	log.Printf("Loaded %d tournaments from cache", len(tournaments))

	// Find tournament with ID 3026
	targetID := "3026"
	targetTournament, found := tournaments[targetID]

	// If not found by string key, try iterating through all tournaments
	if !found {
		log.Printf("Tournament with ID 3026 not found by direct key lookup, searching through all tournaments...")
		for key, tournament := range tournaments {
			if tournament.ID == 3026 {
				targetTournament = tournament
				found = true
				log.Printf("Found tournament with ID 3026 under key: %s", key)
				break
			}
		}
	}

	// Log the tournament if found
	if found {
		log.Printf("Found tournament with ID 3026:")
		log.Printf("Name: %s", targetTournament.Name)
		log.Printf("Club: %s", targetTournament.Club.Name)
		log.Printf("Start Date: %s", targetTournament.StartDate)
		log.Printf("End Date: %s", targetTournament.EndDate)
		log.Printf("Site URL: %s", targetTournament.SiteUrl)
		log.Printf("Signup URL: %s", targetTournament.SignupUrl)

		// Log rules if available
		if targetTournament.Rules != nil {
			log.Printf("Rules URL: %s", targetTournament.Rules.URL)
		}

		return &targetTournament
	}
	log.Printf("Tournament with ID 3026 not found in cache")
	return nil
}
