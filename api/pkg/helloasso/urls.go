package helloasso

import (
	"regexp"
	"strings"
)

// URLRegex returns a regex to find HelloAsso URLs
func URLRegex() *regexp.Regexp {
	// helloAssoURLRegex is a regex to find HelloAsso URLs
	return regexp.MustCompile(`https?://(?:www\.)?` +
		strings.TrimPrefix(strings.ReplaceAll(BaseURL, ".", "\\."), "https://") +
		`/[^\s"']+`)
}

// IsHelloAssoURL checks if a URL is a HelloAsso URL
func IsHelloAssoURL(url string) bool {
	// Use the HelloAsso BaseURL constant to avoid hardcoded URLs
	baseURL := strings.TrimPrefix(BaseURL, "https://")
	return strings.Contains(strings.ToLower(url), strings.ToLower(baseURL))
}
