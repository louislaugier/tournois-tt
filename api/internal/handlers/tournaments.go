package handlers

import (
	"net/http"
	"tournois-tt/api/pkg/fftt"

	"github.com/gin-gonic/gin"
)

func GetTournaments(c *gin.Context) {
	// Get all query parameters
	queryParams := c.Request.URL.Query()

	// Call FFTT API
	resp, err := fftt.GetClient().GetTournaments(queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data from FFTT API"})
		return
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"error": "FFTT API returned an error"})
		return
	}

	// Forward the response directly to the client
	c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
}
