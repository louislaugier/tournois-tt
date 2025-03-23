// Package finder provides services for navigating tournament websites and finding registration forms
package finder

import (
	"fmt"
	"strings"

	"tournois-tt/api/pkg/utils"

	pw "github.com/playwright-community/playwright-go"
)

// CheckWebsiteHeaderForSignupLink scans a website's header/navigation for signup links
func CheckWebsiteHeaderForSignupLink(websiteURL string, browserContext pw.BrowserContext) (string, error) {
	// Skip if URL is known to be irrelevant
	if utils.IsURLToSkip(websiteURL) {
		return "", fmt.Errorf("URL is in skip list: %s", websiteURL)
	}

	// Create a new page
	page, err := browserContext.NewPage()
	if err != nil {
		return "", fmt.Errorf("failed to create page: %w", err)
	}
	defer page.Close()

	// Navigate to the URL
	if _, err := page.Goto(websiteURL, pw.PageGotoOptions{
		WaitUntil: pw.WaitUntilStateNetworkidle,
		Timeout:   pw.Float(DefaultPageTimeout),
	}); err != nil {
		return "", fmt.Errorf("failed to navigate to URL: %w", err)
	}

	// JavaScript to find links in the header/navigation area
	jsScript := `
	() => {
		const signupKeywords = [
			'inscription', 'inscrivez', 'inscrire', 's\'inscrire', 'enregistrement', 
			'signup', 'sign up', 'sign-up', 'register', 'registration', 'enroll',
			'participer', 'participation', 'engagements', 'engagement',
			'inscrire'
		];

		// Function to check if a link text contains signup-related keywords
		const isSignupLink = (text) => {
			if (!text) return false;
			const lowercaseText = text.toLowerCase();
			return signupKeywords.some(keyword => lowercaseText.includes(keyword));
		};

		// Check elements that are likely to be in the header/navigation
		const potentialHeaderElements = [
			'header', 'nav', '.header', '.navigation', '.menu', '.navbar', '.nav-bar', 
			'#header', '#nav', '#main-menu', '#top-menu', '.top-menu', '#primary-menu',
			'.site-navigation', '.main-navigation'
		];

		let signupLinks = [];

		// Look for links in potential header elements
		for (const selector of potentialHeaderElements) {
			const elements = document.querySelectorAll(selector);
			for (const element of elements) {
				const links = element.querySelectorAll('a');
				for (const link of links) {
					if (isSignupLink(link.textContent) || isSignupLink(link.innerText) || 
						isSignupLink(link.getAttribute('title')) || isSignupLink(link.getAttribute('aria-label'))) {
						signupLinks.push({
							href: link.href,
							text: link.textContent || link.innerText,
							isVisible: link.getBoundingClientRect().height > 0 && link.getBoundingClientRect().width > 0
						});
					}
				}
			}
		}

		// If no links found in headers, look at all visible links on the page
		if (signupLinks.length === 0) {
			const allLinks = document.querySelectorAll('a');
			for (const link of allLinks) {
				if (isSignupLink(link.textContent) || isSignupLink(link.innerText) || 
					isSignupLink(link.getAttribute('title')) || isSignupLink(link.getAttribute('aria-label'))) {
					const rect = link.getBoundingClientRect();
					// Check if the link is in the top 30% of the page (likely header area)
					const isInTopPortion = rect.top < (window.innerHeight * 0.3);
					if (isInTopPortion && rect.height > 0 && rect.width > 0) {
						signupLinks.push({
							href: link.href,
							text: link.textContent || link.innerText,
							isVisible: true
						});
					}
				}
			}
		}

		// Sort by priority: visible links first
		signupLinks.sort((a, b) => {
			if (a.isVisible && !b.isVisible) return -1;
			if (!a.isVisible && b.isVisible) return 1;
			return 0;
		});

		return signupLinks.map(link => link.href);
	}
	`

	result, err := page.Evaluate(jsScript)
	if err != nil {
		return "", fmt.Errorf("failed to evaluate script: %w", err)
	}

	var signupLinks []string
	if linksArr, ok := result.([]interface{}); ok {
		for _, link := range linksArr {
			if linkStr, ok := link.(string); ok && linkStr != "" && strings.HasPrefix(linkStr, "http") {
				// Filter out social media links, email links, etc.
				if !utils.IsDomainToSkip(utils.ExtractDomain(linkStr)) {
					signupLinks = append(signupLinks, linkStr)
				}
			}
		}
	}

	if len(signupLinks) > 0 {
		utils.DebugLog("Found %d potential signup links in header", len(signupLinks))
		return signupLinks[0], nil
	}

	return "", fmt.Errorf("no signup links found in header")
}

// PrioritizeLinksByRelevance sorts links by their relevance to registration
func PrioritizeLinksByRelevance(links []string) []string {
	type scoredLink struct {
		url   string
		score int
	}

	scoredLinks := make([]scoredLink, 0, len(links))

	for _, link := range links {
		score := 0
		lowerLink := strings.ToLower(link)

		// Check for registration-related keywords in the URL
		registrationKeywords := []string{"inscription", "register", "signup", "sign-up", "enroll", "participer", "engagement"}
		for _, keyword := range registrationKeywords {
			if strings.Contains(lowerLink, keyword) {
				score += 10
			}
		}

		// URLs with "formulaire" (form) get a bonus
		if strings.Contains(lowerLink, "form") || strings.Contains(lowerLink, "formulaire") {
			score += 5
		}

		// URLs that likely point to pages (not files)
		if strings.Contains(lowerLink, ".html") || strings.Contains(lowerLink, ".php") || !strings.Contains(lowerLink, ".") {
			score += 3
		}

		// URLs with tournament-related terms
		tournamentKeywords := []string{"tournoi", "tournament", "competition", "championnat", "championship"}
		for _, keyword := range tournamentKeywords {
			if strings.Contains(lowerLink, keyword) {
				score += 3
			}
		}

		scoredLinks = append(scoredLinks, scoredLink{url: link, score: score})
	}

	// Sort links by score (descending)
	for i := 0; i < len(scoredLinks)-1; i++ {
		for j := i + 1; j < len(scoredLinks); j++ {
			if scoredLinks[j].score > scoredLinks[i].score {
				scoredLinks[i], scoredLinks[j] = scoredLinks[j], scoredLinks[i]
			}
		}
	}

	// Extract URLs from scored links
	result := make([]string, len(scoredLinks))
	for i, link := range scoredLinks {
		result[i] = link.url
	}

	return result
}
