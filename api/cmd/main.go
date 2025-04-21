package main

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
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
	// testScraping()
	// scripts.RegeocodeFailedTournaments()
}

func testScraping() {
	// browser.CheckBrowserInstallation()

	// tournament := findTournament(3043) // cognac
	// tournament := findTournament(3026) // chalons en champagne
	tournament := findTournament(2799) // epinal
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
	// URL = "www.cognactennisdetable.fr"

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

		// Log the full HTML for debugging
		log.Printf("====== BEGIN HTML CONTENT OF %s ======", URL)
		log.Println(htmlContent)
		log.Printf("====== END HTML CONTENT ======")

		// Special check for inscription link
		inscriptionRegex := regexp.MustCompile(`<a[^>]*href=["']([^"']*inscription[^"']*)["'][^>]*>`)
		inscriptionMatches := inscriptionRegex.FindAllStringSubmatch(htmlContent, -1)
		if len(inscriptionMatches) > 0 {
			log.Printf("FOUND %d INSCRIPTION LINKS DIRECTLY:", len(inscriptionMatches))
			for i, match := range inscriptionMatches {
				if len(match) >= 2 {
					log.Printf("  INSCRIPTION LINK #%d: %s", i+1, match[1])
				}
			}
		} else {
			log.Println("NO DIRECT INSCRIPTION LINKS FOUND IN HTML")
		}

		// Then extract the signup link from the HTML content
		signupLink := finder.GetSignupLinkFromHTML(htmlContent, URL)
		var result string
		if signupLink != nil {
			log.Printf("Found a candidate signup link: %s", *signupLink)
			log.Printf("Fetching content from URL: %s", *signupLink)
			htmlContent, err := browser.FetchPageContent(*signupLink, 2) // Use 2 retries
			if err != nil {
				log.Printf("Error fetching HTML content: %v", err)
				resultCh <- struct {
					url string
					err error
				}{"", err}
				return
			}

			// Log the full HTML for debugging
			log.Printf("====== BEGIN HTML CONTENT OF %s ======", *signupLink)
			log.Println(htmlContent)
			log.Printf("====== END HTML CONTENT ======")

			// Special check for inscription link
			inscriptionRegex := regexp.MustCompile(`<a[^>]*href=["']([^"']*inscription[^"']*)["'][^>]*>`)
			inscriptionMatches := inscriptionRegex.FindAllStringSubmatch(htmlContent, -1)
			if len(inscriptionMatches) > 0 {
				log.Printf("FOUND %d INSCRIPTION LINKS ON SIGNUP PAGE:", len(inscriptionMatches))
				for i, match := range inscriptionMatches {
					if len(match) >= 2 {
						inscriptionLink := match[1]

						// Resolve the URL if it's relative
						if !strings.HasPrefix(inscriptionLink, "http://") && !strings.HasPrefix(inscriptionLink, "https://") {
							if strings.HasPrefix(inscriptionLink, "/") {
								// Absolute path on the site
								baseURL := getBaseURL(*signupLink)
								inscriptionLink = baseURL + inscriptionLink
							} else {
								// Relative path
								resolvedURL := *signupLink
								if !strings.HasSuffix(resolvedURL, "/") {
									lastSlashIdx := strings.LastIndex(resolvedURL, "/")
									if lastSlashIdx > 0 {
										resolvedURL = resolvedURL[:lastSlashIdx+1]
									} else {
										resolvedURL += "/"
									}
								}
								inscriptionLink = resolvedURL + inscriptionLink
							}
						}

						log.Printf("  INSCRIPTION LINK #%d: %s", i+1, inscriptionLink)

						// Set this inscription link as the result
						if strings.Contains(strings.ToLower(inscriptionLink), "inscription") {
							log.Printf("FOUND DIRECT INSCRIPTION LINK: %s", inscriptionLink)
							result = inscriptionLink
							break
						}
					}
				}
			} else {
				log.Println("NO DIRECT INSCRIPTION LINKS FOUND ON SIGNUP PAGE")
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
					var verifyRecursively func(link string, depth int, visitedURLs map[string]bool) string
					verifyRecursively = func(link string, depth int, visitedURLs map[string]bool) string {
						if depth <= 0 {
							log.Printf("Reached maximum recursion depth, stopping")
							return ""
						}

						// Check if we've already visited this URL to prevent loops
						normalizedLink := strings.TrimSuffix(link, "/")
						if visitedURLs[normalizedLink] {
							log.Printf("Already visited URL: %s, skipping to prevent loops", link)
							return ""
						}

						// Mark this URL as visited
						visitedURLs[normalizedLink] = true
						log.Printf("Adding to visited URLs: %s", normalizedLink)

						log.Printf("Recursive verification attempt %d for URL: %s", maxDepth-depth+1, link)
						content, err := browser.FetchPageContent(link, 1)
						if err != nil {
							log.Printf("Error in recursive verification: %v", err)
							return ""
						}

						// Check if this is a signup form page
						if finder.IsSignupFormPage(content) {
							log.Printf("Found a valid signup form page at: %s", link)
							return link
						}

						log.Printf("Page at %s is not a valid signup form, looking for links", link)

						// Analyze content for potential signup links that haven't been visited
						nextLink := finder.GetSignupLinkFromHTML(content, link)
						if nextLink != nil {
							normalizedNextLink := strings.TrimSuffix(*nextLink, "/")

							// Skip if it's a link we've already processed
							if visitedURLs[normalizedNextLink] {
								log.Printf("Found a link we already visited: %s, skipping", *nextLink)
								return ""
							}

							// Check if the link points to a signup page directly
							log.Printf("Found a link on intermediate page: %s", *nextLink)
							nextContent, err := browser.FetchPageContent(*nextLink, 1)
							if err != nil {
								log.Printf("Error fetching content for next link: %v", err)
							} else {
								if finder.IsSignupFormPage(nextContent) {
									log.Printf("Found a valid signup form at: %s", *nextLink)
									return *nextLink
								}

								log.Printf("Page at %s is not a valid signup form, continuing recursion", *nextLink)

								// Special handling for potential special registration pages
								lowerContent := strings.ToLower(nextContent)
								if strings.Contains(lowerContent, "inscription") ||
									strings.Contains(lowerContent, "register") ||
									strings.Contains(lowerContent, "signup") {
									log.Printf("Page contains signup keywords but doesn't match form criteria, adding to candidates")
								}
							}

							// Continue recursion with next link
							result := verifyRecursively(*nextLink, depth-1, visitedURLs)
							if result != "" {
								return result
							}

							// If recursion didn't find anything, try the original page for more links
							log.Printf("Recursion didn't find a valid signup form, looking for more links on: %s", link)
							otherLinks := finder.FindAllSignupLinksFromATags(content, []string{"inscription", "signup", "register"}, link)
							for _, otherLink := range otherLinks {
								normalizedOtherLink := strings.TrimSuffix(otherLink, "/")
								if !visitedURLs[normalizedOtherLink] {
									log.Printf("Trying alternative link: %s", otherLink)
									visitedURLs[normalizedOtherLink] = true
									otherResult := verifyRecursively(otherLink, depth-1, visitedURLs)
									if otherResult != "" {
										return otherResult
									}
								}
							}
						} else {
							log.Printf("No suitable links found on page: %s", link)
						}

						return ""
					}

					// Initialize the map of visited URLs
					visitedURLs := make(map[string]bool)
					result = verifyRecursively(*newLink, maxDepth, visitedURLs)

					// If no signup form was found but we have a candidate link with inscription keywords
					if result == "" {
						log.Printf("No valid signup form found, returning best candidate link: %s", *newLink)
						result = *newLink
					}
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

func findTournament(ID int) *cache.TournamentCache {
	log.Println("Loading tournaments from cache...")
	tournaments, err := cache.LoadTournaments()
	if err != nil {
		log.Fatalf("Failed to load tournaments: %v", err)
	}

	targetID := strconv.Itoa(ID)
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

func testHelloAsso() {
	activities, err := helloasso.SearchActivities(context.Background(), "tournoi tennis de table courbevoie")
	if err != nil {
		log.Println(err)
	}
	log.Println(activities)

	scripts.LogClubEmailAddresses()
	scripts.LogCommitteeAndLeagueEmailAddresses()
}

// Helper function to get the base URL (protocol + domain) from a full URL
func getBaseURL(url string) string {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		parts := strings.SplitN(url, "/", 4)
		if len(parts) >= 3 {
			return parts[0] + "//" + parts[2] // protocol + domain
		}
	}
	return url
}
