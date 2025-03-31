package finder

import (
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// GetSignupLinkFromHTML searches the HTML content for links that look like signup links for table tennis tournaments
// It accepts the HTML content and the current page URL to avoid returning that URL as a match
func GetSignupLinkFromHTML(HTML string, currentPageURL string) *string {
	log.Printf("Analyzing page for signup links: %s", currentPageURL)

	// Log a truncated version of the HTML to aid debugging (first 1000 chars and last 1000 chars)
	htmlLen := len(HTML)
	if htmlLen > 2000 {
		log.Printf("HTML content (truncated) - First 1000 chars:\n%s\n...and last 1000 chars:\n%s",
			HTML[:1000], HTML[htmlLen-1000:])
	} else {
		log.Printf("HTML content (full):\n%s", HTML)
	}

	// Parse the URL to get path segments for loop prevention
	var currentPath string
	if strings.HasPrefix(currentPageURL, "http://") || strings.HasPrefix(currentPageURL, "https://") {
		parts := strings.SplitN(currentPageURL, "/", 4)
		if len(parts) >= 4 {
			currentPath = parts[3]
		}
	}
	log.Printf("Current path: %s", currentPath)

	// Common terms to identify signup links - in French and English
	signupTerms := []string{
		"créer un compte",
		"creer un compte",
		"inscription",
		"inscriptions",
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
		"open",                     // Many tournaments use "open" in their names
		"sign-up",                  // English version of "inscription"
		"signup",                   // English version without hyphen
		"register",                 // English version "register"
		"registration",             // English noun
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
		"open " + strconv.Itoa(currentYear), // Common pattern for tournaments
		"open " + strconv.Itoa(nextYear),    // Common pattern for tournaments
		"open-" + strconv.Itoa(currentYear), // Common pattern for tournaments with hyphen
		"open-" + strconv.Itoa(nextYear),    // Common pattern for tournaments with hyphen
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

	// Store candidate links based on priority
	type candidateLink struct {
		href                string
		isInternalDomain    bool
		containsInscription bool
		containsTournoi     bool
		containsOpen        bool
		priority            int // Higher number means higher priority
	}

	var candidates []candidateLink

	// Extract domain from current page URL
	currentDomain := extractDomain(currentPageURL)
	log.Printf("Current domain: %s", currentDomain)

	// Check for signup-specific URL patterns
	signupURLPatterns := []string{
		"sign-up",
		"signup",
		"register",
		"registration",
		"inscription",
		"formulaire",
	}

	// The regular expressions to find `a` tags
	aTagsPattern := regexp.MustCompile(`<a\s+[^>]*href=["']([^"']+)["'][^>]*>(.*?)</a>`)
	buttonsWithJSPattern := regexp.MustCompile(`<button\s+[^>]*onclick=["'][^"']*location\.href=["']([^"']+)["'][^>]*>(.*?)</button>`)
	tagsWithOnClickPattern := regexp.MustCompile(`<[^>]+onclick=["'][^"']*window\.location(?:\.href)?\s*=\s*["']([^"']+)["'][^>]*>(.*?)</[^>]+>`)

	// Find all links from <a> tags
	matches := aTagsPattern.FindAllStringSubmatch(HTML, -1)

	// Also check buttons with JavaScript href
	buttonMatches := buttonsWithJSPattern.FindAllStringSubmatch(HTML, -1)
	matches = append(matches, buttonMatches...)

	// And elements with onclick handlers
	onclickMatches := tagsWithOnClickPattern.FindAllStringSubmatch(HTML, -1)
	matches = append(matches, onclickMatches...)

	log.Printf("Found %d links (a tags, buttons, onclick elements) to analyze", len(matches))

	// Special pattern to find /inscription links
	inscriptionPathRegex := regexp.MustCompile(`<a\s+[^>]*href=["']([^"']*(?:/inscription|/inscriptions)[^"']*)["'][^>]*>`)
	inscriptionPathMatches := inscriptionPathRegex.FindAllStringSubmatch(HTML, -1)
	if len(inscriptionPathMatches) > 0 {
		log.Printf("FOUND %d LINKS WITH /INSCRIPTION PATH:", len(inscriptionPathMatches))
		for i, match := range inscriptionPathMatches {
			if len(match) >= 2 {
				href := match[1]
				resolvedHref := resolveURL(href, currentPageURL)
				log.Printf("  /INSCRIPTION LINK #%d: %s (resolved: %s)", i+1, href, resolvedHref)

				// Auto-add as a high-priority candidate
				candidate := candidateLink{
					href:                resolvedHref,
					isInternalDomain:    isInternalURL(href, currentPageURL),
					containsInscription: true,
					priority:            12, // Higher than standard internal+inscription
				}

				// Add to candidates
				candidates = append(candidates, candidate)
				log.Printf("Added high-priority /inscription path link: %s, priority: %d", resolvedHref, 12)
			}
		}
	} else {
		log.Println("NO LINKS WITH /INSCRIPTION PATH FOUND")
	}

	// Also check for "sign-up" specific patterns
	signupPathRegex := regexp.MustCompile(`<a\s+[^>]*href=["']([^"']*(?:/sign-up|/signup|/register|/registration)[^"']*)["'][^>]*>`)
	signupPathMatches := signupPathRegex.FindAllStringSubmatch(HTML, -1)
	if len(signupPathMatches) > 0 {
		log.Printf("FOUND %d LINKS WITH ENGLISH SIGNUP PATH:", len(signupPathMatches))
		for i, match := range signupPathMatches {
			if len(match) >= 2 {
				href := match[1]
				resolvedHref := resolveURL(href, currentPageURL)
				log.Printf("  SIGNUP PATH LINK #%d: %s (resolved: %s)", i+1, href, resolvedHref)

				// Auto-add as a high-priority candidate
				candidate := candidateLink{
					href:                resolvedHref,
					isInternalDomain:    isInternalURL(href, currentPageURL),
					containsInscription: true,
					priority:            11, // High priority, but slightly below /inscription
				}

				// Add to candidates
				candidates = append(candidates, candidate)
				log.Printf("Added high-priority English signup path link: %s, priority: %d", resolvedHref, 11)
			}
		}
	} else {
		log.Println("NO LINKS WITH ENGLISH SIGNUP PATHS FOUND")
	}

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		href := match[1]
		linkText := match[2]

		// Skip if it's a fragment URL, anchor, or the current page
		if isSameOrFragmentURL(href, currentPageURL) {
			continue
		}

		// Resolve relative URLs
		href = resolveURL(href, currentPageURL)

		// Process candidate
		candidate := candidateLink{
			href:             href,
			isInternalDomain: isInternalURL(href, currentPageURL),
			priority:         0,
		}

		// Check for signup-specific URL patterns
		for _, pattern := range signupURLPatterns {
			if strings.Contains(strings.ToLower(href), pattern) {
				log.Printf("Found URL with signup pattern '%s': %s", pattern, href)
				candidate.containsInscription = true
				break
			}
		}

		// Check if the LINK TEXT contains keyword cases (case insensitive)
		linkTextContainsInscription := false
		linkTextContainsTournoi := false
		linkTextContainsOpen := false

		if strings.Contains(strings.ToLower(linkText), "inscription") {
			linkTextContainsInscription = true
			log.Printf("Link text contains 'inscription': %s, text: '%s'", href, linkText)
		}
		if strings.Contains(strings.ToLower(linkText), "tournoi") {
			linkTextContainsTournoi = true
			log.Printf("Link text contains 'tournoi': %s, text: '%s'", href, linkText)
		}
		if strings.Contains(strings.ToLower(linkText), "open") {
			linkTextContainsOpen = true
			log.Printf("Link text contains 'open': %s, text: '%s'", href, linkText)
		}

		// Check if URL contains keywords (case insensitive)
		urlContainsInscription := strings.Contains(strings.ToLower(href), "inscription")
		urlContainsTournoi := strings.Contains(strings.ToLower(href), "tournoi")

		// Special pattern matching for "open-YEAR" in URL
		openRegex := regexp.MustCompile(`(?i)open[-]?\d{4}`)
		urlContainsOpen := strings.Contains(strings.ToLower(href), "open") ||
			openRegex.MatchString(href)

		candidate.containsInscription = linkTextContainsInscription || urlContainsInscription
		candidate.containsTournoi = linkTextContainsTournoi || urlContainsTournoi
		candidate.containsOpen = linkTextContainsOpen || urlContainsOpen

		if linkTextContainsInscription {
			log.Printf("Link with text containing 'inscription': %s", href)
		}

		if urlContainsInscription {
			log.Printf("Link with URL containing 'inscription': %s", href)
		}

		if linkTextContainsTournoi {
			log.Printf("Link with text containing 'tournoi': %s", href)
		}

		if urlContainsTournoi {
			log.Printf("Link with URL containing 'tournoi': %s", href)
		}

		if linkTextContainsOpen {
			log.Printf("Link with text containing 'open': %s", href)
		}

		if urlContainsOpen {
			log.Printf("Link with URL containing 'open': %s", href)
		}

		// Modify priority setting to handle the special path case
		if candidate.isInternalDomain && strings.Contains(strings.ToLower(href), "/inscription") {
			candidate.priority = 12 // Highest priority: internal domain + "/inscription" path
			log.Printf("Setting highest priority (12) for link with /inscription path: %s", href)
		} else if candidate.isInternalDomain && candidate.containsInscription {
			candidate.priority = 10 // High priority: internal domain + "inscription" anywhere
		} else if candidate.isInternalDomain && candidate.containsOpen {
			candidate.priority = 9 // High priority: internal domain + "open" (likely tournament)
		} else if candidate.isInternalDomain && candidate.containsTournoi {
			candidate.priority = 8 // High priority: internal domain + "tournoi"
		} else if candidate.isInternalDomain && strings.Contains(strings.ToLower(href), "open") {
			candidate.priority = 7 // Often sites have /open-YEAR for tournaments
		} else if candidate.isInternalDomain {
			candidate.priority = 5 // Medium-high priority: any internal domain link that matches search terms
		} else if candidate.containsInscription {
			candidate.priority = 3 // Medium priority: external domain + "inscription"
		} else if candidate.containsTournoi || candidate.containsOpen {
			candidate.priority = 2 // Medium-low priority: external domain + "tournoi" or "open"
		} else {
			candidate.priority = 1 // Lowest priority: external domain, no specific terms
		}

		candidates = append(candidates, candidate)
		log.Printf("Found candidate link: %s, priority: %d (internal: %v, inscription: %v, tournoi: %v, open: %v)",
			href, candidate.priority, candidate.isInternalDomain, candidate.containsInscription, candidate.containsTournoi, candidate.containsOpen)
	}

	// Find links from buttons
	buttonLinks := findAllSignupLinksFromButtons(HTML, allTerms, currentPageURL)
	for _, link := range buttonLinks {
		candidate := candidateLink{href: link}

		// Handle relative URLs when determining if it's internal
		resolvedURL := link
		if !strings.HasPrefix(link, "http://") && !strings.HasPrefix(link, "https://") {
			// This is a relative URL, it belongs to the same domain
			candidate.isInternalDomain = true

			// For logs and determining if it contains inscription
			if strings.HasPrefix(link, "/") {
				// Absolute path on same domain
				baseURL := getBaseURL(currentPageURL)
				resolvedURL = baseURL + link
			} else {
				// Relative path
				resolvedURL = currentPageURL
				if !strings.HasSuffix(resolvedURL, "/") {
					// Remove last path component if it doesn't end with /
					lastSlashIdx := strings.LastIndex(resolvedURL, "/")
					if lastSlashIdx > 0 {
						resolvedURL = resolvedURL[:lastSlashIdx+1]
					} else {
						resolvedURL += "/"
					}
				}
				resolvedURL += link
			}
			log.Printf("Resolved relative URL: %s to %s", link, resolvedURL)
		} else {
			// Check if absolute link is on the same domain
			linkDomain := extractDomain(link)
			candidate.isInternalDomain = (linkDomain != "" && currentDomain != "" && linkDomain == currentDomain)
		}

		// Check if URL contains keywords (case insensitive)
		urlContainsInscription := strings.Contains(strings.ToLower(resolvedURL), "inscription")
		urlContainsTournoi := strings.Contains(strings.ToLower(resolvedURL), "tournoi")
		urlContainsOpen := strings.Contains(strings.ToLower(resolvedURL), "open")

		candidate.containsInscription = urlContainsInscription
		candidate.containsTournoi = urlContainsTournoi
		candidate.containsOpen = urlContainsOpen

		if urlContainsInscription {
			log.Printf("Button link with URL containing 'inscription': %s", link)
		}

		if urlContainsTournoi {
			log.Printf("Button link with URL containing 'tournoi': %s", link)
		}

		if urlContainsOpen {
			log.Printf("Button link with URL containing 'open': %s", link)
		}

		// Modify priority setting to handle the special path case
		if candidate.isInternalDomain && strings.Contains(strings.ToLower(resolvedURL), "/inscription") {
			candidate.priority = 12 // Highest priority: internal domain + "/inscription" path
			log.Printf("Setting highest priority (12) for link with /inscription path: %s", resolvedURL)
		} else if candidate.isInternalDomain && candidate.containsInscription {
			candidate.priority = 10 // High priority: internal domain + "inscription" anywhere
		} else if candidate.isInternalDomain && candidate.containsOpen {
			candidate.priority = 9 // High priority: internal domain + "open" (likely tournament)
		} else if candidate.isInternalDomain && candidate.containsTournoi {
			candidate.priority = 8 // High priority: internal domain + "tournoi"
		} else if candidate.isInternalDomain && strings.Contains(strings.ToLower(resolvedURL), "open") {
			candidate.priority = 7 // Often sites have /open-YEAR for tournaments
		} else if candidate.isInternalDomain {
			candidate.priority = 5 // Medium-high priority: any internal domain link that matches search terms
		} else if candidate.containsInscription {
			candidate.priority = 3 // Medium priority: external domain + "inscription"
		} else if candidate.containsTournoi || candidate.containsOpen {
			candidate.priority = 2 // Medium-low priority: external domain + "tournoi" or "open"
		} else {
			candidate.priority = 1 // Lowest priority: external domain, no specific terms
		}

		candidates = append(candidates, candidate)
		log.Printf("Found candidate button link: %s, priority: %d (internal: %v, inscription: %v, tournoi: %v, open: %v)",
			link, candidate.priority, candidate.isInternalDomain, candidate.containsInscription, candidate.containsTournoi, candidate.containsOpen)
	}

	// Find links from onclick elements
	onclickLinks := findAllSignupLinksFromOnclickElements(HTML, allTerms, currentPageURL)
	for _, link := range onclickLinks {
		candidate := candidateLink{href: link}

		// Handle relative URLs when determining if it's internal
		resolvedURL := link
		if !strings.HasPrefix(link, "http://") && !strings.HasPrefix(link, "https://") {
			// This is a relative URL, it belongs to the same domain
			candidate.isInternalDomain = true

			// For logs and determining if it contains inscription
			if strings.HasPrefix(link, "/") {
				// Absolute path on same domain
				baseURL := getBaseURL(currentPageURL)
				resolvedURL = baseURL + link
			} else {
				// Relative path
				resolvedURL = currentPageURL
				if !strings.HasSuffix(resolvedURL, "/") {
					// Remove last path component if it doesn't end with /
					lastSlashIdx := strings.LastIndex(resolvedURL, "/")
					if lastSlashIdx > 0 {
						resolvedURL = resolvedURL[:lastSlashIdx+1]
					} else {
						resolvedURL += "/"
					}
				}
				resolvedURL += link
			}
			log.Printf("Resolved relative URL: %s to %s", link, resolvedURL)
		} else {
			// Check if absolute link is on the same domain
			linkDomain := extractDomain(link)
			candidate.isInternalDomain = (linkDomain != "" && currentDomain != "" && linkDomain == currentDomain)
		}

		// Check if URL contains keywords (case insensitive)
		urlContainsInscription := strings.Contains(strings.ToLower(resolvedURL), "inscription")
		urlContainsTournoi := strings.Contains(strings.ToLower(resolvedURL), "tournoi")
		urlContainsOpen := strings.Contains(strings.ToLower(resolvedURL), "open")

		candidate.containsInscription = urlContainsInscription
		candidate.containsTournoi = urlContainsTournoi
		candidate.containsOpen = urlContainsOpen

		if urlContainsInscription {
			log.Printf("Onclick link with URL containing 'inscription': %s", link)
		}

		if urlContainsTournoi {
			log.Printf("Onclick link with URL containing 'tournoi': %s", link)
		}

		if urlContainsOpen {
			log.Printf("Onclick link with URL containing 'open': %s", link)
		}

		// Modify priority setting to handle the special path case
		if candidate.isInternalDomain && strings.Contains(strings.ToLower(resolvedURL), "/inscription") {
			candidate.priority = 12 // Highest priority: internal domain + "/inscription" path
			log.Printf("Setting highest priority (12) for link with /inscription path: %s", resolvedURL)
		} else if candidate.isInternalDomain && candidate.containsInscription {
			candidate.priority = 10 // High priority: internal domain + "inscription" anywhere
		} else if candidate.isInternalDomain && candidate.containsOpen {
			candidate.priority = 9 // High priority: internal domain + "open" (likely tournament)
		} else if candidate.isInternalDomain && candidate.containsTournoi {
			candidate.priority = 8 // High priority: internal domain + "tournoi"
		} else if candidate.isInternalDomain && strings.Contains(strings.ToLower(resolvedURL), "open") {
			candidate.priority = 7 // Often sites have /open-YEAR for tournaments
		} else if candidate.isInternalDomain {
			candidate.priority = 5 // Medium-high priority: any internal domain link that matches search terms
		} else if candidate.containsInscription {
			candidate.priority = 3 // Medium priority: external domain + "inscription"
		} else if candidate.containsTournoi || candidate.containsOpen {
			candidate.priority = 2 // Medium-low priority: external domain + "tournoi" or "open"
		} else {
			candidate.priority = 1 // Lowest priority: external domain, no specific terms
		}

		candidates = append(candidates, candidate)
		log.Printf("Found candidate onclick link: %s, priority: %d (internal: %v, inscription: %v, tournoi: %v, open: %v)",
			link, candidate.priority, candidate.isInternalDomain, candidate.containsInscription, candidate.containsTournoi, candidate.containsOpen)
	}

	// Special check for link context - examine surrounding text
	linkContextRegex := regexp.MustCompile(`<[^>]*>([^<]{0,50}inscription[^<]{0,50})<[^>]*>.*?<a\s+[^>]*href=["']([^"']+)["'][^>]*>`)
	contextMatches := linkContextRegex.FindAllStringSubmatch(HTML, -1)
	if len(contextMatches) > 0 {
		log.Printf("Found links with inscription context: %d", len(contextMatches))
		for _, match := range contextMatches {
			if len(match) >= 3 {
				context := match[1]
				href := match[2]
				log.Printf("Link with inscription context: href='%s', context: '%s'", href, context)

				// Skip if same as current page or we already have this as a candidate
				alreadyCandidate := false
				for _, c := range candidates {
					if c.href == href {
						alreadyCandidate = true
						// If it's already a candidate, boost its priority
						if !c.containsInscription {
							c.containsInscription = true
							if c.isInternalDomain {
								c.priority = 10 // Boost to highest priority
							} else {
								c.priority = 3 // Boost external link priority
							}
							log.Printf("Boosted priority of existing candidate with inscription context: %s to %d", href, c.priority)
						}
						break
					}
				}

				if !alreadyCandidate && !isSameOrFragmentURL(href, currentPageURL) {
					// Create a new candidate with high priority
					candidate := candidateLink{
						href:                href,
						isInternalDomain:    isInternalURL(href, currentPageURL),
						containsInscription: true,
						priority:            0,
					}

					// Set priority based on internal/external
					if candidate.isInternalDomain {
						candidate.priority = 10 // Highest priority for internal + inscription context
					} else {
						candidate.priority = 3 // Medium priority for external + inscription context
					}

					log.Printf("Added new candidate from inscription context: %s, priority: %d", href, candidate.priority)
					candidates = append(candidates, candidate)
				}
			}
		}
	}

	// Add debug info about all links found
	log.Printf("SUMMARY: Found %d candidate links total", len(candidates))
	for i, c := range candidates {
		log.Printf("Candidate #%d: href=%s, priority=%d, internal=%v, inscription=%v, tournoi=%v, open=%v",
			i+1, c.href, c.priority, c.isInternalDomain, c.containsInscription, c.containsTournoi, c.containsOpen)
	}

	// Filter out links that point to the current page or parent directories to prevent loops
	var filteredCandidates []candidateLink
	for _, c := range candidates {
		// Skip links that are the same as the current page URL
		if normalizeURL(c.href) == normalizeURL(currentPageURL) {
			log.Printf("Filtering out link that points to current page: %s", c.href)
			continue
		}

		// Skip links that point to parent directories
		if currentPath != "" && isInternalURL(c.href, currentPageURL) {
			resolvedURL := resolveURL(c.href, currentPageURL)
			resolvedPath := ""
			if strings.HasPrefix(resolvedURL, "http://") || strings.HasPrefix(resolvedURL, "https://") {
				parts := strings.SplitN(resolvedURL, "/", 4)
				if len(parts) >= 4 {
					resolvedPath = parts[3]
				}
			}

			// Check if resolvedPath is a parent or the same as currentPath
			if resolvedPath != "" && strings.HasPrefix(currentPath, resolvedPath) && resolvedPath != currentPath {
				log.Printf("Filtering out link that points to parent directory: %s -> %s", c.href, resolvedPath)
				continue
			}
		}

		filteredCandidates = append(filteredCandidates, c)
	}

	if len(filteredCandidates) != len(candidates) {
		log.Printf("Filtered out %d links that could cause loops, %d candidates remain",
			len(candidates)-len(filteredCandidates), len(filteredCandidates))
	}

	// If we have filtered candidates, return the highest priority one
	if len(filteredCandidates) > 0 {
		// Sort candidates by priority (descending)
		sort.Slice(filteredCandidates, func(i, j int) bool {
			return filteredCandidates[i].priority > filteredCandidates[j].priority
		})

		bestCandidate := filteredCandidates[0]
		log.Printf("Returning highest priority filtered link: %s (priority: %d)",
			bestCandidate.href, bestCandidate.priority)
		return &bestCandidate.href
	} else if len(candidates) > 0 {
		// Fallback to all candidates if filtering removed everything
		// Sort candidates by priority (descending)
		sort.Slice(candidates, func(i, j int) bool {
			return candidates[i].priority > candidates[j].priority
		})

		bestCandidate := candidates[0]
		log.Printf("Returning highest priority link (no filtered candidates): %s (priority: %d)",
			bestCandidate.href, bestCandidate.priority)
		return &bestCandidate.href
	}

	log.Printf("GetSignupLinkFromHTML found no signup links on page: %s", currentPageURL)
	return nil
}

// Find all potential signup links in <a> tags, returns a slice of URLs
func FindAllSignupLinksFromATags(HTML string, terms []string, currentPageURL string) []string {
	var signupLinks []string
	// Regular expression to find all <a> tags with href attributes
	aTagRegex := regexp.MustCompile(`<a\s+[^>]*href=["']([^"']+)["'][^>]*>(.*?)</a>`)
	matches := aTagRegex.FindAllStringSubmatch(HTML, -1)

	// Debug count of a tags
	log.Printf("FindAllSignupLinksFromATags: found %d <a> tags to analyze", len(matches))
	log.Printf("DEBUG: Listing all links found in the HTML:")

	// Extract the current domain to identify internal vs external links
	currentDomain := extractDomain(currentPageURL)

	for i, match := range matches {
		if len(match) >= 3 {
			href := match[1]
			linkText := trimAndCleanText(match[2])

			// Skip empty links or anchors
			if href == "" || href == "#" {
				continue
			}

			// Determine if this is an internal link
			isInternal := false
			resolvedURL := href

			// Handle relative URLs
			if !strings.HasPrefix(href, "http://") && !strings.HasPrefix(href, "https://") {
				isInternal = true // Relative URLs are internal

				// Resolve the relative URL
				if strings.HasPrefix(href, "/") {
					// Absolute path on same domain
					baseURL := getBaseURL(currentPageURL)
					resolvedURL = baseURL + href
				} else {
					// Relative path
					resolvedURL = currentPageURL
					if !strings.HasSuffix(resolvedURL, "/") {
						// Remove last path component if it doesn't end with /
						lastSlashIdx := strings.LastIndex(resolvedURL, "/")
						if lastSlashIdx > 0 {
							resolvedURL = resolvedURL[:lastSlashIdx+1]
						} else {
							resolvedURL += "/"
						}
					}
					resolvedURL += href
				}
			} else {
				// Check if absolute URL is on the same domain
				linkDomain := extractDomain(href)
				isInternal = (linkDomain != "" && currentDomain != "" && linkDomain == currentDomain)
			}

			// Check if link contains "inscription" (case insensitive)
			hasInscription := false
			if strings.Contains(strings.ToLower(linkText), "inscription") || strings.Contains(strings.ToLower(resolvedURL), "inscription") {
				hasInscription = true
			}

			// Debug info for each link
			log.Printf("Link #%d: href='%s', text='%s', isInternal=%v, hasInscription=%v, resolved=%s",
				i+1, href, linkText, isInternal, hasInscription, resolvedURL)

			// Check if this is a signup link
			for _, term := range terms {
				lowerLinkText := strings.ToLower(linkText)
				lowerTerm := strings.ToLower(term)
				// Check if URL contains the term (case insensitive)
				lowerHref := strings.ToLower(href)
				lowerResolvedURL := strings.ToLower(resolvedURL)

				if strings.Contains(lowerLinkText, lowerTerm) || strings.Contains(lowerHref, lowerTerm) || strings.Contains(lowerResolvedURL, lowerTerm) {
					log.Printf("Found potential signup link: %s with text: %s (matched term: %s)", href, linkText, term)
					signupLinks = append(signupLinks, href)
					break
				}

				// Special case for 'open' as it's a common pattern for tournament pages
				if isInternal && (strings.Contains(lowerHref, "open") || strings.Contains(lowerResolvedURL, "open") || strings.Contains(lowerLinkText, "open")) {
					log.Printf("Found potential 'open' tournament link: %s with text: %s", href, linkText)
					signupLinks = append(signupLinks, href)
					break
				}
			}
		}
	}

	log.Printf("FindAllSignupLinksFromATags: found %d potential signup links", len(signupLinks))
	return signupLinks
}

// Find all potential signup links in <button> tags, returns a slice of URLs or placeholders
func findAllSignupLinksFromButtons(HTML string, terms []string, currentPageURL string) []string {
	var links []string

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
			containsSignupTerm := false
			for _, term := range terms {
				if strings.Contains(buttonText, strings.ToLower(term)) {
					containsSignupTerm = true
					break
				}
			}

			if containsSignupTerm {
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
					links = append(links, href)
					continue
				}

				// If no href attribute, look for ID to link it with potential JavaScript
				idRegex := regexp.MustCompile(`<button\s+[^>]*id=["']([^"']+)["'][^>]*>`)
				idMatch := idRegex.FindStringSubmatch(match[0])

				if len(idMatch) >= 2 {
					buttonID := "#" + idMatch[1] // Return as a CSS selector
					log.Printf("Found potential signup button with ID: %s and text: %s", buttonID, buttonText)
					links = append(links, buttonID)
					continue
				}

				// Return a placeholder for a button with no ID or href
				placeholder := "#button_with_text:" + buttonText
				log.Printf("Found potential signup button with text: %s but no direct link", buttonText)
				links = append(links, placeholder)
			}
		}
	}
	return links
}

