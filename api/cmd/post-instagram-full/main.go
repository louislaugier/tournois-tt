package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"tournois-tt/api/internal/config"
	"tournois-tt/api/pkg/cache"
	igimage "tournois-tt/api/pkg/image"
	instagramapi "tournois-tt/api/pkg/instagram/api"
)

func main() {
	idFlag := flag.Int("id", 0, "Tournament ID to post to Instagram (feed + story + threads)")
	flag.Parse()

	if *idFlag == 0 {
		log.Fatal("Please provide tournament ID via --id flag")
	}

	log.Printf("🏓 Posting tournament %d to Instagram (feed + story + threads)...\n", *idFlag)

	// Load tournaments
	tournaments, err := loadTournaments()
	if err != nil {
		log.Fatalf("Failed to load tournaments: %v", err)
	}

	// Find tournament
	var tournament cache.TournamentCache
	found := false
	for _, t := range tournaments {
		if t.ID == *idFlag {
			tournament = t
			found = true
			break
		}
	}

	if !found {
		log.Fatalf("Tournament %d not found in cache", *idFlag)
	}

	log.Printf("📋 Tournament: %s", tournament.Name)
	log.Printf("   Type: %s", tournament.Type)
	log.Printf("   Club: %s", tournament.Club.Name)
	log.Println()

	// Convert to image data
	tournamentImage := convertToImage(tournament)

	// Create Instagram client
	instagramConfig := instagramapi.Config{
		AccessToken:        config.InstagramAccessToken,
		PageID:             config.InstagramPageID,
		ThreadsAccessToken: config.ThreadsAccessToken,
		ThreadsUserID:      config.ThreadsUserID,
		Enabled:            config.InstagramEnabled,
		ThreadsEnabled:     config.ThreadsEnabled,
	}

	client := instagramapi.NewClient(instagramConfig)

	// Test connection
	if err := client.TestConnection(); err != nil {
		log.Fatalf("❌ Instagram API connection failed: %v", err)
	}
	log.Println("✅ Instagram API connection successful")
	log.Println()

	// Confirm
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("⚠️  ABOUT TO POST TO INSTAGRAM (FEED + STORY + THREADS)")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("This will post to:")
	fmt.Println("  ✓ Instagram Feed (1080x1080)")
	fmt.Println("  ✓ Instagram Story (1080x1920)")
	if config.ThreadsEnabled {
		fmt.Println("  ✓ Threads")
	}
	fmt.Println()
	fmt.Printf("Tournament: %s\n", tournament.Name)
	fmt.Println()
	fmt.Print("Continue? (yes/no): ")
	
	var response string
	fmt.Scanln(&response)
	
	if response != "yes" {
		log.Println("❌ Cancelled by user")
		return
	}

	// Post to Instagram (feed + story + threads)
	log.Println("📸 Posting to Instagram...")
	notification, err := client.PostTournament(tournamentImage)
	if err != nil {
		log.Fatalf("❌ Failed to post: %v", err)
	}

	if !notification.Success {
		log.Fatalf("❌ Posting failed: %s", notification.Error)
	}

	log.Println()
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("✅ SUCCESS - Posted to Instagram!")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println()
	log.Printf("Post ID: %s", notification.MessageID)
	log.Printf("Tournament: %s (ID: %d)", tournament.Name, tournament.ID)
}

func loadTournaments() ([]cache.TournamentCache, error) {
	possiblePaths := []string{
		"./api/cache/data.json",
		"./cache/data.json",
		"../cache/data.json",
		"../../cache/data.json",
	}

	var dataPath string
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			dataPath = path
			break
		}
	}

	if dataPath == "" {
		wd, _ := os.Getwd()
		current := wd
		for current != "/" && current != "." {
			candidate := filepath.Join(current, "api", "cache", "data.json")
			if _, err := os.Stat(candidate); err == nil {
				dataPath = candidate
				break
			}

			candidate = filepath.Join(current, "cache", "data.json")
			if _, err := os.Stat(candidate); err == nil {
				dataPath = candidate
				break
			}

			current = filepath.Dir(current)
		}
	}

	if dataPath == "" {
		return nil, fmt.Errorf("data.json not found")
	}

	payload, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, err
	}

	var tournaments []cache.TournamentCache
	if err := json.Unmarshal(payload, &tournaments); err != nil {
		return nil, err
	}

	return tournaments, nil
}

func convertToImage(t cache.TournamentCache) igimage.TournamentImage {
	address := formatAddress(t.Address)

	rulesURL := ""
	if t.Rules != nil && t.Rules.URL != "" {
		rulesURL = t.Rules.URL
	}

	tournamentURL := fmt.Sprintf("https://tournois-tt.fr/%d", t.ID)

	clubName := t.Club.Name
	if t.Club.Identifier != "" {
		clubName = fmt.Sprintf("%s (%s)", t.Club.Name, t.Club.Identifier)
	}

	return igimage.TournamentImage{
		Name:          t.Name,
		Type:          t.Type,
		Club:          clubName,
		Endowment:     t.Endowment,
		StartDate:     t.StartDate,
		EndDate:       t.EndDate,
		Address:       address,
		RulesURL:      rulesURL,
		TournamentID:  t.ID,
		TournamentURL: tournamentURL,
	}
}

func formatAddress(addr cache.Address) string {
	parts := []string{}

	if addr.DisambiguatingDescription != "" {
		parts = append(parts, addr.DisambiguatingDescription)
	}

	if addr.StreetAddress != "" {
		parts = append(parts, addr.StreetAddress)
	}

	locality := addr.AddressLocality
	if addr.PostalCode != "" && locality != "" {
		parts = append(parts, fmt.Sprintf("%s %s", addr.PostalCode, locality))
	} else if locality != "" {
		parts = append(parts, locality)
	}

	if len(parts) == 0 {
		return "Adresse non disponible"
	}

	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += ", " + parts[i]
	}
	return result
}

