package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/instagram"
)

// E2E Test for Instagram & Threads Posting Integration
// This posts a REAL Instagram post with a random or specified tournament
//
// Usage:
//   go run cmd/test-instagram-e2e/main.go
//
// Test specific tournament:
//   TEST_TOURNAMENT_ID=3348 go run cmd/test-instagram-e2e/main.go
//
// Required environment variables:
//   INSTAGRAM_ENABLED=true
//   INSTAGRAM_ACCESS_TOKEN=your_token
//   INSTAGRAM_PAGE_ID=your_page_id
//   THREADS_ENABLED=true (optional)
//   THREADS_ACCESS_TOKEN=your_token (if Threads enabled)
//   THREADS_USER_ID=your_id (if Threads enabled)

func main() {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  Instagram & Threads Post E2E Test")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Load configuration
	config := instagram.Config{
		AccessToken:        os.Getenv("INSTAGRAM_ACCESS_TOKEN"),
		PageID:             os.Getenv("INSTAGRAM_PAGE_ID"),
		ThreadsAccessToken: os.Getenv("THREADS_ACCESS_TOKEN"),
		ThreadsUserID:      os.Getenv("THREADS_USER_ID"),
		Enabled:            os.Getenv("INSTAGRAM_ENABLED") == "true",
		ThreadsEnabled:     os.Getenv("THREADS_ENABLED") == "true",
	}

	// Validate Instagram configuration
	if !config.Enabled {
		log.Fatal("❌ INSTAGRAM_ENABLED must be set to 'true'")
	}
	if config.AccessToken == "" {
		log.Fatal("❌ INSTAGRAM_ACCESS_TOKEN is required")
	}
	if config.PageID == "" {
		log.Fatal("❌ INSTAGRAM_PAGE_ID is required")
	}

	fmt.Println("✅ Configuration loaded")
	fmt.Println()
	fmt.Printf("  • Instagram Page ID: %s\n", config.PageID)
	fmt.Printf("  • Instagram Token: %s...%s (length: %d)\n",
		config.AccessToken[:min(10, len(config.AccessToken))],
		config.AccessToken[max(0, len(config.AccessToken)-10):],
		len(config.AccessToken))
	if config.ThreadsEnabled {
		fmt.Printf("  • Threads User ID: %s\n", config.ThreadsUserID)
		fmt.Printf("  • Threads enabled: ✅\n")
	} else {
		fmt.Printf("  • Threads enabled: ❌\n")
	}
	fmt.Println()

	// Create Instagram client
	client := instagram.NewClient(config)

	// Load tournaments
	fmt.Println("📂 Loading tournaments from data.json...")
	tournaments, err := loadTournaments()
	if err != nil {
		log.Fatalf("❌ Failed to load tournaments: %v", err)
	}

	if len(tournaments) == 0 {
		log.Fatal("❌ No tournaments found in data.json")
	}

	fmt.Printf("✅ Loaded %d tournaments from data.json\n", len(tournaments))
	fmt.Println()

	// Pick tournament - either specific ID or random
	var selectedTournament cache.TournamentCache
	var randomIndex int

	if testID := os.Getenv("TEST_TOURNAMENT_ID"); testID != "" {
		// Use specific tournament for testing
		tournamentID := 0
		fmt.Sscanf(testID, "%d", &tournamentID)

		for _, t := range tournaments {
			if t.ID == tournamentID {
				selectedTournament = t
				break
			}
		}

		if selectedTournament.ID == 0 {
			log.Fatalf("❌ Tournament ID %s not found in data.json", testID)
		}

		fmt.Printf("🎯 Using specified tournament: %s (ID: %d)\n",
			selectedTournament.Name, selectedTournament.ID)
		fmt.Println()
	} else {
		// Pick random tournament
		randomIndex = rand.Intn(len(tournaments))
		selectedTournament = tournaments[randomIndex]

		fmt.Printf("🎲 Randomly selected tournament #%d: %s (ID: %d)\n",
			randomIndex+1, selectedTournament.Name, selectedTournament.ID)
		fmt.Println()
	}

	// Convert to TournamentImage
	tournamentImage := convertToImage(selectedTournament)

	fmt.Println("📋 Tournament details:")
	fmt.Printf("  • Name: %s\n", tournamentImage.Name)
	fmt.Printf("  • Type: %s\n", tournamentImage.Type)
	fmt.Printf("  • Club: %s\n", tournamentImage.Club)
	fmt.Printf("  • Endowment: %d €\n", tournamentImage.Endowment)
	fmt.Printf("  • Dates: %s to %s\n", tournamentImage.StartDate, tournamentImage.EndDate)
	fmt.Printf("  • Address: %s\n", tournamentImage.Address)
	fmt.Printf("  • URL: %s\n", tournamentImage.TournamentURL)
	fmt.Println()

	// Generate image (saved to instagram-images folder)
	fmt.Println("🖼️  Generating tournament image...")
	imagePath, err := instagram.GenerateTournamentImage(tournamentImage)
	if err != nil {
		log.Fatalf("❌ Failed to generate image: %v", err)
	}
	// Don't cleanup - we want to keep the image in instagram-images folder

	fileInfo, err := os.Stat(imagePath)
	if err != nil {
		log.Fatalf("❌ Failed to stat image: %v", err)
	}

	fmt.Printf("✅ Image generated: %s\n", imagePath)
	fmt.Printf("   Size: %d bytes (%.2f KB)\n", fileInfo.Size(), float64(fileInfo.Size())/1024)
	fmt.Println()

	// Final confirmation
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("⚠️  ABOUT TO POST TO INSTAGRAM & THREADS")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("This will create REAL posts on Instagram (feed + story)")
	if config.ThreadsEnabled {
		fmt.Println("and Threads.")
	} else {
		fmt.Println("(Threads is disabled).")
	}
	fmt.Println()

	fmt.Println()
	fmt.Println("📸 Posting to Instagram & Threads...")

	// Post to Instagram & Threads
	notification, err := client.PostTournament(tournamentImage)
	if err != nil {
		log.Fatalf("❌ Failed to post: %v", err)
	}

	if !notification.Success {
		log.Fatalf("❌ Post failed: %s", notification.Error)
	}

	// Success!
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("🎉 SUCCESS! Posted to Instagram & Threads!")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Printf("  • Instagram Post ID: %s\n", notification.MessageID)
	fmt.Printf("  • Posted at: %s\n", notification.SentAt.Format(time.RFC3339))
	fmt.Printf("  • Tournament: %s (ID: %d)\n", tournamentImage.Name, tournamentImage.TournamentID)
	fmt.Println()
	fmt.Println("✅ Check your Instagram feed & story and Threads for the post")
	fmt.Println()
}