// Find all potential signup links in elements with onclick attributes, returns a slice of URLs or placeholders
func findAllSignupLinksFromOnclickElements(HTML string, terms []string, currentPageURL string) []string {
	var links []string

	// Find all elements with onclick attributes - fixed regex without backreference
	// Changed from: onclickRegex := regexp.MustCompile(`<([a-z0-9]+)\s+[^>]*onclick=["']([^"']+)["'][^>]*>(.*?)</\1>`)
	// To a pattern that doesn't use backreferences (which aren't supported in Go's regexp)
	onclickRegex := regexp.MustCompile(`<[a-z0-9]+\s+[^>]*onclick=["']([^"']+)["'][^>]*>(.*?)</[a-z0-9]+>`)
	matches := onclickRegex.FindAllStringSubmatch(HTML, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			// Element tag is no longer captured in this approach
			onclickValue := match[1] // Onclick JavaScript
			elementText := match[2]  // Element text content

			// Remove HTML tags from element text
			cleanedText := trimAndCleanText(elementText)

			// Check if element text contains signup terms
			containsSignupTerm := false
			for _, term := range terms {
				if strings.Contains(strings.ToLower(cleanedText), strings.ToLower(term)) {
					containsSignupTerm = true
					break
				}
			}

			// Direct check for "Inscriptions" (case insensitive)
			if strings.Contains(strings.ToLower(cleanedText), "inscriptions") {
				containsSignupTerm = true
				log.Printf("Found onclick element with 'Inscriptions': %s", cleanedText)
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
					links = append(links, href)
					continue
				}

				// Return the onclick value as a placeholder
				placeholder := "#onclick_element:" + onclickValue
				log.Printf("Found potential signup onclick element: %s", placeholder)
				links = append(links, placeholder)
			}
		}
	}
	return links
}

