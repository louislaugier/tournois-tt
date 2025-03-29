package finder

import (
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// GetSignupLinkFromHTML searches the HTML content for links that look like signup links for table tennis tournaments
// It accepts the HTML content and the current page URL to avoid returning that URL as a match
func GetSignupLinkFromHTML(HTML string, currentPageURL string) *string {
	// Add debug logging for current page URL
	log.Printf("GetSignupLinkFromHTML analyzing page with URL: %s", currentPageURL)

	// Common French signup/registration related terms
	signupTerms := []string{
		"créer un compte",
		"creer un compte",
		"inscription",
		"s'inscrire",
		"enregistrement",
		"participer",
		"participation",
		"tournoi",
		"compétition",
		"competition",
		"engager",
		"engagement",
		"engagements",
		"se connecter",             // Many sites require login first
		"formulaire",               // Form
		"formulaire d'inscription", // Form
		"formulaire inscription",   // Form
	}

	// Get current year and next year for tournament-specific links
	currentYear := time.Now().Year()
	nextYear := currentYear + 1
	yearTerms := []string{
		"inscription tournoi",
		"inscription " + strconv.Itoa(currentYear),
		"inscription " + strconv.Itoa(nextYear),
		"tournoi " + strconv.Itoa(currentYear),
		"tournoi " + strconv.Itoa(nextYear),
		"ping " + strconv.Itoa(currentYear), // Ping is often used for table tennis in French
		"tennis de table " + strconv.Itoa(currentYear),
	}

	// French seasons for tournament terms (with and without accents)
	seasons := map[string]string{
		"printemps": "printemps", // No accent
		"été":       "ete",
		"automne":   "automne", // No accent
		"hiver":     "hiver",   // No accent
	}
	seasonTerms := []string{}
	for accented, nonAccented := range seasons {
		// Add version with accent
		seasonTerms = append(seasonTerms, "tournoi de "+accented)
		seasonTerms = append(seasonTerms, "tournoi d'"+accented)
		seasonTerms = append(seasonTerms, "tournoi "+accented) // Direct without preposition

		// Add version without accent
		if accented != nonAccented {
			seasonTerms = append(seasonTerms, "tournoi de "+nonAccented)
			seasonTerms = append(seasonTerms, "tournoi d'"+nonAccented)
			seasonTerms = append(seasonTerms, "tournoi "+nonAccented) // Direct without preposition
		}
	}

	// French months for tournament terms (with and without accents)
	months := map[string]string{
		"janvier":   "janvier", // No accent
		"février":   "fevrier",
		"mars":      "mars",    // No accent
		"avril":     "avril",   // No accent
		"mai":       "mai",     // No accent
		"juin":      "juin",    // No accent
		"juillet":   "juillet", // No accent
		"août":      "aout",
		"septembre": "septembre", // No accent
		"octobre":   "octobre",   // No accent
		"novembre":  "novembre",  // No accent
		"décembre":  "decembre",
	}
	monthTerms := []string{}
	for accented, nonAccented := range months {
		// For accented version
		if strings.HasPrefix(accented, "a") || strings.HasPrefix(accented, "o") ||
			strings.HasPrefix(accented, "é") || strings.HasPrefix(accented, "â") {
			monthTerms = append(monthTerms, "tournoi d'"+accented)
		} else {
			monthTerms = append(monthTerms, "tournoi de "+accented)
		}
		monthTerms = append(monthTerms, "tournoi "+accented) // Direct without preposition

		// For non-accented version (only if different)
		if accented != nonAccented {
			if strings.HasPrefix(nonAccented, "a") || strings.HasPrefix(nonAccented, "o") {
				monthTerms = append(monthTerms, "tournoi d'"+nonAccented)
			} else {
				monthTerms = append(monthTerms, "tournoi de "+nonAccented)
			}
			monthTerms = append(monthTerms, "tournoi "+nonAccented) // Direct without preposition
		}
	}

	// Combine all terms to check
	allTerms := append(signupTerms, yearTerms...)
	allTerms = append(allTerms, seasonTerms...)
	allTerms = append(allTerms, monthTerms...)

	// Try finding regular links (<a> tags)
	if href := findSignupLinkFromATags(HTML, allTerms, currentPageURL); href != nil {
		// Double-check before returning to make sure it's not the current page
		if isSameOrFragmentURL(*href, currentPageURL) {
			log.Printf("WARNING: findSignupLinkFromATags returned current page URL: %s - this should not happen!", *href)
			// Continue to next method rather than returning
		} else {
			log.Printf("GetSignupLinkFromHTML returning link from <a> tags: %s", *href)
			return href
		}
	}

	// Try finding buttons
	if href := findSignupLinkFromButtons(HTML, allTerms, currentPageURL); href != nil {
		// Double-check before returning to make sure it's not the current page
		if isSameOrFragmentURL(*href, currentPageURL) {
			log.Printf("WARNING: findSignupLinkFromButtons returned current page URL: %s - this should not happen!", *href)
			// Continue to next method rather than returning
		} else {
			log.Printf("GetSignupLinkFromHTML returning link from <button> tags: %s", *href)
			return href
		}
	}

	// Try finding elements with onclick attributes
	if href := findSignupLinkFromOnclickElements(HTML, allTerms, currentPageURL); href != nil {
		// Double-check before returning to make sure it's not the current page
		if isSameOrFragmentURL(*href, currentPageURL) {
			log.Printf("WARNING: findSignupLinkFromOnclickElements returned current page URL: %s - this should not happen!", *href)
			// Don't return same-page URLs
		} else {
			log.Printf("GetSignupLinkFromHTML returning link from onclick elements: %s", *href)
			return href
		}
	}

	log.Printf("GetSignupLinkFromHTML found no signup links on page: %s", currentPageURL)
	return nil
}

// Find signup links in <a> tags
func findSignupLinkFromATags(HTML string, terms []string, currentPageURL string) *string {
	// Find all links in the HTML
	linkRegex := regexp.MustCompile(`<a\s+[^>]*href=["']([^"']+)["'][^>]*>(.*?)</a>`)
	matches := linkRegex.FindAllStringSubmatch(HTML, -1)

	log.Printf("findSignupLinkFromATags: found %d <a> tags to analyze", len(matches))

	matchCount := 0
	for _, match := range matches {
		if len(match) >= 3 {
			href := match[1]
			matchCount++

			// Log every 50th link to avoid excessive logging
			if matchCount%50 == 0 {
				log.Printf("Processing link #%d: %s", matchCount, href)
			}

			// Skip if this is the current page URL or a fragment of the current page
			if isSameOrFragmentURL(href, currentPageURL) {
				continue
			}

			linkText := strings.ToLower(match[2])

			// Remove HTML tags from link text for better matching
			linkText = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(linkText, "")
			linkText = strings.TrimSpace(linkText)

			// Only log non-empty link text for readability
			if linkText != "" {
				log.Printf("Link #%d text: '%s', URL: %s", matchCount, linkText, href)
			}

			// Check if link text contains any of the signup terms
			for _, term := range terms {
				if strings.Contains(linkText, strings.ToLower(term)) {
					log.Printf("Found potential signup link: %s with text: %s (matched term: %s)", href, linkText, term)
					return &href
				}
			}

			// Also check if the URL itself contains signup-related terms
			hrefLower := strings.ToLower(href)
			for _, term := range terms {
				if containsTermInUrl(hrefLower, term) {
					log.Printf("Found potential signup link in URL: %s (matched term: %s in URL)", href, term)
					return &href
				}
			}
		}
	}
	log.Printf("findSignupLinkFromATags: no matching links found among %d links", matchCount)
	return nil
}

// Find signup links in <button> tags
func findSignupLinkFromButtons(HTML string, terms []string, currentPageURL string) *string {
	// Find all buttons with clickable content
	buttonRegex := regexp.MustCompile(`<button\s+[^>]*>(.*?)</button>`)
	matches := buttonRegex.FindAllStringSubmatch(HTML, -1)

	for _, match := range matches {
		if len(match) >= 2 {
			buttonText := strings.ToLower(match[1])

			// Remove HTML tags from button text
			buttonText = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(buttonText, "")
			buttonText = strings.TrimSpace(buttonText)

			// Check if button text contains any signup terms
			for _, term := range terms {
				if strings.Contains(buttonText, strings.ToLower(term)) {
					// Extract any href or data-href attributes
					hrefRegex := regexp.MustCompile(`<button\s+[^>]*(?:data-href|data-url|data-link)=["']([^"']+)["'][^>]*>`)
					hrefMatch := hrefRegex.FindStringSubmatch(match[0])

					if len(hrefMatch) >= 2 {
						href := hrefMatch[1]

						// Skip if this is the current page URL
						if isSameOrFragmentURL(href, currentPageURL) {
							continue
						}

						log.Printf("Found potential signup button with href: %s and text: %s", href, buttonText)
						return &href
					}

					// If no href attribute, look for ID to link it with potential JavaScript
					idRegex := regexp.MustCompile(`<button\s+[^>]*id=["']([^"']+)["'][^>]*>`)
					idMatch := idRegex.FindStringSubmatch(match[0])

					if len(idMatch) >= 2 {
						buttonID := "#" + idMatch[1] // Return as a CSS selector
						log.Printf("Found potential signup button with ID: %s and text: %s", buttonID, buttonText)
						return &buttonID
					}

					// Return a placeholder for a button with no ID or href
					placeholder := "#button_with_text:" + buttonText
					log.Printf("Found potential signup button with text: %s but no direct link", buttonText)
					return &placeholder
				}
			}
		}
	}
	return nil
}

// Find signup links in elements with onclick attributes
func findSignupLinkFromOnclickElements(HTML string, terms []string, currentPageURL string) *string {
	// Find all elements with onclick attributes
	onclickRegex := regexp.MustCompile(`<([a-z0-9]+)\s+[^>]*onclick=["']([^"']+)["'][^>]*>(.*?)</\1>`)
	matches := onclickRegex.FindAllStringSubmatch(HTML, -1)

	for _, match := range matches {
		if len(match) >= 4 {
			// Element tag is captured but not needed
			onclickValue := match[2] // Onclick JavaScript
			elementText := match[3]  // Element text content

			// Remove HTML tags from element text
			elementText = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(elementText, "")
			elementText = strings.TrimSpace(strings.ToLower(elementText))

			// Check if element text contains signup terms
			containsSignupTerm := false
			for _, term := range terms {
				if strings.Contains(elementText, strings.ToLower(term)) {
					containsSignupTerm = true
					break
				}
			}

			if containsSignupTerm {
				// Try to extract URL from onclick JavaScript
				urlRegex := regexp.MustCompile(`(?:window\.location|location\.href)\s*=\s*["']([^"']+)["']`)
				urlMatch := urlRegex.FindStringSubmatch(onclickValue)

				if len(urlMatch) >= 2 {
					href := urlMatch[1]

					// Skip if this is the current page URL
					if isSameOrFragmentURL(href, currentPageURL) {
						continue
					}

					log.Printf("Found potential signup onclick element with URL: %s", href)
					return &href
				}

				// Return the onclick value as a placeholder
				placeholder := "#onclick_element:" + onclickValue
				log.Printf("Found potential signup onclick element: %s", placeholder)
				return &placeholder
			}
		}
	}
	return nil
}

// Helper function to check if a URL contains a term in various formats
func containsTermInUrl(url string, term string) bool {
	normalizedTerm := strings.ToLower(term)
	return strings.Contains(url, strings.Replace(normalizedTerm, " ", "", -1)) ||
		strings.Contains(url, strings.Replace(normalizedTerm, " ", "-", -1)) ||
		strings.Contains(url, strings.Replace(normalizedTerm, " ", "_", -1))
}

// Helper function to check if a URL is the same as or a fragment of the current page URL
func isSameOrFragmentURL(url string, currentPageURL string) bool {
	// Add debug logging
	defer func() {
		// This will be executed when the function returns
		// We'll log both URLs to see what's being compared
		log.Printf("Comparing URLs - Found URL: %s, Current page URL: %s", url, currentPageURL)
	}()

	// Specially handle javascript: links
	if strings.HasPrefix(strings.ToLower(url), "javascript:") {
		// Extract any actual URL from the JavaScript if possible
		jsURL := extractURLFromJavaScriptLink(url)
		if jsURL != "" {
			// Compare the extracted URL instead
			log.Printf("Extracted URL %s from JavaScript link %s", jsURL, url)
			return isSameOrFragmentURL(jsURL, currentPageURL)
		}

		// If we can't extract a URL but it's a JavaScript event with no clear navigation,
		// we won't consider it a same-page link
		log.Printf("JavaScript link %s doesn't contain a clear navigation URL", url)
		return false
	}

	// Normalize URLs for comparison
	normalizedURL := normalizeURL(url)
	normalizedCurrentPageURL := normalizeURL(currentPageURL)

	// Check for empty URLs
	if normalizedURL == "" {
		log.Printf("Skipping empty URL")
		return true
	}

	// Check for fragment-only URLs (starting with #)
	if strings.HasPrefix(normalizedURL, "#") {
		log.Printf("Skipping fragment-only URL: %s", url)
		return true
	}

	// Check for self-references (./ or /)
	if normalizedURL == "./" || normalizedURL == "/" || normalizedURL == "." {
		log.Printf("Skipping self-reference URL: %s", url)
		return true
	}

	// Direct comparison of normalized URLs
	if normalizedURL == normalizedCurrentPageURL {
		log.Printf("URL matches exactly: %s == %s", normalizedURL, normalizedCurrentPageURL)
		return true
	}

	// Check for relative vs absolute URL variations
	if !strings.HasPrefix(normalizedURL, "http://") && !strings.HasPrefix(normalizedURL, "https://") {
		// This is a relative URL; try to resolve it against the current page

		// Handle "./path" format
		relativeURL := strings.TrimPrefix(normalizedURL, "./")

		// Handle query parameters
		currentURLBase := normalizedCurrentPageURL
		if strings.Contains(currentURLBase, "?") {
			currentURLBase = strings.Split(currentURLBase, "?")[0]
		}

		// Check if current page ends with the relative URL
		if strings.HasSuffix(currentURLBase, relativeURL) {
			log.Printf("Relative URL resolves to current page: %s (base: %s)", relativeURL, currentURLBase)
			return true
		}

		// Try joining the URLs (very simplified URL resolution)
		// Extract base path of current page URL
		var basePath string
		if strings.HasPrefix(currentURLBase, "http://") || strings.HasPrefix(currentURLBase, "https://") {
			// Extract domain and path
			parts := strings.SplitN(currentURLBase, "/", 4)
			if len(parts) >= 4 {
				// For http://domain.com/path
				// parts[0] = "http:", parts[1] = "", parts[2] = "domain.com", parts[3] = "path"
				basePath = parts[0] + "//" + parts[2] + "/" // Reconstruct base URL

				// Handle paths that don't end with slash
				lastPathComponent := parts[3]
				if lastPathComponent != "" && !strings.HasSuffix(lastPathComponent, "/") {
					// Remove the last path component if it doesn't end with a slash
					// (assuming it's a file or endpoint, not a directory)
					lastSlashIndex := strings.LastIndex(lastPathComponent, "/")
					if lastSlashIndex >= 0 {
						basePath += lastPathComponent[:lastSlashIndex+1]
					}
				} else {
					basePath += parts[3]
				}
			} else {
				// Handle http://domain.com with no path
				basePath = currentURLBase
				if !strings.HasSuffix(basePath, "/") {
					basePath += "/"
				}
			}
		} else {
			// If currentURLBase doesn't have a protocol, treat it as a path
			lastSlashIndex := strings.LastIndex(currentURLBase, "/")
			if lastSlashIndex >= 0 {
				basePath = currentURLBase[:lastSlashIndex+1]
			} else {
				basePath = currentURLBase
				if !strings.HasSuffix(basePath, "/") {
					basePath += "/"
				}
			}
		}

		// Combine base path with relative URL
		possibleResolvedURL := basePath + relativeURL
		log.Printf("Comparing possible resolved URL: %s with current page: %s", possibleResolvedURL, normalizedCurrentPageURL)

		if normalizeURL(possibleResolvedURL) == normalizedCurrentPageURL {
			log.Printf("Resolved relative URL matches current page")
			return true
		}
	}

	// Additional check for absolute URLs with different protocols
	// Remove protocol for comparison
	noProtocolURL := strings.TrimPrefix(strings.TrimPrefix(normalizedURL, "https://"), "http://")
	noProtocolCurrentURL := strings.TrimPrefix(strings.TrimPrefix(normalizedCurrentPageURL, "https://"), "http://")

	if noProtocolURL == noProtocolCurrentURL {
		log.Printf("URLs match when ignoring protocol: %s ~ %s", url, currentPageURL)
		return true
	}

	// Handle potential SPA route paths (common in frameworks like React, Vue, Angular)
	// Extract domain parts for comparison
	currentDomain := extractDomain(currentPageURL)
	urlDomain := extractDomain(url)

	// If the domains match and the URL starts with a path like /signup or /register
	if currentDomain != "" && urlDomain != "" && currentDomain == urlDomain {
		// Common SPA paths that indicate it's likely a different view
		potentialSPARoutes := []string{
			"/signup", "/register", "/inscription", "/create-account",
			"/formulaire", "/form", "/enroll", "/join", "/participate",
		}

		// Extract path component
		urlPath := extractPathFromUrl(url)

		// Check if it's a significant different path that indicates a new view
		for _, route := range potentialSPARoutes {
			if strings.HasPrefix(urlPath, route) {
				log.Printf("URL %s appears to be a SPA route to a signup form, not considering it same page", url)
				return false
			}
		}

		// If none of the special routes match, fallback to the regular comparison
	}

	log.Printf("URLs are different: %s ≠ %s", url, currentPageURL)
	return false
}

// Helper function to normalize a URL for comparison
func normalizeURL(url string) string {
	// Convert to lowercase
	normalized := strings.ToLower(url)

	// Remove trailing slash if present
	if strings.HasSuffix(normalized, "/") {
		normalized = normalized[:len(normalized)-1]
	}

	// Remove anchor fragment if present
	if hashIndex := strings.Index(normalized, "#"); hashIndex >= 0 {
		normalized = normalized[:hashIndex]
	}

	// Remove query parameters for base URL comparison
	if questionIndex := strings.Index(normalized, "?"); questionIndex >= 0 {
		normalized = normalized[:questionIndex]
	}

	return normalized
}

// Extract a domain from a URL
func extractDomain(url string) string {
	// Handle URLs with protocol
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		parts := strings.SplitN(url, "/", 4)
		if len(parts) >= 3 {
			return parts[2] // Domain is the third part
		}
	}
	return ""
}