// loadTournaments loads tournaments from data.json
func loadTournaments() ([]cache.TournamentCache, error) {
	// Find data.json
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
		// Try to find from working directory
		wd, _ := os.Getwd()
		for wd != "/" && wd != "." {
			testPath := filepath.Join(wd, "api", "cache", "data.json")
			if _, err := os.Stat(testPath); err == nil {
				dataPath = testPath
				break
			}
			// Also try without api prefix
			testPath = filepath.Join(wd, "cache", "data.json")
			if _, err := os.Stat(testPath); err == nil {
				dataPath = testPath
				break
			}
			wd = filepath.Dir(wd)
		}
	}

	if dataPath == "" {
		return nil, fmt.Errorf("could not find data.json")
	}

	// Read file
	data, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read data.json: %w", err)
	}

	// Parse JSON
	var tournaments []cache.TournamentCache
	if err := json.Unmarshal(data, &tournaments); err != nil {
		return nil, fmt.Errorf("failed to parse data.json: %w", err)
	}

	return tournaments, nil
}

// convertToImage converts TournamentCache to TournamentImage
func convertToImage(t cache.TournamentCache) instagram.TournamentImage {
	// Format address
	address := formatAddress(t.Address)

	// Build rules URL
	rulesURL := ""
	if t.Rules != nil && t.Rules.URL != "" {
		rulesURL = t.Rules.URL
	}

	// Build tournament URL
	tournamentURL := fmt.Sprintf("https://tournois-tt.fr/%d", t.ID)

	// Format club name with identifier
	clubName := t.Club.Name
	if t.Club.Identifier != "" {
		clubName = fmt.Sprintf("%s (%s)", t.Club.Name, t.Club.Identifier)
	}

	return instagram.TournamentImage{
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

// formatAddress formats address for display
func formatAddress(addr cache.Address) string {
	parts := []string{}

	if addr.DisambiguatingDescription != "" {
		parts = append(parts, addr.DisambiguatingDescription)
	}

	if addr.StreetAddress != "" {
		parts = append(parts, addr.StreetAddress)
	}

	if addr.PostalCode != "" && addr.AddressLocality != "" {
		parts = append(parts, fmt.Sprintf("%s %s", addr.PostalCode, addr.AddressLocality))
	} else if addr.AddressLocality != "" {
		parts = append(parts, addr.AddressLocality)
	}

	if len(parts) > 0 {
		return strings.Join(parts, ", ")
	}

	return "Adresse non disponible"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
