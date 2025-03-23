package signup

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"tournois-tt/api/pkg/cache"
	"tournois-tt/api/pkg/pdf"

	pw "github.com/playwright-community/playwright-go"
)

// PDF processing constants
const (
	maxURLsToProcess = 30 // Limit number of URLs to process to avoid excessive validation
	maxRedirections  = 5  // Maximum number of recursively followed links to find signup form
)

// ExtractSignupURLFromPDFFile extracts and validates signup URLs from PDF content
func ExtractSignupURLFromPDFFile(tournament cache.TournamentCache, tournamentDate time.Time, rulesURL string, browserContext pw.BrowserContext) (string, error) {
	debugLog("Checking rules PDF for signup URL: %s", rulesURL)

	// Maximum number of retry attempts for PDF extraction errors
	const maxPdfRetries = 3
	// Delay between retries (increases with each retry)
	var retryDelayBase = 5 * time.Second

	// Verify PDF file validity
	if !isPDFFile(rulesURL) {
		return "", fmt.Errorf("rules file is not a PDF: %s", rulesURL)
	}

	// Extract text from PDF using the pkg/pdf implementation with retries
	debugLog("Extracting text from PDF: %s", rulesURL)

	var pdfText string
	var fetchDuration, processDuration time.Duration
	var attemptsMade int

	// First, try to read the text from a pre-extracted file (useful for testing)
	tempFilePath := "/tmp/debug_pdf.txt"
	if fileContent, err := readFileContent(tempFilePath); err == nil && fileContent != "" {
		debugLog("Using pre-extracted PDF text from %s (%d characters)", tempFilePath, len(fileContent))
		pdfText = fileContent
	} else {
		// If no pre-extracted file exists, extract the text from the PDF
		for attemptsMade = 0; attemptsMade < maxPdfRetries; attemptsMade++ {
			// Exponential backoff on retries
			if attemptsMade > 0 {
				retryDelay := retryDelayBase * time.Duration(attemptsMade)
				log.Printf("PDF extraction error, retrying in %v (attempt %d/%d)",
					retryDelay, attemptsMade+1, maxPdfRetries)
				time.Sleep(retryDelay)
			}

			result := pdf.ProcessURLWithExtractor(rulesURL, pdf.ExtractTextFromBytes)
			fetchDuration = result.FetchDuration
			processDuration = result.Duration

			// If no error or not a temporary error, break the retry loop
			if result.Error == nil {
				pdfText = result.Text
				break
			}

			// Check if this is a temporary error that can be retried
			if !isPdfExtractionRetryableError(result.Error.Error()) {
				return "", fmt.Errorf("failed to extract text from PDF: %w", result.Error)
			}
		}

		// If we still don't have any text after all retries
		if pdfText == "" {
			return "", fmt.Errorf("failed to extract text from PDF after %d attempts", attemptsMade+1)
		}
	}

	debugLog("PDF processing took %v (fetch: %v, extraction: %v)",
		(fetchDuration + processDuration).Round(time.Millisecond),
		fetchDuration.Round(time.Millisecond),
		processDuration.Round(time.Millisecond))

	// Process the PDF text to extract URLs using the URL extraction logic
	return processExtractedPDFText(pdfText, tournament, tournamentDate, browserContext)
}

