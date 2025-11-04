package scripts

import (
	"fmt"
	"log"

	igimage "tournois-tt/api/pkg/image"
)

// test_instagram_image.go
// This script generates a sample tournament image for visual verification
// Usage: go run scripts/test_instagram_image.go

func main() {
	fmt.Println("🏓 Generating sample tournament image...")

	// Create sample tournament data
	tournament := igimage.TournamentImage{
		Name:          "Tournoi National de Paris 2025",
		Type:          "Tournoi jeunes",
		Club:          "Paris Tennis de Table Club",
		Endowment:     1500,
		StartDate:     "2025-11-15",
		EndDate:       "2025-11-16",
		Address:       "Gymnase Jean Jaurès, 123 Rue du Sport, 75001 Paris",
		RulesURL:      "https://example.com/rules.pdf",
		TournamentID:  12345,
		TournamentURL: "https://tournois-tt.fr/12345",
	}

	// Generate image
	imagePath, err := igimage.GenerateTournamentImage(tournament)
	if err != nil {
		log.Fatalf("❌ Failed to generate image: %v", err)
	}

	fmt.Printf("✅ Image generated successfully!\n")
	fmt.Printf("📁 File location: %s\n", imagePath)
	fmt.Printf("\n")
	fmt.Printf("You can now:\n")
	fmt.Printf("1. Open the image to verify the design\n")
	fmt.Printf("2. Check that brand colors are correct:\n")
	fmt.Printf("   - Header: #1FBAD6 (cyan)\n")
	fmt.Printf("   - Accent: #9B59B6 (purple)\n")
	fmt.Printf("   - Background: #242730 (dark)\n")
	fmt.Printf("3. Verify tournament details are properly formatted\n")
	fmt.Printf("\n")
	fmt.Printf("Note: The image will remain in the temp directory for inspection.\n")
	fmt.Printf("Delete it manually when done: rm %s\n", imagePath)
}
