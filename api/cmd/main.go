package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
	"tournois-tt/api/internal/crons"
	"tournois-tt/api/internal/router"
	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/finder"
	"tournois-tt/api/pkg/helloasso"
	"tournois-tt/api/pkg/pdf"
	"tournois-tt/api/pkg/scraper/browser"
	"tournois-tt/api/scripts"
)

// Run geocoding refresh in a background goroutine
// go func() {
// 	tournaments.RefreshTournamentsAndGeocoding()
// 	tournaments.RefreshSignupURLs()
// }()

func start() {
	crons.Schedule()

	r := router.NewRouter()

	log.Printf("Server starting...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func main() {
	start()
	// test()
}

func test() {
	// browser.CheckBrowserInstallation()

	tournament := findTournament3026()
	if tournament == nil {
		return
	}

	// Extract the content first, as this shouldn't be affected by browser issues
	log.Println("Extracting PDF content from tournament rules URL...")
	content, err := pdf.ExtractFileContentFromURL(tournament.Rules.URL)
	if err != nil {
		log.Printf("Failed to extract content from URL: %v", err)
		return
	}

	log.Println("Getting signup URL from rules content...")
	signupURL := finder.GetSignupURLFromRulesContent(content)
	if signupURL == nil {
		log.Println("nil signupURL")
		return
	}
	log.Printf("signupURL: %s", *signupURL)

	log.Println("Extracting all URLs from rules content...")
	URLs := finder.GetURLsFromText(content)
	log.Printf("Extracted URLs: %v", URLs)

	log.Println("Determining most probable signup URL...")
	URL := finder.GetMostProbableSignupURL(URLs, content)
	log.Printf("Most probable URL: %s", URL)

	if URL == "" {
		log.Println("Unable to find a valid URL for registration form")
		return
	}

	// URL = "cctt.fr"
	// todo: détecter cognac international form

	// Ensure the URL has the necessary protocol prefix
	if !strings.HasPrefix(URL, "http://") && !strings.HasPrefix(URL, "https://") {
		URL = "https://" + URL
		log.Printf("Added protocol prefix to URL: %s", URL)
	}

	defer func() {
		log.Println("Shutting down browser at end of main function")
		browser.ShutdownBrowser()
	}()

	// SIMPLIFIED APPROACH FOR DOCKER ENVIRONMENT
	// We'll use a single browser instance and page for validation
	// rather than the recursive search which tends to cause issues
	log.Println("Using Docker-optimized mode for signup validation")

	// Build the browser with multiple retries
	log.Println("Setting up browser with careful retries...")
	_, _, _, err = browser.SetupWithRetry(5)
	if err != nil {
		log.Printf("ERROR: Browser setup failed even after retries: %v", err)
		return
	}

	log.Println("Browser setup completed successfully")

	// Wait for the browser to stabilize
	log.Println("Waiting for browser to stabilize...")
	time.Sleep(2 * time.Second)

	// Check if browser is healthy before attempting to use it
	log.Println("Checking browser health before proceeding...")
	if !browser.IsHealthy() {
		log.Println("Browser is unhealthy, attempting to restore it")
		restarted, err := browser.RestartIfUnhealthy()
		if err != nil || !restarted {
			log.Printf("Failed to restore browser health: %v", err)
			return
		}
		// Allow time for the browser to stabilize after restart
		log.Println("Waiting for browser to stabilize after restart...")
		time.Sleep(2 * time.Second)
	}

	// Simple direct validation with careful timeout handling
	log.Println("Performing direct URL validation instead of recursive search...")
	signupURLResult := ""

	// Set a timeout for the validation operation
	timeoutCh := make(chan bool, 1)
	resultCh := make(chan struct {
		url string
		err error
	}, 1)

	// Run the validation in a goroutine with panic recovery
	log.Println("Starting URL validation in background goroutine...")
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("PANIC RECOVERED in validation: %v", r)
				resultCh <- struct {
					url string
					err error
				}{"", fmt.Errorf("panic occurred: %v", r)}
			}
		}()

		log.Println("Getting signup link from URL")
		// First fetch HTML content from the URL
		log.Printf("Fetching content from URL: %s", URL)
		htmlContent, err := browser.FetchPageContent(URL, 2) // Use 2 retries
		if err != nil {
			log.Printf("Error fetching HTML content: %v", err)
			resultCh <- struct {
				url string
				err error
			}{"", err}
			return
		}

		// Then extract the signup link from the HTML content
		signupLink := finder.GetSignupLinkFromHTML(htmlContent, URL)
		var result string
		if signupLink != nil {
			log.Printf("Fetching content from URL: %s", URL)
			htmlContent, err := browser.FetchPageContent(*signupLink, 2) // Use 2 retries
			if err != nil {
				log.Printf("Error fetching HTML content: %v", err)
				resultCh <- struct {
					url string
					err error
				}{"", err}
				return
			}

			isSignupFormPage := finder.IsSignupFormPage(htmlContent)
			if isSignupFormPage {
				result = *signupLink
			} else {
				newLink := finder.GetSignupLinkFromHTML(htmlContent, *signupLink)
				if newLink != nil {
					// Try recursively up to 3 more times
					maxDepth := 3

					// Simple recursive verification
					var verifyRecursively func(link string, depth int) string
					verifyRecursively = func(link string, depth int) string {
						if depth <= 0 {
							return ""
						}

						log.Printf("Recursive verification attempt %d for URL: %s", maxDepth-depth+1, link)
						content, err := browser.FetchPageContent(link, 1)
						if err != nil {
							log.Printf("Error in recursive verification: %v", err)
							return ""
						}

						if finder.IsSignupFormPage(content) {
							return link
						}

						nextLink := finder.GetSignupLinkFromHTML(content, link)
						if nextLink != nil {
							return verifyRecursively(*nextLink, depth-1)
						}

						return ""
					}

					result = verifyRecursively(*newLink, maxDepth)
				}
			}
		}

		log.Println("Scraping completed, sending results")
		resultCh <- struct {
			url string
			err error
		}{result, nil}
	}()

	// Set timeout for the operation (60 seconds)
	log.Println("Starting timeout timer (60 seconds)...")
	go func() {
		time.Sleep(60 * time.Second)
		log.Println("Timeout timer expired, sending timeout signal")
		timeoutCh <- true
	}()

	// Wait for either result or timeout
	log.Println("Waiting for either validation result or timeout...")
	select {
	case result := <-resultCh:
		if result.err != nil {
			log.Printf("Error validating signup URL: %v", result.err)
		} else if result.url != "" {
			signupURLResult = result.url
			log.Println("Signup URL validated successfully:", signupURLResult)
		} else {
			log.Println("URL does not appear to be a valid signup form")
		}
	case <-timeoutCh:
		log.Println("Validation operation timed out after 60 seconds")
	}

	// Return the result
	log.Println("ok789", signupURLResult, "")
	log.Println("======== Application completed ========")

}

func findTournament3026() *cache.TournamentCache {
	log.Println("Loading tournaments from cache...")
	tournaments, err := cache.LoadTournaments()
	if err != nil {
		log.Fatalf("Failed to load tournaments: %v", err)
	}

	targetID := "3026"
	log.Printf("Looking for tournament with ID %s", targetID)
	targetTournament, found := tournaments[targetID]

	if !found {
		log.Printf("Tournament with ID 3026 not found in cache")
		return nil
	}

	if targetTournament.Rules != nil {
		log.Printf("Rules URL: %s", targetTournament.Rules.URL)
	} else {
		log.Printf("Tournament has no rules URL")
	}

	return &targetTournament
}

func test2() {
	activities, err := helloasso.SearchActivities(context.Background(), "tournoi tennis de table courbevoie")
	if err != nil {
		log.Println(err)
	}
	log.Println(activities)

	scripts.LogClubEmailAddresses()
	scripts.LogCommitteeAndLeagueEmailAddresses()
}
