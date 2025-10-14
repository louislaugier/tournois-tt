package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type newsletterRequest struct {
	Email string `json:"email"`
	Scope string `json:"scope"` // optional: all | region | departement
	Area  string `json:"area"`  // optional: region name or departement code
}

// NewsletterHandler subscribes an email to Brevo list ID 11
func NewsletterHandler(c *gin.Context) {
	var req newsletterRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email requis"})
		return
	}

	apiKey := os.Getenv("BREVO_API_KEY")
	if apiKey == "" {
		log.Printf("BREVO_API_KEY manquant")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "configuration manquante"})
		return
	}

	// Brevo add contact to list 11
	attrs := map[string]any{}
	if req.Scope != "" {
		attrs["SCOPE"] = req.Scope
	}
	if req.Area != "" {
		attrs["AREA"] = req.Area
	}

	payload := map[string]any{
		"email":         req.Email,
		"listIds":       []int{11},
		"updateEnabled": true,
		"attributes":    attrs,
	}
	body, _ := json.Marshal(payload)

	reqHTTP, _ := http.NewRequest("POST", "https://api.brevo.com/v3/contacts", bytes.NewReader(body))
	reqHTTP.Header.Set("api-key", apiKey)
	reqHTTP.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(reqHTTP)
	if err != nil {
		log.Printf("brevo error: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "erreur d'inscription"})
		return
	}
	defer resp.Body.Close()

	// Read response body (can be empty on 204)
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 300 {
		// Try to decode Brevo error
		var be map[string]any
		_ = json.Unmarshal(respBody, &be)
		log.Printf("brevo error: status=%d body=%s", resp.StatusCode, string(respBody))
		c.JSON(http.StatusBadGateway, gin.H{"error": be})
		return
	}

	// 201 Created => contact created; 204 No Content => updated existing
	result := "ok"
	switch resp.StatusCode {
	case http.StatusCreated:
		result = "created"
	case http.StatusNoContent:
		result = "updated"
	default:
		// Some responses may be 200 with body; keep generic
		result = "ok"
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "result": result})
}
