// Package navigation provides services for navigating tournament websites and finding registration forms
package navigation

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

	// Navigate to the website
	response, err := page.Goto(websiteURL, pw.PageGotoOptions{
		WaitUntil: pw.WaitUntilStateNetworkidle,
		Timeout:   pw.Float(DefaultPageTimeout),
	})

	if err != nil {
		return "", fmt.Errorf("failed to navigate to %s: %w", websiteURL, err)
	}

	if response.Status() >= 400 {
		return "", fmt.Errorf("received error status %d", response.Status())
	}

	// JavaScript to check for signup links in header/navigation elements
	jsScript := `
	() => {
		// Keywords related to registration/signup
		const registrationKeywords = [
			'inscription', 'inscrire', 's\'inscrire', 'inscrivez-vous',
			'inscription en ligne', 'signup', 'sign up', 'sign-up', 'register',
			'registration', 'participer', 'participation', 'engage', 'engagement', 
			'form', 'formulaire', 'login', 'connexion', 'compte', 'account', 'créer un compte',
			'continue', 'continuer', 'next', 'suivant', 'étape suivante'
		];

		// Priority levels
		const PRIORITY = {
			HIGH: 'HIGH',
			MEDIUM: 'MEDIUM',
			LOW: 'LOW'
		};

		// Find all links in the document
		const allLinks = Array.from(document.querySelectorAll('a'));
		const headerLinks = [];
		
		// First, try to find links in typical header/navigation elements
		const navElements = document.querySelectorAll('header, nav, .header, .navigation, .nav, .navbar, .menu, .main-menu');
		
		if (navElements.length > 0) {
			// Get links from navigation elements
			for (const nav of navElements) {
				const navLinks = Array.from(nav.querySelectorAll('a'));
				headerLinks.push(...navLinks);
			}
		} else {
			// Fallback: If no clear nav elements, consider links in top 300px of page as potential header links
			for (const link of allLinks) {
				const rect = link.getBoundingClientRect();
				if (rect.top < 300) {
					headerLinks.push(link);
				}
			}
		}
		
		// Calculate relevance score for each link
		const scoredLinks = headerLinks.map(link => {
			const text = (link.textContent || '').toLowerCase().trim();
			const href = link.href.toLowerCase();
			const title = (link.getAttribute('title') || '').toLowerCase();
			
			// Skip empty links and special links
			if (!href || href === '#' || href.startsWith('javascript:') || href.includes('mailto:')) {
				return { link, score: 0, priority: PRIORITY.LOW };
			}
			
			let score = 0;
			let priority = PRIORITY.LOW;
			
			// Check for exact keyword matches
			for (const keyword of registrationKeywords) {
				// Strong indicator in link text
				if (text === keyword || text.includes(keyword)) {
					score += 10;
					priority = PRIORITY.HIGH;
				}
				
				// Keyword in URL path
				if (href.includes('/' + keyword) || href.includes(keyword + '.')) {
					score += 8;
					if (priority !== PRIORITY.HIGH) priority = PRIORITY.MEDIUM;
				}
				
				// Keyword in title attribute
				if (title.includes(keyword)) {
					score += 5;
					if (priority !== PRIORITY.HIGH) priority = PRIORITY.MEDIUM;
				}
			}
			
			// Boost score for buttons or button-like elements
			if (link.classList.contains('btn') || 
				link.classList.contains('button') || 
				link.style.display === 'block' ||
				link.style.padding) {
				score += 3;
			}
			
			return { link: link.href, text, score, priority };
		}).filter(item => item.score > 0);
		
		// Sort by score (highest first)
		scoredLinks.sort((a, b) => b.score - a.score);
		
		return scoredLinks.slice(0, 5); // Return top 5 candidates
	}
	`

	// Execute the JavaScript
	result, err := page.Evaluate(jsScript)
	if err != nil {
		return "", fmt.Errorf("failed to evaluate JavaScript: %w", err)
	}

	// Process the results
	if resultArray, ok := result.([]interface{}); ok && len(resultArray) > 0 {
		bestLinks := make([]map[string]interface{}, 0, len(resultArray))

		// Convert the results to a usable format
		for _, item := range resultArray {
			if linkData, ok := item.(map[string]interface{}); ok {
				bestLinks = append(bestLinks, linkData)
			}
		}

		// If we found potential signup links, check them
		for _, linkData := range bestLinks {
			if linkURL, ok := linkData["link"].(string); ok {
				utils.DebugLog("Checking potential signup link: %s", linkURL)

				// Check if the link URL is valid
				if linkURL == "" || utils.IsURLToSkip(linkURL) {
					continue
				}

				// Try to navigate to the link URL
				linkPage, err := browserContext.NewPage()
				if err != nil {
					utils.DebugLog("Failed to create page for link %s: %v", linkURL, err)
					continue
				}

				// Navigate to the potential signup link
				_, err = linkPage.Goto(linkURL, pw.PageGotoOptions{
					WaitUntil: pw.WaitUntilStateNetworkidle,
					Timeout:   pw.Float(15000), // 15 seconds is enough for checking a link
				})

				if err != nil {
					utils.DebugLog("Failed to navigate to link %s: %v", linkURL, err)
					linkPage.Close()
					continue
				}

				// Simple checks for form elements
				pageContent, err := linkPage.Content()
				if err != nil {
					linkPage.Close()
					continue
				}

				// Check for form elements and common signup indicators
				hasForm := strings.Contains(pageContent, "<form") ||
					strings.Contains(pageContent, "<input type=\"text\"") ||
					strings.Contains(pageContent, "<input type=\"email\"")

				hasRegistrationText := false
				for _, keyword := range registrationKeywords {
					if strings.Contains(strings.ToLower(pageContent), keyword) {
						hasRegistrationText = true
						break
					}
				}

				linkPage.Close()

				// If this looks like a signup form, return it
				if hasForm && hasRegistrationText {
					return linkURL, nil
				}
			}
		}
	}

	return "", fmt.Errorf("no signup links found in website header")
}
