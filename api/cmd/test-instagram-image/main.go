package main

import (
    "flag"
    "fmt"
    "log"
    "math/rand"
    "strings"

    "tournois-tt/api/pkg/cache"
    igimage "tournois-tt/api/pkg/image"
)

func main() {
	storyOnly := flag.Bool("story-only", false, "Generate only story image (1080x1920)")
	flag.Parse()

	log.Println("🏓 Testing Instagram image generation...")

	// Load tournaments
	tournaments, err := cache.LoadTournaments()
	if err != nil {
		log.Fatalf("Failed to load tournaments: %v", err)
	}

	if len(tournaments) == 0 {
		log.Fatal("No tournaments found in cache")
	}

	// Pick a random tournament
	var tournamentID string
	var tournament cache.TournamentCache
	for id, t := range tournaments {
		tournamentID = id
		tournament = t
		break // Take first one
	}

	// Use a different random one if we want variety
	keys := make([]string, 0, len(tournaments))
	for k := range tournaments {
		keys = append(keys, k)
	}
	if len(keys) > 0 {
		randomIndex := rand.Intn(len(keys))
		tournamentID = keys[randomIndex]
		tournament = tournaments[tournamentID]
	}

	log.Printf("Selected tournament: %s - %s", tournamentID, tournament.Name)
	log.Printf("Type: %s", tournament.Type)
	log.Printf("Club: %+v", tournament.Club)
	log.Printf("Start: %s", tournament.StartDate)
	log.Printf("End: %s", tournament.EndDate)

    // Build full address for LIEU
    addressParts := []string{}
    if tournament.Address.DisambiguatingDescription != "" {
        addressParts = append(addressParts, tournament.Address.DisambiguatingDescription)
    }
    if tournament.Address.StreetAddress != "" {
        addressParts = append(addressParts, tournament.Address.StreetAddress)
    }
    locality := tournament.Address.AddressLocality
    if tournament.Address.PostalCode != "" && locality != "" {
        addressParts = append(addressParts, fmt.Sprintf("%s %s", tournament.Address.PostalCode, locality))
    } else if locality != "" {
        addressParts = append(addressParts, locality)
    }
    address := "Adresse non disponible"
    if len(addressParts) > 0 {
        address = strings.Join(addressParts, ", ")
    }

    clubName := tournament.Club.Name
    if tournament.Club.Identifier != "" {
        clubName = fmt.Sprintf("%s (%s)", tournament.Club.Name, tournament.Club.Identifier)
    }

    rulesURL := ""
    if tournament.Rules != nil && tournament.Rules.URL != "" {
        rulesURL = tournament.Rules.URL
    }

    tournamentImage := igimage.TournamentImage{
		Name:          tournament.Name,
		Type:          tournament.Type,
        Club:          clubName,
		Endowment:     tournament.Endowment,
		StartDate:     tournament.StartDate,
		EndDate:       tournament.EndDate,
		Address:       address,
		RulesURL:      rulesURL,
		Page:          tournament.Page,
		TournamentID:  tournament.ID,
		TournamentURL: fmt.Sprintf("https://tournois-tt.fr/%d", tournament.ID),
	}

	var feedPath, storyPath string
	
	if *storyOnly {
		// Generate ONLY story image (1080x1920)
		log.Println("📸 Generating Instagram story image (1080x1920)...")
		storyPath, err = igimage.GenerateTournamentStoryImage(tournamentImage)
		if err != nil {
			log.Fatalf("Failed to generate story image: %v", err)
		}

		log.Printf("✅ Story image generated successfully!")
		log.Printf("📁 Story Path: %s", storyPath)
		log.Println("")
		log.Println("You can view the story image at:")
		log.Printf("  %s", storyPath)
	} else {
		// Generate feed image (1080x1080)
		log.Println("📸 Generating Instagram feed image (1080x1080)...")
		feedPath, err = igimage.GenerateTournamentImage(tournamentImage)
		if err != nil {
			log.Fatalf("Failed to generate feed image: %v", err)
		}

		log.Printf("✅ Feed image generated successfully!")
		log.Printf("📁 Feed Path: %s", feedPath)
		log.Println("")
		
		// Generate story image (1080x1920)
		log.Println("📸 Generating Instagram story image (1080x1920)...")
		storyPath, err = igimage.GenerateTournamentStoryImage(tournamentImage)
		if err != nil {
			log.Fatalf("Failed to generate story image: %v", err)
		}

		log.Printf("✅ Story image generated successfully!")
		log.Printf("📁 Story Path: %s", storyPath)
		log.Println("")
		log.Println("You can view the images at:")
		log.Printf("  Feed:  %s", feedPath)
		log.Printf("  Story: %s", storyPath)
	}
}

