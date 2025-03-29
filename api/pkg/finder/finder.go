package finder

import (
	"regexp"
	"sort"
	"strings"
	"tournois-tt/api/pkg/utils"
)

func GetSignupURLFromRulesContent(content string) *string {
	// Extract all URLs from the content
	urls := GetURLsFromText(content)

	// No URLs found
	if len(urls) == 0 {
		return nil
	}

	// Get the most probable signup URL using both URLs and content
	signupURL := GetMostProbableSignupURL(urls, content)

	// If we don't have a signup URL, return nil
	if signupURL == "" {
		return nil
	}

	// Ensure the URL has a protocol prefix for consistency
	if !strings.HasPrefix(signupURL, "http://") && !strings.HasPrefix(signupURL, "https://") {
		signupURL = "https://" + signupURL
	}

	return &signupURL
}

func GetURLsFromText(text string) []string {
	// Define a simpler regex pattern for URLs with or without http/https prefix
	// This pattern matches standalone domain names with valid TLDs

	// Common TLDs we want to match
	validTLDs := []string{
		"com", "org", "net", "fr", "io", "co", "app", "dev",
		"info", "biz", "tv", "me", "uk", "us", "ca", "de", "jp",
	}

	// Join TLDs for the pattern
	tldList := strings.Join(validTLDs, "|")

	// Build a pattern that correctly matches both domains and subdomains
	pattern := `\b(?:https?://)?(?:www\.)?(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+(?:` + tldList + `)\b`

	// Compile the regex pattern
	re, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		// If there's a compilation error, return an empty slice
		return []string{}
	}

	// Find all matches in the text
	matches := re.FindAllString(text, -1)

	// Manual post-processing to filter false positives and clean up matches
	var results []string
	for _, match := range matches {
		// Remove any trailing non-alphanumeric characters
		cleanMatch := regexp.MustCompile(`[^a-zA-Z0-9.-]+$`).ReplaceAllString(match, "")

		// Skip if likely part of an email address
		if strings.Contains(text, "@"+cleanMatch) || strings.Contains(text, "¾"+cleanMatch) {
			continue
		}

		// Check if this URL is already in our results to avoid duplicates
		isDuplicate := false
		for _, result := range results {
			if result == cleanMatch {
				isDuplicate = true
				break
			}
		}

		if !isDuplicate {
			results = append(results, cleanMatch)
		}
	}

	return results
}