// Helper function to trim and clean text from HTML elements
func trimAndCleanText(text string) string {
	// Remove HTML tags
	cleanedText := regexp.MustCompile(`<[^>]+>`).ReplaceAllString(text, " ")

	// Normalize whitespace
	cleanedText = regexp.MustCompile(`\s+`).ReplaceAllString(cleanedText, " ")

	// Trim spaces
	return strings.TrimSpace(cleanedText)
}

// Helper function for min value (Go 1.21+ has this built-in, but adding for compatibility)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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

	// Check for exact matches first (with normalization)
	if normalizeURL(url) == normalizeURL(currentPageURL) {
		log.Printf("LOOP PREVENTION: Found URL is the same as current page URL")
		return true
	}

	// Check for relative links that point to the same page
	currentPath := extractPathFromUrl(currentPageURL)
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		// Resolve the relative URL
		resolvedURL := resolveURL(url, currentPageURL)
		if normalizeURL(resolvedURL) == normalizeURL(currentPageURL) {
			log.Printf("LOOP PREVENTION: Resolved URL %s matches current page URL", resolvedURL)
			return true
		}

		// Check for navigation to parent directories that would cause loops
		resolvedPath := extractPathFromUrl(resolvedURL)
		if currentPath != "" && resolvedPath != "" && strings.HasPrefix(currentPath, resolvedPath) {
			log.Printf("LOOP PREVENTION: URL %s points to a parent directory of current page", url)
			return true
		}
	}

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

	// Check for fragment-only URLs (starting with #)
	if strings.HasPrefix(url, "#") {
		log.Printf("LOOP PREVENTION: Skipping fragment-only URL: %s", url)
		return true
	}

	// Check for self-references (./ or /)
	if url == "./" || url == "/" || url == "." {
		log.Printf("LOOP PREVENTION: Skipping self-reference URL: %s", url)
		return true
	}

	// Additional check for absolute URLs with different protocols
	// Remove protocol for comparison
	noProtocolURL := strings.TrimPrefix(strings.TrimPrefix(normalizeURL(url), "https://"), "http://")
	noProtocolCurrentURL := strings.TrimPrefix(strings.TrimPrefix(normalizeURL(currentPageURL), "https://"), "http://")

	if noProtocolURL == noProtocolCurrentURL {
		log.Printf("LOOP PREVENTION: URLs match when ignoring protocol: %s ~ %s", url, currentPageURL)
		return true
	}

	// If we get here, the URLs are likely different
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

