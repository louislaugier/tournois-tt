package pdf_processor

import (
	"log"
	"strings"
	"time"

	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/pdf"
	"tournois-tt/api/pkg/scraper/services/helloasso"
	"tournois-tt/api/pkg/utils"

	pw "github.com/playwright-community/playwright-go"
)

// processExtractedPDFText processes extracted PDF text to find and validate signup URLs
func processExtractedPDFText(pdfText string, patterns *pdf.RegexPatterns, tournament cache.TournamentCache,
	tournamentDate time.Time, browserContext pw.BrowserContext,
	validator func(string, cache.TournamentCache, time.Time, pw.BrowserContext) (string, error)) (string, error) {
	// First, check for direct signup references in the PDF text
	signupMatches := pdf.FindURLsByPattern(pdfText, patterns.SignupURLRegex)
	if len(signupMatches) > 0 {
		utils.DebugLog("Found %d explicit signup references in PDF", len(signupMatches))

		// Try to validate these explicit signup URLs
		validURL, found := tryValidateURLs(signupMatches, tournament, tournamentDate, browserContext, validator)
		if found {
			log.Printf("Found valid signup URL from explicit signup reference: %s", validURL)
			return validURL, nil
		}
	}

	// Next, check for tournament-specific subdomains like "tournoi.cctt.fr" which are explicitly mentioned
	tournoiURLs := pdf.FindURLsByPattern(pdfText, patterns.TournoiSubdomainRegex)
	if len(tournoiURLs) > 0 {
		utils.DebugLog("Found %d explicit tournoi subdomain URLs in PDF", len(tournoiURLs))

		// For each tournoi subdomain URL, try the following paths:
		for _, tournoiURL := range tournoiURLs {
			// Clean up the URL
			tournoiBase := pdf.EnsureURLProtocol(tournoiURL)
			if strings.HasSuffix(tournoiBase, "/") {
				tournoiBase = tournoiBase[:len(tournoiBase)-1]
			}

			// Common signup URL paths for tournament sites
			signupPaths := []string{
				"/sign-up/",
				"/signup/",
				"/inscription/",
				"/inscriptions/",
				"/register/",
				"/registration/",
				"/participer/",
				"/compte/",
				"/",
			}

			utils.DebugLog("Trying various signup paths for tournoi domain: %s", tournoiBase)
			for _, path := range signupPaths {
				signupURL := tournoiBase + path
				utils.DebugLog("Validating potential signup URL: %s", signupURL)

				// Validate this URL
				validURL, err := validator(signupURL, tournament, tournamentDate, browserContext)
				if err != nil {
					utils.DebugLog("Error validating URL %s: %v", signupURL, err)
					continue
				}

				if validURL != "" {
					log.Printf("Found valid signup URL from tournoi subdomain: %s", validURL)
					return validURL, nil
				}
			}
		}
	}

	// Check for payment references - lower priority than direct signup references
	paymentMatches := pdf.FindURLsByPattern(pdfText, patterns.PaymentURLRegex)
	if len(paymentMatches) > 0 {
		utils.DebugLog("Found %d payment references in PDF", len(paymentMatches))

		// Try to validate payment reference URLs
		validURL, found := tryValidateURLs(paymentMatches, tournament, tournamentDate, browserContext, validator)
		if found {
			log.Printf("Found valid signup URL from payment reference: %s", validURL)
			return validURL, nil
		}
	}

	// Check for HelloAsso URLs next
	helloAssoURLs := pdf.FindURLsByPattern(pdfText, helloasso.URLRegex())
	if len(helloAssoURLs) > 0 {
		utils.DebugLog("Found %d HelloAsso URLs in PDF", len(helloAssoURLs))

		// Limit the number of URLs to validate
		urlsToValidate := pdf.LimitURLs(helloAssoURLs, MaxURLsToProcess)

		// Try to validate the HelloAsso URLs
		validURL, found := tryValidateURLs(urlsToValidate, tournament, tournamentDate, browserContext, validator)
		if found {
			return validURL, nil
		}
	}

	// Look for domain-only registration instructions, like "inscriptions sur site cctt.fr"
	domainURLs := pdf.FindDomainOnlyReferences(pdfText)
	if len(domainURLs) > 0 {
		utils.DebugLog("Found %d domain-only references in PDF", len(domainURLs))

		// First try the explicit tournoi.domain.tld pattern for each domain
		for _, domainURL := range domainURLs {
			// Extract the domain without protocol
			domain := domainURL
			if strings.HasPrefix(domain, "https://") {
				domain = strings.TrimPrefix(domain, "https://")
			} else if strings.HasPrefix(domain, "http://") {
				domain = strings.TrimPrefix(domain, "http://")
			}

			// Skip if this is already a tournoi subdomain (we handled those above)
			if strings.HasPrefix(domain, "tournoi.") {
				continue
			}

			// Try the tournoi subdomain with various signup paths
			tournoiDomain := "https://tournoi." + domain
			utils.DebugLog("Trying tournoi subdomain for domain reference: %s", tournoiDomain)

			// Common signup URL paths for tournament sites
			signupPaths := []string{
				"/sign-up/",
				"/signup/",
				"/inscription/",
				"/inscriptions/",
				"/register/",
				"/registration/",
				"/participer/",
				"/compte/",
				"/",
			}

			for _, path := range signupPaths {
				signupURL := tournoiDomain + path
				utils.DebugLog("Validating potential signup URL: %s", signupURL)

				// Validate this URL
				validURL, err := validator(signupURL, tournament, tournamentDate, browserContext)
				if err != nil {
					utils.DebugLog("Error validating URL %s: %v", signupURL, err)
					continue
				}

				if validURL != "" {
					log.Printf("Found valid signup URL from domain reference: %s", validURL)
					return validURL, nil
				}
			}
		}

		// Try to find signup URLs on each domain's homepage
		validURL, err := validateDomainURLs(domainURLs, tournament, tournamentDate, browserContext, validator)
		if err != nil {
			utils.DebugLog("Error validating domain URLs: %v", err)
		} else if validURL != "" {
			log.Printf("Found valid signup URL on domain website: %s", validURL)
			return validURL, nil
		}

		// Try with common tournament subdomains for each domain
		utils.DebugLog("Trying common tournament subdomains for %d domains", len(domainURLs))
		for _, domainURL := range domainURLs {
			for _, subdomain := range pdf.GenerateCommonTournamentSubdomains(domainURL) {
				utils.DebugLog("Validating generated subdomain: %s", subdomain)
				validURL, err := validator(subdomain, tournament, tournamentDate, browserContext)
				if err != nil {
					utils.DebugLog("Error validating subdomain %s: %v", subdomain, err)
					continue
				}

				if validURL != "" {
					log.Printf("Found valid signup URL from generated subdomain: %s", validURL)
					return validURL, nil
				}
			}
		}
	}

	// Last resort, look for any URLs that might be registration related
	// but weren't caught by the specific patterns above
	registrationURLs := findRegistrationURLsInPDF(pdfText)
	if len(registrationURLs) > 0 {
		utils.DebugLog("Found %d potential registration URLs in PDF", len(registrationURLs))

		// Try recursive form navigation
		validURL, found, err := tryRecursiveFormNavigation(registrationURLs, tournament, tournamentDate, browserContext, validator)
		if err != nil {
			log.Printf("Error in recursive form navigation: %v", err)
		} else if found {
			log.Printf("Found valid signup URL via recursive navigation: %s", validURL)
			return validURL, nil
		}
	}

	utils.DebugLog("No valid signup URL found in PDF")
	return "", nil
}