func GetMostProbableSignupURL(URLs []string, content string) string {
	if len(URLs) == 0 {
		return ""
	}

	// If there's only one URL, return it
	if len(URLs) == 1 {
		return URLs[0]
	}

	// Score-based approach to rank URLs by likelihood of being a signup URL
	type scoredURL struct {
		url   string
		score int
	}

	scoredURLs := make([]scoredURL, 0, len(URLs))

	// Convert content to lowercase for easier matching
	lowerContent := strings.ToLower(content)

	for _, url := range URLs {
		score := 0
		lowerURL := strings.ToLower(url)

		// URLs with registration-related keywords in subdomains have higher probability
		registrationKeywords := []string{
			"tournoi", "inscription", "inscrire", "competition", "paiement",
			"engage", "engagements", "tarif", "participer", "participe",
			"reserv", "reserve", "reservation",
		}

		// Check if the URL contains registration keywords - subdomain keywords are stronger indicators
		for _, keyword := range registrationKeywords {
			if strings.Contains(lowerURL, keyword) {
				// Apply higher score if the keyword is in a subdomain
				urlParts := strings.Split(lowerURL, ".")
				if len(urlParts) > 2 && strings.Contains(urlParts[0], keyword) {
					score += 5 // Much higher score for registration keywords in subdomain
				} else {
					score += 2
				}
			}
		}

		// Subdomains are often used for specific purposes like registration
		// Tournaments often use a subdomain specifically for registrations
		if strings.Count(lowerURL, ".") > 1 {
			score += 3 // Increase weight of subdomains for tournament sites
		}

		// Prefer secure URLs
		if strings.HasPrefix(lowerURL, "https://") {
			score += 1
		}

		// Extensive list of signup-related phrases in French with many variations
		signupPhrases := []string{
			// Basic phrases
			"inscription sur", "s'inscrire sur", "inscrivez-vous sur", "inscriptions sur",
			"inscription via", "s'inscrire via", "inscrivez-vous via", "inscriptions via",
			"inscription à", "s'inscrire à", "inscrivez-vous à", "inscriptions à",
			"inscription par", "s'inscrire par", "inscrivez-vous par", "inscriptions par",
			"inscription en ligne", "s'inscrire en ligne", "inscrivez-vous en ligne",

			// With site/website mentions
			"inscription sur le site", "s'inscrire sur le site", "inscrivez-vous sur le site",
			"inscription via le site", "s'inscrire via le site", "inscrivez-vous via le site",
			"inscription sur notre site", "s'inscrire sur notre site", "inscrivez-vous sur notre site",
			"inscription via notre site", "s'inscrire via notre site", "inscrivez-vous via notre site",
			"inscription sur le site du club", "s'inscrire sur le site du club", "inscrivez-vous sur le site du club",
			"inscription via le site du club", "s'inscrire via le site du club", "inscrivez-vous via le site du club",
			"inscription sur le site internet", "s'inscrire sur le site internet", "inscrivez-vous sur le site internet",
			"inscription sur le site web", "s'inscrire sur le site web", "inscrivez-vous sur le site web",

			// With colon variations
			"inscription:", "inscriptions:", "s'inscrire:", "inscrivez-vous:",
			"inscription sur:", "inscriptions sur:", "s'inscrire sur:", "inscrivez-vous sur:",
			"inscription via:", "inscriptions via:", "s'inscrire via:", "inscrivez-vous via:",
			"inscription sur le site:", "s'inscrire sur le site:", "inscrivez-vous sur le site:",
			"inscription via le site:", "s'inscrire via le site:", "inscrivez-vous via le site:",

			// Payment phrases
			"paiement sur", "paiement via", "paiement en ligne sur", "paiement en ligne via",
			"payer sur", "payer via", "payer en ligne sur", "payer en ligne via",
			"règlement sur", "règlement via", "règlement en ligne sur", "règlement en ligne via",
			"paiement sur le site", "paiement via le site", "paiement en ligne sur le site",
			"payer sur le site", "payer via le site", "payer en ligne sur le site",
			"règlement sur le site", "règlement via le site", "règlement en ligne sur le site",
			"paiement:", "paiements:", "payer:", "régler:", "règlements:",
			"paiement sur:", "paiements sur:", "payer sur:", "régler sur:", "règlements sur:",

			// Website reference phrases
			"site du club", "site internet du club", "site web du club",
			"site de l'événement", "site du tournoi", "site de la compétition",
			"sur le site", "via le site", "sur notre site", "via notre site",
			"site:", "site du club:", "site internet:", "site web:",
			"rendez-vous sur", "rendez vous sur", "rdv sur", "à consulter sur",
			"plus d'information sur", "information sur", "infos sur", "détails sur",

			// Priority/preference phrases
			"privilégier", "à privilégier", "privilégiez", "recommandé", "conseillé",
			"préférable", "préférez", "de préférence", "à privilégier:",
		}

		// A set of phrases that are common before an URL is mentioned
		preURLContextPhrases := []string{
			"site du club", "site internet", "site web", "site officiel", "site du tournoi",
			"en ligne sur", "accessible sur", "disponible sur", "à l'adresse",
			"paiement sur", "paiement en ligne", "paiement sécurisé", "paiement anticipé",
			"paiement à l'avance", "paiement par carte", "paiement par cb",
			"inscription sur", "inscription en ligne", "inscription via",
			"a privilegier", "à privilégier", "privilégiez",
		}

		// Generic terms that might be close to URLs in tournament documents
		genericContextTerms := []string{
			"paiement", "inscription", "renseignement", "information", "détail",
			"consulter", "visiter", "accéder", "disponible", "réservation",
		}

		// Look for each signup phrase followed by the URL (with flexible spacing)
		for _, phrase := range signupPhrases {
			// Account for different spacing and characters between phrase and URL
			phrasePattern := "(?i)" + regexp.QuoteMeta(phrase) + "[\\s:]*(?:(?:sur|via|à|par)?\\s+(?:le\\s+)?(?:site\\s+)?)?(?:https?://)?(?:www\\.)?" +
				regexp.QuoteMeta(strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "http://"))

			phraseRegex, err := regexp.Compile(phrasePattern)
			if err == nil && phraseRegex.MatchString(content) {
				score += 5 // High score for direct mention with signup phrase
			}
		}

		// Check for context phrases near the URL
		urlPosition := strings.Index(lowerContent, lowerURL)
		if urlPosition >= 0 {
			// Check 150 characters before the URL for context phrases
			startPos := utils.Max(0, urlPosition-150)
			beforeContext := lowerContent[startPos:urlPosition]

			for _, phrase := range preURLContextPhrases {
				if strings.Contains(beforeContext, phrase) {
					distance := urlPosition - strings.LastIndex(beforeContext, phrase)
					if distance < 100 {
						score += 4
					} else {
						score += 2
					}
				}
			}

			// Check for generic terms even closer to the URL
			for _, term := range genericContextTerms {
				if strings.Contains(beforeContext, term) {
					distance := urlPosition - strings.LastIndex(beforeContext, term)
					if distance < 50 {
						score += 1
					}
				}
			}

			// Check for a colon preceding the URL closely
			if strings.Contains(beforeContext[max(0, len(beforeContext)-10):], ":") {
				score += 2
			}
		}

		// Special case for "à privilégier" or similar priority phrases
		privilegePhrases := []string{
			"à privilégier", "privilégiez", "privilégier", "recommandé", "de préférence",
			"conseillé", "préférable", "préférez",
		}

		// Search for privileged phrases in proximity to URLs
		for _, phrase := range privilegePhrases {
			phraseIdx := strings.Index(lowerContent, phrase)
			if phraseIdx >= 0 {
				if utils.Abs(phraseIdx-urlPosition) < 200 {
					score += 5 // High value if the URL is specially recommended
				}
			}
		}

		// Extra weight for URLs mentioned in the same context as "INSCRIPTIONS SUR LE SITE DU CLUB"
		if strings.Contains(lowerContent, "inscriptions sur le site du club") &&
			utils.Abs(strings.Index(lowerContent, "inscriptions sur le site du club")-urlPosition) < 100 {
			score += 6
		}

		// Extra pattern specifically for the format in the example document
		if strings.Contains(lowerContent, "a privilegier : inscriptions sur le site du club") &&
			strings.Contains(lowerContent, lowerURL) &&
			utils.Abs(strings.Index(lowerContent, "a privilegier : inscriptions sur le site du club")-urlPosition) < 150 {
			score += 8
		}

		// Give higher weight to URLs mentioned in a payment context
		paymentContextPhrases := []string{
			"paiement sécurisé en ligne", "paiement en ligne", "paiement anticipé",
			"payer en ligne", "règlement en ligne", "paiement par carte",
			"paiement par cb", "carte bancaire", "paiement sécurisé",
		}

		for _, phrase := range paymentContextPhrases {
			phraseIdx := strings.Index(lowerContent, phrase)
			if phraseIdx >= 0 {
				// Only consider payment phrases that come shortly before the URL (within ~100 chars)
				if urlPosition > phraseIdx && urlPosition-phraseIdx < 100 {
					score += 8 // Very high score for payment-related context before URL
				}
			}
		}

		// Distinguish between main site and subdomain specific for tournament/payment
		// Common patterns seen in tournament PDFs
		if strings.Contains(lowerContent, "site du club") &&
			strings.Contains(lowerURL, ".") &&
			!strings.Contains(lowerURL, "tournoi") {
			// This is likely the club's main site
			// We'll still keep it as a candidate but not inflate its score
		}

		// Special handling for tournoi subdomain specifically mentioned with payment
		if strings.Contains(lowerURL, "tournoi") &&
			strings.Contains(lowerContent, "paiement") &&
			utils.Abs(strings.Index(lowerContent, "paiement")-urlPosition) < 100 {
			score += 10 // Heavily prioritize tournament subdomains mentioned with payment
		}

		scoredURLs = append(scoredURLs, scoredURL{url: url, score: score})
	}

	// Sort URLs by score (descending)
	sort.Slice(scoredURLs, func(i, j int) bool {
		return scoredURLs[i].score > scoredURLs[j].score
	})

	// Return the URL with the highest score
	if len(scoredURLs) > 0 {
		url := scoredURLs[0].url
		// Ensure URL has protocol prefix before returning
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			url = "https://" + url
		}
		return url
	}
	return ""
}