// readFileContent reads the content of a file and returns it as a string
func readFileContent(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// isPDFFile checks if a URL points to a PDF file based on extension
func isPDFFile(urlStr string) bool {
	ext := strings.ToLower(filepath.Ext(urlStr))
	return ext == ".pdf"
}

// isPdfExtractionRetryableError determines if a PDF extraction error can be retried
func isPdfExtractionRetryableError(errStr string) bool {
	// Check for typical temporary PDF extraction error patterns
	retryablePatterns := []string{
		"timeout",
		"connection reset",
		"connection refused",
		"temporary",
		"network",
		"stream error",
		"EOF",
		"unexpected EOF",
		"HTTP status",
		"TLS handshake",
		"download",
		"i/o timeout",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(strings.ToLower(errStr), strings.ToLower(pattern)) {
			return true
		}
	}

	return false
}

// processExtractedPDFText processes extracted PDF text to find and validate signup URLs
func processExtractedPDFText(pdfText string, tournament cache.TournamentCache, tournamentDate time.Time, browserContext pw.BrowserContext) (string, error) {
	// First, check for direct signup references in the PDF text
	signupMatches := FindURLsByPattern(pdfText, GetSignupURLRegex())
	if len(signupMatches) > 0 {
		debugLog("Found %d explicit signup references in PDF", len(signupMatches))
		
		// Try to validate these explicit signup URLs
		validURL, found := tryValidateURLs(signupMatches, tournament, tournamentDate, browserContext)
		if found {
			log.Printf("Found valid signup URL from explicit signup reference: %s", validURL)
			return validURL, nil
		}
	}
	
	// Next, check for tournament-specific subdomains like "tournoi.cctt.fr" which are explicitly mentioned
	tournoiURLs := FindURLsByPattern(pdfText, GetTournoiSubdomainRegex())
	if len(tournoiURLs) > 0 {
		debugLog("Found %d explicit tournoi subdomain URLs in PDF", len(tournoiURLs))
		
		// For each tournoi subdomain URL, try the following paths:
		for _, tournoiURL := range tournoiURLs {
			// Clean up the URL
			tournoiBase := ensureURLProtocol(tournoiURL)
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
			
			debugLog("Trying various signup paths for tournoi domain: %s", tournoiBase)
			for _, path := range signupPaths {
				signupURL := tournoiBase + path
				debugLog("Validating potential signup URL: %s", signupURL)
				
				// Validate this URL
				validURL, err := ValidateSignupURL(signupURL, tournament, tournamentDate, browserContext)
				if err != nil {
					debugLog("Error validating URL %s: %v", signupURL, err)
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
	paymentMatches := FindURLsByPattern(pdfText, GetPaymentURLRegex())
	if len(paymentMatches) > 0 {
		debugLog("Found %d payment references in PDF", len(paymentMatches))
		
		// Try to validate payment reference URLs
		validURL, found := tryValidateURLs(paymentMatches, tournament, tournamentDate, browserContext)
		if found {
			log.Printf("Found valid signup URL from payment reference: %s", validURL)
			return validURL, nil
		}
	}
	
	// Check for HelloAsso URLs next
	helloAssoURLs := FindURLsByPattern(pdfText, helloAssoURLRegex)
	if len(helloAssoURLs) > 0 {
		debugLog("Found %d HelloAsso URLs in PDF", len(helloAssoURLs))

		// Limit the number of URLs to validate
		urlsToValidate := limitURLs(helloAssoURLs, maxURLsToProcess)

		// Try to validate the HelloAsso URLs
		validURL, found := tryValidateURLs(urlsToValidate, tournament, tournamentDate, browserContext)
		if found {
			return validURL, nil
		}
	}

	// Look for domain-only registration instructions, like "inscriptions sur site cctt.fr"
	domainURLs := findDomainOnlyReferences(pdfText)
	if len(domainURLs) > 0 {
		debugLog("Found %d domain-only references in PDF", len(domainURLs))

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
			debugLog("Trying tournoi subdomain for domain reference: %s", tournoiDomain)
			
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
				debugLog("Validating potential signup URL: %s", signupURL)
				
				// Validate this URL
				validURL, err := ValidateSignupURL(signupURL, tournament, tournamentDate, browserContext)
				if err != nil {
					debugLog("Error validating URL %s: %v", signupURL, err)
					continue
				}
				
				if validURL != "" {
					log.Printf("Found valid signup URL from tournoi subdomain: %s", validURL)
					return validURL, nil
				}
			}
		}

		// If direct tournoi subdomain approach didn't work, try recursive navigation
		// Limit the number of URLs to validate
		urlsToValidate := limitURLs(domainURLs, maxURLsToProcess)

		// Try to validate the domain URLs with recursive navigation
		validURL, found, err := tryRecursiveFormNavigation(urlsToValidate, tournament, tournamentDate, browserContext)
		if err != nil {
			log.Printf("Warning: Error during recursive form navigation: %v", err)
		}
		if found {
			return validURL, nil
		}
	}

	// If no HelloAsso URLs or domain references, look for registration-related URLs
	debugLog("Looking for registration-related URLs in PDF")
	registrationURLs := findRegistrationURLsInPDF(pdfText)

	if len(registrationURLs) > 0 {
		debugLog("Found %d potential registration URLs in PDF", len(registrationURLs))

		// Limit the number of URLs to validate
		urlsToValidate := limitURLs(registrationURLs, maxURLsToProcess)

		// Try to validate the registration URLs first with simple validation
		validURL, found := tryValidateURLs(urlsToValidate, tournament, tournamentDate, browserContext)
		if found {
			return validURL, nil
		}

		// If simple validation didn't find anything, try recursive navigation
		validURL, found, err := tryRecursiveFormNavigation(urlsToValidate, tournament, tournamentDate, browserContext)
		if err != nil {
			log.Printf("Warning: Error during recursive form navigation: %v", err)
		}
		if found {
			return validURL, nil
		}
	}

	// Find domain references in the PDF text
	domains := findDomainsInText(pdfText)
	debugLog("Found %d domain references in PDF: %v", len(domains), domains)

	if len(domains) > 0 {
		// First, try to validate any domains directly
		validURL, err := validateDomainURLs(domains, tournament, tournamentDate, browserContext)
		if err != nil {
			debugLog("Error validating domain URLs: %v", err)
		} else if validURL != "" {
			log.Printf("Found valid signup URL from domain: %s", validURL)
			return validURL, nil
		}

		// If direct validation fails, try to generate common tournament subdomains
		debugLog("Trying common tournament subdomains for %d domains", len(domains))
		for _, domain := range domains {
			tournoiSubdomains := generateCommonTournamentSubdomains(domain)
			for _, subdomain := range tournoiSubdomains {
				// Clean up the URL
				subdomain = ensureURLProtocol(subdomain)
				debugLog("Validating generated subdomain: %s", subdomain)
				
				// Validate this URL
				validURL, err := ValidateSignupURL(subdomain, tournament, tournamentDate, browserContext)
				if err != nil {
					debugLog("Error validating subdomain %s: %v", subdomain, err)
					continue
				}
				
				if validURL != "" {
					log.Printf("Found valid signup URL from generated subdomain: %s", validURL)
					return validURL, nil
				}
			}
		}
	}

	// No valid signup URL found
	debugLog("No valid signup URL found in PDF")
	return "", fmt.Errorf("no valid signup URL found in PDF")
}

// findDomainOnlyReferences searches for domain-only references in PDF text
// Examples: "inscriptions sur cctt.fr" or "INSCRIPTIONS SUR LE SITE DU CLUB cctt.fr"
func findDomainOnlyReferences(text string) []string {
	// Domain regex pattern to find standalone domains
	domainRegex := regexp.MustCompile(`\b([a-zA-Z0-9][-a-zA-Z0-9]*\.)+[a-zA-Z]{2,}\b`)
	
	var domainURLs []string
	
	// First, look for domains near registration keywords
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		lineLower := strings.ToLower(line)
		
		// Check if line contains registration keywords
		containsKeyword := false
		for _, keyword := range registrationKeywords {
			if strings.Contains(lineLower, keyword) {
				containsKeyword = true
				break
			}
		}
		
		// Only process lines that contain registration keywords
		if containsKeyword {
			// Find domain references in this line
			domains := domainRegex.FindAllString(line, -1)
			for _, domain := range domains {
				// Convert domain to URL format
				url := ensureURLProtocol(domain)
				if !Contains(domainURLs, url) {
					debugLog("Found domain reference in registration context: %s -> %s", domain, url)
					domainURLs = append(domainURLs, url)
				}
			}
		}
	}
	
	// Also look for specific phrases like "inscription(s) sur (le site) X.com"
	inscriptionSiteRegex := regexp.MustCompile(`(?i)inscriptions?\s+sur(?:\s+le\s+site(?:\s+du\s+club)?)?[\s:]+([a-zA-Z0-9][-a-zA-Z0-9]*\.)+[a-zA-Z]{2,}`)
	matches := inscriptionSiteRegex.FindAllStringSubmatch(text, -1)
	for _, match := range matches {
		if len(match) > 0 {
			// Extract the domain from the matched text
			domainPart := match[0]
			domains := domainRegex.FindAllString(domainPart, -1)
			for _, domain := range domains {
				url := ensureURLProtocol(domain)
				if !Contains(domainURLs, url) {
					debugLog("Found explicit inscription site reference: %s -> %s", domain, url)
					domainURLs = append(domainURLs, url)
				}
			}
		}
	}

	// Also look for cases like "A PRIVILEGIER : INSCRIPTIONS SUR LE SITE DU CLUB cctt.fr"
	privilegierRegex := regexp.MustCompile(`(?i)(?:a\s+privilegier\s*[:]\s*)?inscriptions?\s+sur\s+(?:le\s+)?site\s+(?:du\s+)?club\s+([a-zA-Z0-9][-a-zA-Z0-9]*\.)+[a-zA-Z]{2,}`)
	privilegierMatches := privilegierRegex.FindAllStringSubmatch(text, -1)
	for _, match := range privilegierMatches {
		if len(match) > 0 {
			// Extract the domain from the matched text
			domainPart := match[0]
			debugLog("Found pattern match: %s", domainPart)
			domains := domainRegex.FindAllString(domainPart, -1)
			for _, domain := range domains {
				url := ensureURLProtocol(domain)
				if !Contains(domainURLs, url) {
					debugLog("Found 'A PRIVILEGIER' inscription site reference: %s -> %s", domain, url)
					// This was found in an "A PRIVILEGIER" context, so give it higher priority by adding it first
					domainURLs = append([]string{url}, domainURLs...)
				}
			}
		}
	}
	
	return domainURLs
}

// ensureURLProtocol ensures a URL has a protocol prefix
func ensureURLProtocol(url string) string {
	url = strings.TrimSpace(url)
	if !strings.HasPrefix(strings.ToLower(url), "http://") && 
	   !strings.HasPrefix(strings.ToLower(url), "https://") {
		// Default to https, will fall back to http if needed
		return "https://" + url
	}
	return url
}

// generateCommonTournamentSubdomains creates a list of possible subdomains to check
// for a domain found in a PDF that may refer to tournament registration
func generateCommonTournamentSubdomains(domain string) []string {
	// Strip any protocol if present
	domain = strings.TrimPrefix(strings.TrimPrefix(domain, "https://"), "http://")
	
	// Common subdomains for tournament registration sites
	commonSubdomains := []string{
		"tournoi", "inscription", "inscriptions", "competition",
		"competitions", "register", "signup", "toornament", "tournament",
	}
	
	results := []string{}
	
	// Add the base domain
	results = append(results, "https://"+domain)
	
	// Add common tournament-related subdomains
	for _, subdomain := range commonSubdomains {
		results = append(results, "https://"+subdomain+"."+domain)
	}
	
	return results
}

// tryRecursiveFormNavigation tries to find a registration form by recursively following links
func tryRecursiveFormNavigation(urls []string, tournament cache.TournamentCache, tournamentDate time.Time, browserContext pw.BrowserContext) (string, bool, error) {
	// Expand the URLs list with potential subdomains to check
	expandedUrls := []string{}
	
	for _, url := range urls {
		// Add the original URL
		expandedUrls = append(expandedUrls, url)
		
		// Extract domain and generate subdomains to try
		domain := strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "http://")
		// Only generate subdomains if this looks like a base domain (no path components)
		if !strings.Contains(domain, "/") {
			subdomains := generateCommonTournamentSubdomains(domain)
			// Skip the first one which is just the original domain with https
			if len(subdomains) > 1 {
				for _, subdomain := range subdomains[1:] {
					if !Contains(expandedUrls, subdomain) {
						expandedUrls = append(expandedUrls, subdomain)
					}
				}
			}
		}
	}
	
	debugLog("Expanded %d URLs to %d URLs including potential subdomains", 
		len(urls), len(expandedUrls))
	
	// Try each URL (original and subdomains) with recursive navigation
	for _, url := range expandedUrls {
		debugLog("Starting recursive form navigation from: %s", url)
		finalURL, found, err := findRegistrationFormRecursively(url, tournament, tournamentDate, browserContext, 0)
		if err != nil {
			debugLog("Error in recursive form navigation for %s: %v", url, err)
			continue
		}
		
		if found {
			log.Printf("Found valid signup URL via recursive navigation: %s", finalURL)
			return finalURL, true, nil
		}
	}
	
	return "", false, nil
}

// findRegistrationFormRecursively recursively navigates through pages looking for a registration form
func findRegistrationFormRecursively(url string, tournament cache.TournamentCache, tournamentDate time.Time, browserContext pw.BrowserContext, depth int) (string, bool, error) {
	// Check recursion depth limit
	if depth >= maxRedirections {
		debugLog("Reached maximum recursion depth (%d) for URL: %s", maxRedirections, url)
		return "", false, nil
	}
	
	debugLog("Checking for registration form at depth %d: %s", depth, url)
	
	// First, check if the current URL is a valid signup form directly
	validURL, err := ValidateSignupURL(url, tournament, tournamentDate, browserContext)
	if err != nil {
		return "", false, fmt.Errorf("error validating URL %s: %w", url, err)
	}
	
	if validURL != "" {
		debugLog("Found valid signup form at depth %d: %s", depth, validURL)
		return validURL, true, nil
	}
	
	// If not a valid form, look for registration-related links on this page
	page, err := browserContext.NewPage()
	if err != nil {
		return "", false, fmt.Errorf("failed to create new page: %w", err)
	}
	defer page.Close()
	
	// First try with HTTPS
	currentURL := url
	resp, err := page.Goto(currentURL, pw.PageGotoOptions{
		Timeout:   pw.Float(30000),
		WaitUntil: pw.WaitUntilStateNetworkidle,
	})
	
	// If HTTPS failed and URL was using https://, retry with http://
	if err != nil && strings.HasPrefix(url, "https://") {
		httpURL := "http://" + strings.TrimPrefix(url, "https://")
		debugLog("HTTPS navigation failed, retrying with HTTP: %s", httpURL)
		currentURL = httpURL
		resp, err = page.Goto(httpURL, pw.PageGotoOptions{
			Timeout:   pw.Float(30000),
			WaitUntil: pw.WaitUntilStateNetworkidle,
		})
	}
	
	if err != nil {
		return "", false, fmt.Errorf("failed to navigate to %s: %w", currentURL, err)
	}
	
	// Check if navigation was successful
	if resp == nil || resp.Status() >= 400 {
		status := 0
		if resp != nil {
			status = resp.Status()
		}
		return "", false, fmt.Errorf("failed to navigate to %s, status: %d", currentURL, status)
	}
	
	// Get all links on the page that might be related to registration
	links, err := findRegistrationLinksOnPage(page)
	if err != nil {
		return "", false, fmt.Errorf("failed to find links on page %s: %w", currentURL, err)
	}
	
	debugLog("Found %d potential registration links on page %s", len(links), currentURL)
	
	// Recursively check each registration link
	for _, link := range links {
		debugLog("Following registration link at depth %d: %s", depth, link)
		finalURL, found, err := findRegistrationFormRecursively(link, tournament, tournamentDate, browserContext, depth+1)
		if err != nil {
			debugLog("Error following link %s: %v", link, err)
			continue
		}
		
		if found {
			return finalURL, true, nil
		}
	}
	
	return "", false, nil
}

// findRegistrationLinksOnPage finds links on a page that might lead to registration forms
func findRegistrationLinksOnPage(page pw.Page) ([]string, error) {
	// Registration-related text patterns to look for in links
	registrationTextPatterns := []string{
		"inscription", "register", "signup", "s'inscrire", "formulaire",
		"form", "enregistrement", "participer", "participation",
		"engagement", "engagements",
	}
	
	// Execute JavaScript to find all links and check their text/href
	jsScript := `
	() => {
		const links = Array.from(document.querySelectorAll('a'));
		const regKeywords = ` + fmt.Sprintf("%v", registrationTextPatterns) + `;
		const regLinks = [];
		
		links.forEach(link => {
			const href = link.href;
			if (!href || href.startsWith('javascript:') || href.startsWith('#') || href === '') return;
			
			const text = (link.textContent || '').toLowerCase();
			const title = (link.getAttribute('title') || '').toLowerCase();
			const hrefLower = href.toLowerCase();
			
			// Check if link text, title, or href contains registration keywords
			for (const keyword of regKeywords) {
				if (text.includes(keyword) || title.includes(keyword) || hrefLower.includes(keyword)) {
					regLinks.push(href);
					break;
				}
			}
		});
		
		return regLinks;
	}
	`
	
	result, err := page.Evaluate(jsScript)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate JavaScript: %w", err)
	}
	
	// Convert result to string slice
	links := []string{}
	if result != nil {
		// Cast the result to a slice of interfaces
		if resultArr, ok := result.([]interface{}); ok {
			for _, item := range resultArr {
				if linkStr, ok := item.(string); ok {
					links = append(links, linkStr)
				}
			}
		}
	}
	
	return links, nil
}

