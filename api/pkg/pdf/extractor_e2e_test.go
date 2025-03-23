package pdf_test

import (
	"fmt"
	"testing"
	"time"
	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/pdf"
	"tournois-tt/api/pkg/scraper/browser"
)

// TestRefreshSignupURLs tests PDF URL extraction for tournament signup
// It first checks for a signup link in the pdf (with french phrases like "s'inscrire ici" or "inscription sur ...")
// Often, that link is not the one we're looking for (not the direct signup link, but it still could be), but rather a link to the organizing club's website
// Often, when the link is not the direct signup link, the landing page (after clicking on the link in the PDF rules) can contain a direct link the signup page
// That direct link may or may or may not be on the same domain
// It can happen that the landing page contains a link that looks like it will take you to the signup page, but the target page is not the signup page. This does not necessarily mean that the wrong link was clicked, but rather another page containing more info on the tournament, with again a link redirecting you to a dedicated signup page. This recursive logic can happen for a a few pages before reaching the signup form.
// The signup form being visible is the indication that the final signup url for the tournament has been reached
func TestExtractSignupURLFromPDF(t *testing.T) {
	// Skip the test by default since it's resource-intensive
	// To run this test, use: go test -tags=e2e ./api/internal/crons/tournaments/... -v -run TestRefreshSignupURLs
	//t.Skip("Skipping e2e test that requires a browser instance. Run with -tags=e2e to include")

	// Test cases
	testCases := []struct {
		name                   string
		PDFRulesURL            string
		ExpectedFinalSignupURL string
	}{
		// Example of a tournament with a signup URL in the PDF
		{
			name:                   "Noisy-le-Grand",
			PDFRulesURL:            "https://apiv2.fftt.com/api/files/357572/reglement%20tournoi%202025%20fftt%20(2).pdf",
			ExpectedFinalSignupURL: "https://tournoi.cctt.fr/sign-up/",
		},
	}

	// Setup browser
	_, _, browserContext, err := browser.Setup()
	if err != nil {
		t.Fatalf("Failed to set up browser: %v", err)
	}
	defer browser.ShutdownBrowser()

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock tournament
			tournament := cache.TournamentCache{
				Name:      tc.name,
				StartDate: "2022-06-15",
				EndDate:   "2022-06-16",
				// Use a club name derived from the test case name
				Club: cache.Club{
					Name: tc.name,
				},
			}
			tournamentDate, _ := time.Parse("2006-01-02", tournament.StartDate)

			// Simple logging
			fmt.Printf("Testing %s: %s\n", tc.name, tc.PDFRulesURL)

			// Extract URL
			extractedURL, err := pdf.ExtractSignupURLFromPDF(
				tournament,
				tournamentDate,
				tc.PDFRulesURL,
				browserContext,
			)

			// Log extraction result
			if err != nil {
				t.Logf("Error extracting URL: %v", err)
			} else if extractedURL != "" {
				t.Logf("Found URL: %s", extractedURL)
			} else {
				t.Logf("No URL found.")
			}

			// Test expectations
			if extractedURL != tc.ExpectedFinalSignupURL {
				t.Errorf("Expected to find a signup URL for %s, but found %s", extractedURL)
			}
		})
	}
}

// Create another function just like the one above, but with HelloAsso expected URLs
// The idea is to test that our system will detect the HelloAsso url in the pdf, navigate to it to see if it matches with the tournament date and save it directly without looking for signup forms because we know that it's helloasso