func IsSignupFormPage(HTML string) bool {
	// Check if the HTML contains a form
	containsForm := strings.Contains(strings.ToLower(HTML), "<form")
	log.Printf("IsSignupFormPage check - contains form: %v", containsForm)

	// Check if nom/prénom are in form labels or inputs
	nameFieldsInFormElements := false

	// Check for label elements with text content containing nom/prénom
	labelRegex := regexp.MustCompile(`<label[^>]*>([^<]*(?:nom|prénom|prenom)[^<]*)</label>`)
	if labelRegex.MatchString(HTML) {
		nameFieldsInFormElements = true
		log.Printf("IsSignupFormPage check - found nom/prénom in label elements")
	}

	// Check for input elements with name/id/placeholder attributes containing nom/prénom
	inputRegex := regexp.MustCompile(`<input[^>]*(?:name|id|placeholder)=["']([^"']*(?:nom|prénom|prenom)[^"']*)["'][^>]*>`)
	if inputRegex.MatchString(HTML) {
		nameFieldsInFormElements = true
		log.Printf("IsSignupFormPage check - found nom/prénom in input attributes")
	}

	// Check for tableau links
	containsTableauLinks := false

	// Check for tableau + letter
	tableauLetterRegex := regexp.MustCompile(`<a[^>]*>([^<]*tableau\s+[A-Za-z][^<]*)</a>`)
	tableauLetterMatches := tableauLetterRegex.FindAllString(HTML, -1)
	if len(tableauLetterMatches) > 0 {
		containsTableauLinks = true
		log.Printf("IsSignupFormPage check - found %d tableau + letter links", len(tableauLetterMatches))
	}

	// Check for tableau + number
	tableauNumberRegex := regexp.MustCompile(`<a[^>]*>([^<]*tableau\s+\d+[^<]*)</a>`)
	tableauNumberMatches := tableauNumberRegex.FindAllString(HTML, -1)
	if len(tableauNumberMatches) > 0 {
		containsTableauLinks = true
		log.Printf("IsSignupFormPage check - found %d tableau + number links", len(tableauNumberMatches))
	}

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
	tableauClassementMatches := tableauClassementRegex.FindAllString(HTML, -1)
	if len(tableauClassementMatches) > 0 {
		containsTableauLinks = true
		log.Printf("IsSignupFormPage check - found %d tableau + classement links", len(tableauClassementMatches))
	}

	// Also check for child elements within links
	childElementsRegexPattern := `<a[^>]*>.*?(?:` +
		`tableau\s+[A-Za-z]|` +
		`tableau\s+\d+|` +
		classementRegexPattern +
		`).*?</a>`
	childElementsRegex := regexp.MustCompile(childElementsRegexPattern)
	childElementsMatches := childElementsRegex.FindAllString(HTML, -1)
	if len(childElementsMatches) > 0 {
		containsTableauLinks = true
		log.Printf("IsSignupFormPage check - found %d tableau links with child elements", len(childElementsMatches))
	}

	// Check for inscription page with many tableau links
	// If we find multiple tableau links (>=5), consider it a valid signup page even without a form
	hasMultipleTableauLinks := false
	totalTableauLinks := len(tableauLetterMatches) + len(tableauNumberMatches) +
		len(tableauClassementMatches) + len(childElementsMatches)
	if totalTableauLinks >= 5 {
		hasMultipleTableauLinks = true
		log.Printf("IsSignupFormPage check - found %d total tableau links, considering as signup portal", totalTableauLinks)
	}

	// Check for links containing "/inscriptions/p/tableau"
	inscriptionTableauLinkRegex := regexp.MustCompile(`<a[^>]*href=["'][^"']*?/inscriptions/p/tableau[^"']*["'][^>]*>`)
	inscriptionTableauMatches := inscriptionTableauLinkRegex.FindAllString(HTML, -1)
	if len(inscriptionTableauMatches) >= 3 {
		hasMultipleTableauLinks = true
		log.Printf("IsSignupFormPage check - found %d /inscriptions/p/tableau links, considering as signup portal",
			len(inscriptionTableauMatches))
	}

	// Modified decision logic:
	// A page is valid if:
	// 1. It has a form AND (name fields OR tableau links) - original logic
	// 2. OR it has multiple tableau links (>=5) - new logic
	// 3. OR it has multiple links to /inscriptions/p/tableau paths (>=3) - new logic
	isValid := (containsForm && (nameFieldsInFormElements || containsTableauLinks)) ||
		hasMultipleTableauLinks

	log.Printf("IsSignupFormPage check - final decision: %v (form: %v, name fields: %v, tableau links: %v, multiple tableau links: %v)",
		isValid, containsForm, nameFieldsInFormElements, containsTableauLinks, hasMultipleTableauLinks)

	return isValid
}

// Helper function to check if a URL is on the same domain as the current page URL
func isInternalURL(url string, currentPageURL string) bool {
	// Absolute URLs beginning with / are internal
	if strings.HasPrefix(url, "/") {
		return true
	}

	// URLs without http(s): that don't start with / are relative and internal
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return true
	}

	// For absolute URLs, compare domains
	urlDomain := extractDomain(url)
	currentDomain := extractDomain(currentPageURL)

	return urlDomain != "" && currentDomain != "" && urlDomain == currentDomain
}

// Helper function to resolve a relative URL against the current page URL
func resolveURL(url string, currentPageURL string) string {
	// If it's an absolute URL, return as is
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}

	// Handle absolute paths
	if strings.HasPrefix(url, "/") {
		baseURL := getBaseURL(currentPageURL)
		return baseURL + url
	}

	// Handle relative paths
	resolvedURL := currentPageURL
	if !strings.HasSuffix(resolvedURL, "/") {
		// Remove last path component if it doesn't end with /
		lastSlashIdx := strings.LastIndex(resolvedURL, "/")
		if lastSlashIdx > 0 {
			resolvedURL = resolvedURL[:lastSlashIdx+1]
		} else {
			resolvedURL += "/"
		}
	}
	return resolvedURL + url
}