// Extract just the path component from a URL
func extractPathFromUrl(url string) string {
	// Handle URLs with protocol
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		parts := strings.SplitN(url, "/", 4)
		if len(parts) >= 4 {
			return "/" + parts[3] // Path is the fourth part, add leading slash
		}
		return "/"
	}

	// For paths without protocol
	if strings.HasPrefix(url, "/") {
		return url
	}

	return ""
}

// Extract a URL from a JavaScript link like "javascript:window.location='https://example.com'"
func extractURLFromJavaScriptLink(jsLink string) string {
	// Common patterns in JavaScript navigation
	patterns := []string{
		`window\.location(?:\.href)?\s*=\s*["']([^"']+)["']`,
		`location\.href\s*=\s*["']([^"']+)["']`,
		`window\.open\(["']([^"']+)["']`,
		`location\.replace\(["']([^"']+)["']`,
		`navigate\(["']([^"']+)["']`,
		`navigateTo\(["']([^"']+)["']`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(jsLink)
		if len(matches) >= 2 {
			return matches[1]
		}
	}

	return ""
}

func IsSignupFormPage(HTML string) bool {
	// Check if the HTML contains a form
	containsForm := strings.Contains(strings.ToLower(HTML), "<form")

	// Check if nom/prénom are in form labels or inputs
	nameFieldsInFormElements := false

	// Check for label elements with text content containing nom/prénom
	labelRegex := regexp.MustCompile(`<label[^>]*>([^<]*(?:nom|prénom|prenom)[^<]*)</label>`)
	if labelRegex.MatchString(HTML) {
		nameFieldsInFormElements = true
	}

	// Check for input elements with name/id/placeholder attributes containing nom/prénom
	inputRegex := regexp.MustCompile(`<input[^>]*(?:name|id|placeholder)=["']([^"']*(?:nom|prénom|prenom)[^"']*)["'][^>]*>`)
	if inputRegex.MatchString(HTML) {
		nameFieldsInFormElements = true
	}

	// Check for tableau links
	containsTableauLinks := false

	// Check for tableau + letter
	tableauLetterRegex := regexp.MustCompile(`<a[^>]*>([^<]*tableau\s+[A-Za-z][^<]*)</a>`)

	// Check for tableau + number
	tableauNumberRegex := regexp.MustCompile(`<a[^>]*>([^<]*tableau\s+\d+[^<]*)</a>`)

	// Check for tableau + classement (all possible variations)
	classementPatterns := []string{
		`tableau\s+\d{3}`,                   // tableau 599, tableau 600, etc.
		`tableau\s+-\d{3}`,                  // tableau -599, tableau -600, etc.
		`tableau\s+\d{3}\s*pts`,             // tableau 599 pts, tableau 600 pts, etc.
		`tableau\s+-\d{3}\s*pts`,            // tableau -599 pts, tableau -600 pts, etc.
		`tableau\s+\d{3}\s*points`,          // tableau 599 points, tableau 600 points, etc.
		`tableau\s+-\d{3}\s*points`,         // tableau -599 points, tableau -600 points, etc.
		`tableau\s+-\s*de\s+\d{3}\s*pts`,    // tableau - de 599 pts, tableau - de 600 pts, etc.
		`tableau\s+-\s*de\s+\d{3}\s*points`, // tableau - de 599 points, tableau - de 600 points, etc.
	}

	classementRegexPattern := strings.Join(classementPatterns, "|")
	tableauClassementRegex := regexp.MustCompile(`<a[^>]*>([^<]*(?:` + classementRegexPattern + `)[^<]*)</a>`)

	// Also check for child elements within links
	childElementsRegexPattern := `<a[^>]*>.*?(?:` +
		`tableau\s+[A-Za-z]|` +
		`tableau\s+\d+|` +
		classementRegexPattern +
		`).*?</a>`
	childElementsRegex := regexp.MustCompile(childElementsRegexPattern)

	// Check if any of the tableau patterns match
	if tableauLetterRegex.MatchString(HTML) ||
		tableauNumberRegex.MatchString(HTML) ||
		tableauClassementRegex.MatchString(HTML) ||
		childElementsRegex.MatchString(HTML) {
		containsTableauLinks = true
	}

	// Return true if the page contains a form and either name fields or tableau links
	return containsForm && (nameFieldsInFormElements || containsTableauLinks)
}