// limitURLs limits the number of URLs to process
func limitURLs(urls []string, maxCount int) []string {
	if len(urls) <= maxCount {
		return urls
	}
	return urls[:maxCount]
}

// tryValidateURLs attempts to validate a list of URLs and returns the first valid one
func tryValidateURLs(urls []string, tournament cache.TournamentCache, tournamentDate time.Time, browserContext pw.BrowserContext) (string, bool) {
	// Limit the number of URLs to validate
	urlsToValidate := limitURLs(urls, maxURLsToProcess)
	
	for _, url := range urlsToValidate {
		// Clean up the URL
		cleanURL := ensureURLProtocol(url)
		debugLog("Validating URL: %s", cleanURL)
		
		// Validate this URL
		validURL, err := ValidateSignupURL(cleanURL, tournament, tournamentDate, browserContext)
		if err != nil {
			debugLog("Error validating URL %s: %v", cleanURL, err)
			continue
		}
		
		if validURL != "" {
			return validURL, true
		}
	}
	
	return "", false
}

// findRegistrationURLsInPDF finds URLs in text that might be related to registration
func findRegistrationURLsInPDF(text string) []string {
	// Find all URLs in text using the shared utility function
	allURLs := urlRegex.FindAllString(text, -1)
	var registrationURLs []string

	// Look for URLs near registration keywords
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		line = strings.ToLower(line)

		// Check if line contains registration keywords
		containsKeyword := false
		for _, keyword := range registrationKeywords {
			if strings.Contains(line, keyword) {
				containsKeyword = true
				break
			}
		}

		if containsKeyword {
			// Check current line and surrounding lines for URLs
			startIdx := Max(0, i-2)
			endIdx := Min(len(lines)-1, i+2)

			for j := startIdx; j <= endIdx; j++ {
				urlsInLine := urlRegex.FindAllString(lines[j], -1)
				for _, url := range urlsInLine {
					// Only add unique URLs
					if !Contains(registrationURLs, url) {
						registrationURLs = append(registrationURLs, url)
					}
				}
			}
		}
	}

	// Also include URLs from domains commonly used for registration
	registrationDomains := []string{
		"inscription", "register", "signup", "helloasso", "billetweb", "weezevent",
		"eventbrite", "form", "formulaire",
	}

	for _, url := range allURLs {
		urlLower := strings.ToLower(url)
		for _, domain := range registrationDomains {
			if strings.Contains(urlLower, domain) && !Contains(registrationURLs, url) {
				registrationURLs = append(registrationURLs, url)
				break
			}
		}
	}

	return registrationURLs
}

