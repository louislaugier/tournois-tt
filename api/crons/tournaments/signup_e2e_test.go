package tournaments_test

import (
	"fmt"
	"testing"
	"time"

	"tournois-tt/api/crons/tournaments/signup"
	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/pdf/extraction"
	"tournois-tt/api/pkg/scraper/browser"
)

// TestRefreshSignupURLs tests PDF URL extraction for tournament signup
// It first checks for a signup link in the pdf (with french phrases like "s'inscrire ici" or "inscription sur ...")
// Often, that link is not the one we're looking for (not the direct signup link, but it still could be), but rather a link to the organizing club's website
// Often, when the link is not the direct signup link, you will find that the landing page (after clicking on the link in the PDF rules) can contain a direct link the signup page
// That direct link may or may or may not be on the same domain
// It can happen that the landing page contains a link to looks like it will take you to the signup page, but the target page is not the signup page. This does not necessarily mean that the wrong link was clicked, but rather another page containing more info on the tournament, with again a link redirecting you to a dedicated signup page. This recursive logic can happen for a a few pages before reaching the signup form.
// The signup form is the indication that the signup page has been reached
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
			extractedURL, err := extraction.ExtractSignupURLFromPDF(
				tournament,
				tournamentDate,
				tc.PDFRulesURL,
				browserContext,
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

// Create another function just like the one above, but with HelloAsso expected URLs
// The idea is to test that our system will detect the HelloAsso url in the pdf, navigate to it to see if it matches with the tournament date and save it directly without looking for signup forms because we know that it's helloasso
