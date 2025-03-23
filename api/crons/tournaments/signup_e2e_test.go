package tournaments_test

import (
	"fmt"
	"testing"
	"time"

	"tournois-tt/api/crons/tournaments/signup"
	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/scraper/browser"
	"tournois-tt/api/pkg/scraper/services/pdf_processor"
)

// TestRefreshSignupURLs tests PDF URL extraction for tournament signup
func TestRefreshSignupURLs(t *testing.T) {
	t.Parallel()
	signup.Debug = true

	testCases := []struct {
		name                   string
		PDFRulesURL            string
		expectedFinalSignupURL string
	}{
		{
			name:                   "CCTT 2025",
			PDFRulesURL:            "https://apiv2.fftt.com/api/files/357572/reglement%20tournoi%202025%20fftt%20(2).pdf",
			expectedFinalSignupURL: "https://tournoi.cctt.fr/sign-up/",
		},
		// {
		// 	name:                   "PPCF 2025",
		// 	PDFRulesURL:            "https://apiv2.fftt.com/api/files/390629/reglement_tournoi_2025.pdf",
		// 	expectedFinalSignupURL: "https://www.ppcflines.fr/tournoi/inscription-etape1",
		// },
		// {
		// 	name:                   "Open Catalan",
		// 	PDFRulesURL:            "https://apiv2.fftt.com/api/files/413718/Reglement%20OPEN%20CATALAN%202025%20V2.pdf",
		// 	expectedFinalSignupURL: "https://p018sukd.forms.app/formulaire-de-inscription-au-tournoi",
		// },
	}

	// Initialize browser
	cfg := browser.DefaultConfig()
	cfg.Headless = true
	browserInstance, pwInstance, err := browser.Init(cfg)
	if err != nil {
		t.Fatalf("Failed to initialize browser: %v", err)
	}
	defer pwInstance.Stop()
	defer browserInstance.Close()

	// Create browser context
	browserContext, err := browser.NewContext(browserInstance, cfg)
	if err != nil {
		t.Fatalf("Failed to create browser context: %v", err)
	}
	defer browserContext.Close()

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create simple tournament object with just the necessary information
			tournament := cache.TournamentCache{
				Name:      "Tournoi " + tc.name,
				StartDate: time.Now().Add(24 * time.Hour).Format("2006-01-02"),
				// Use a club name derived from the test case name
				Club: cache.Club{
					Name: tc.name,
				},
			}
			tournamentDate, _ := time.Parse("2006-01-02", tournament.StartDate)

			// Simple logging
			fmt.Printf("Testing %s: %s\n", tc.name, tc.PDFRulesURL)

			// Extract URL
			extractedURL, err := pdf_processor.ExtractSignupURL(
				tournament,
				tournamentDate,
				tc.PDFRulesURL,
				browserContext,
				signup.ValidateSignupURL,
			)

			// Log extraction result
			if err != nil {
				t.Logf("Error extracting URL: %v", err)
			}

			fmt.Printf("Found URL: %s\n", extractedURL)
			fmt.Printf("Expected URL: %s\n", tc.expectedFinalSignupURL)

			// Validate the result
			if extractedURL == "" {
				t.Errorf("No URL found for %s", tc.name)
				return
			}

			if extractedURL == tc.expectedFinalSignupURL {
				fmt.Printf("SUCCESS: URLs match exactly\n")
			} else {
				t.Errorf("URL %s does not match expected URL %s", extractedURL, tc.expectedFinalSignupURL)
			}
		})
	}
}
