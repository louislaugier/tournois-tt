package tournaments_test

import (
	"fmt"
	"net/url"
	"testing"
	"time"

	"tournois-tt/api/crons/tournaments/signup"
	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/scraper/browser"
)

// TestRefreshSignupURLs tests PDF URL extraction for tournament signup
func TestRefreshSignupURLs(t *testing.T) {
	t.Parallel()
	signup.Debug = true

	testCases := []struct {
		name         string
		pdfURL       string
		expectedURL  string
		tournamentID int
		clubName     string
	}{
		{
			name:         "CCTT 2025",
			pdfURL:       "https://apiv2.fftt.com/api/files/357572/reglement%20tournoi%202025%20fftt%20(2).pdf",
			expectedURL:  "https://tournoi.cctt.fr/sign-up/",
			tournamentID: 357572,
			clubName:     "CCTT Châlons-en-Champagne",
		},
		// {
		// 	name:         "PPCF 2025",
		// 	pdfURL:       "https://apiv2.fftt.com/api/files/390629/reglement_tournoi_2025.pdf",
		// 	expectedURL:  "https://www.ppcflines.fr/tournoi/inscription-etape1",
		// 	tournamentID: 390629,
		// 	clubName:     "PPCF Lines",
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
			// Create tournament
			tournament := cache.TournamentCache{
				ID:        tc.tournamentID,
				Name:      "Tournoi " + tc.name,
				StartDate: time.Now().Add(24 * time.Hour).Format("2006-01-02"),
				Club: cache.Club{
					Name: tc.clubName,
				},
			}
			tournamentDate, _ := time.Parse("2006-01-02", tournament.StartDate)

			// Extract URL
			extractedURL, err := signup.ExtractSignupURLFromPDFFile(
				tournament,
				tournamentDate,
				tc.pdfURL,
				browserContext,
			)

			// Log any extraction errors but continue
			if err != nil {
				t.Logf("Error extracting URL: %v", err)
			}

			// Print results
			fmt.Printf("Test %s: Found URL: %s\n", tc.name, extractedURL)
			fmt.Printf("Expected URL: %s\n", tc.expectedURL)

			// Validate the result
			if extractedURL == "" {
				t.Errorf("No URL found for %s", tc.name)
				return
			}

			// Compare domains
			extractedDomain := getDomainFromURL(extractedURL)
			expectedDomain := getDomainFromURL(tc.expectedURL)

			if extractedURL == tc.expectedURL {
				fmt.Printf("SUCCESS: URLs match exactly\n")
			} else if extractedDomain == expectedDomain {
				fmt.Printf("SUCCESS: URLs have same domain: %s\n", extractedDomain)
			} else {
				t.Errorf("URL %s does not match expected URL %s", extractedURL, tc.expectedURL)
			}
		})
	}
}

// getDomainFromURL extracts the domain from a URL
func getDomainFromURL(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return parsedURL.Hostname()
}
