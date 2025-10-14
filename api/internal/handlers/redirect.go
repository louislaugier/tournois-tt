package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
	"tournois-tt/api/pkg/cache"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RedirectRulesHandler redirects /:id to the rules PDF URL if available, else to /
func RedirectRulesHandler(c *gin.Context) {
	idStr := c.Param("id")
	// Only allow numeric ids
	if _, err := strconv.Atoi(idStr); err != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}

	// Read directly from in-memory cache to avoid any season filters
	if t, ok := cache.GetCachedTournament(mustAtoi(idStr)); ok {
		if t.Rules != nil && t.Rules.URL != "" {
			// Track the redirect in GA4
			go trackRedirect(c, t.ID, t.Name, t.Rules.URL)

			c.Redirect(http.StatusFound, t.Rules.URL)
			return
		}
	}

	// Fallback to home if no match
	c.Redirect(http.StatusFound, "/")
}

func mustAtoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

// trackRedirect sends a server-side event to GA4 Measurement Protocol
func trackRedirect(c *gin.Context, tournamentID int, tournamentName, rulesURL string) {
	measurementID := os.Getenv("GA_MEASUREMENT_ID")
	apiSecret := os.Getenv("GA_API_SECRET")

	if measurementID == "" || apiSecret == "" {
		return // Skip tracking if not configured
	}

	// Generate or retrieve client_id (using IP as fallback)
	clientID := uuid.New().String()

	// Build Measurement Protocol payload
	payload := map[string]any{
		"client_id": clientID,
		"events": []map[string]any{
			{
				"name": "click",
				"params": map[string]any{
					"event_category":       "short_link",
					"event_label":          fmt.Sprintf("Tournament %d", tournamentID),
					"link_url":             rulesURL,
					"link_type":            "rules",
					"link_source":          "short_url",
					"content_group":        "redirect",
					"tournament_id":        tournamentID,
					"tournament_name":      tournamentName,
					"session_id":           time.Now().Unix(),
					"engagement_time_msec": 100,
				},
			},
		},
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://www.google-analytics.com/mp/collect?measurement_id=%s&api_secret=%s", measurementID, apiSecret)

	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("GA4 tracking error: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		log.Printf("GA4 tracking failed: status=%d", resp.StatusCode)
	}
}