// findDomainsInText extracts all domain references from the text
func findDomainsInText(text string) []string {
	// Regular expression to find domain names in text
	domainRegex := regexp.MustCompile(`\b(?:https?:\/\/)?(?:www\.)?([a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+\.[a-zA-Z0-9-.]+|[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+)\b`)
	
	matches := domainRegex.FindAllStringSubmatch(text, -1)
	
	// Extract the domains from the regex matches
	var domains []string
	seen := make(map[string]bool) // To avoid duplicates
	
	for _, match := range matches {
		if len(match) >= 2 {
			domain := strings.ToLower(match[1])
			
			// Skip common domains that are unlikely to be tournament sites
			if strings.Contains(domain, "google.com") || 
			   strings.Contains(domain, "gmail.com") ||
			   strings.Contains(domain, "outlook.com") ||
			   strings.Contains(domain, "hotmail.com") {
				continue
			}
			
			// Skip if already seen
			if seen[domain] {
				continue
			}
			
			seen[domain] = true
			domains = append(domains, domain)
		}
	}
	
	return domains
}

// validateDomainURLs attempts to validate URLs generated from domain references
func validateDomainURLs(domains []string, tournament cache.TournamentCache, tournamentDate time.Time, browserContext pw.BrowserContext) (string, error) {
	for _, domain := range domains {
		// Ensure the domain has a protocol
		url := ensureURLProtocol(domain)
		
		// Check if it's a valid URL
		debugLog("Validating domain URL: %s", url)
		
		// Try validating with the main domain
		validURL, err := ValidateSignupURL(url, tournament, tournamentDate, browserContext)
		if err != nil {
			debugLog("Error validating domain URL %s: %v", url, err)
			continue
		}
		
		if validURL != "" {
			return validURL, nil
		}
		
		// If the main domain doesn't work, try with common paths
		commonPaths := []string{
			"/tournoi/",
			"/tournament/",
			"/club/",
			"/events/",
			"/evenements/",
			"/competitions/",
		}
		
		for _, path := range commonPaths {
			pathURL := strings.TrimSuffix(url, "/") + path
			debugLog("Validating domain URL with path: %s", pathURL)
			
			validURL, err := ValidateSignupURL(pathURL, tournament, tournamentDate, browserContext)
			if err != nil {
				debugLog("Error validating domain URL with path %s: %v", pathURL, err)
				continue
			}
			
			if validURL != "" {
				return validURL, nil
			}
		}
	}
	
	return "", fmt.Errorf("no valid signup URL found from domains")
}
