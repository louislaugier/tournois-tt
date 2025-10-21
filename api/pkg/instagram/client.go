package instagram

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tournois-tt/api/internal/config"
)

const (
	// Instagram Graph API base URL
	GraphAPIBaseURL = "https://graph.instagram.com/v18.0"

	// Retry configuration
	maxRetries       = 3
	initialBackoff   = 1 * time.Second
	maxBackoff       = 30 * time.Second
	rateLimitBackoff = 60 * time.Second
)

// Client represents an Instagram API client
type Client struct {
	config     Config
	httpClient *http.Client
}

// NewClient creates a new Instagram API client
func NewClient(config Config) *Client {
	// Try to load the latest token from storage
	if token, err := LoadToken(); err == nil && token != "" {
		config.AccessToken = token
	}

	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// PostTournament posts a tournament image to Instagram feed
func (c *Client) PostTournament(tournament TournamentImage) (*TournamentNotification, error) {
	if !c.config.Enabled {
		return nil, fmt.Errorf("Instagram posting is disabled in configuration")
	}

	// Validate credentials are present
	if c.config.AccessToken == "" {
		return nil, fmt.Errorf("Instagram access token is not configured")
	}
	if c.config.PageID == "" {
		return nil, fmt.Errorf("Instagram page ID is not configured")
	}

	notification := &TournamentNotification{
		Tournament: tournament,
		SentAt:     time.Now(),
		Success:    false,
	}

	// Generate the tournament image
	imagePath, err := GenerateTournamentImage(tournament)
	if err != nil {
		notification.Error = fmt.Sprintf("Failed to generate image: %v", err)
		return notification, err
	}
	defer CleanupImage(imagePath)

	// Post the image
	postID, err := c.postImage(imagePath, tournament)
	if err != nil {
		notification.Error = fmt.Sprintf("Failed to post image: %v", err)
		return notification, err
	}

	notification.MessageID = postID // Reusing MessageID field for post ID
	notification.Success = true

	log.Printf("Successfully posted tournament %d (%s) to Instagram - Post ID: %s",
		tournament.TournamentID, tournament.Name, postID)

	return notification, nil
}

// postImage posts a tournament image to Instagram using the Content Publishing API
func (c *Client) postImage(imagePath string, tournament TournamentImage) (string, error) {
	// Step 1: Create container (upload image metadata)
	containerID, err := c.createMediaContainer(imagePath, tournament)
	if err != nil {
		return "", fmt.Errorf("failed to create media container: %w", err)
	}

	log.Printf("✅ Media container created: %s", containerID)

	// Step 2: Publish the container
	postID, err := c.publishMediaContainer(containerID)
	if err != nil {
		return "", fmt.Errorf("failed to publish media: %w", err)
	}

	log.Printf("✅ Post published: %s", postID)

	// Also save locally for record keeping
	if err := c.saveImageLocally(imagePath, tournament); err != nil {
		log.Printf("Warning: Failed to save image locally: %v", err)
	}

	return postID, nil
}

// createMediaContainer creates a media container with the image
// Note: Instagram requires the image to be accessible via a public HTTPS URL
func (c *Client) createMediaContainer(imagePath string, tournament TournamentImage) (string, error) {
	// Prepare caption (convert endowment from cents to euros)
	caption := fmt.Sprintf(`🎾 %s

🏆 Type: %s
🏓 Club: %s
💰 Dotation: %d €
📅 Du %s au %s
📍 %s

🔗 Plus d'infos: %s

#TennisDeTable #PingPong #FFTT #Tournoi`,
		tournament.Name,
		tournament.Type,
		tournament.Club,
		tournament.Endowment/100, // Convert cents to euros
		formatDate(tournament.StartDate),
		formatDate(tournament.EndDate),
		tournament.Address,
		tournament.TournamentURL,
	)

	// TEMPORARY: Use mock image for E2E testing
	// TODO: Implement proper image hosting and uncomment the code below
	imageURL := "https://us-metro.org/wp-content/uploads/2022/06/banniere-tennis-de-table-1400x788-1.jpg"

	// // Save image to public directory first
	// timestamp := time.Now().Unix()
	// publicFilename := fmt.Sprintf("tournament_%d_%d.png", tournament.TournamentID, timestamp)
	// publicPath := filepath.Join("./instagram-images", publicFilename)
	//
	// if err := os.MkdirAll("./instagram-images", 0755); err != nil {
	// 	return "", fmt.Errorf("failed to create images directory: %w", err)
	// }
	//
	// imageData, err := os.ReadFile(imagePath)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to read image: %w", err)
	// }
	//
	// if err := os.WriteFile(publicPath, imageData, 0644); err != nil {
	// 	return "", fmt.Errorf("failed to save image: %w", err)
	// }
	//
	// // Construct public URL - Instagram requires HTTPS
	// imageURL := fmt.Sprintf("https://tournois-tt.fr/instagram-images/%s", publicFilename)

	log.Printf("📸 Using image URL: %s", imageURL)

	// Create container via Instagram API (URL-encode parameters)
	createURL := fmt.Sprintf("%s/%s/media?image_url=%s&caption=%s&access_token=%s",
		GraphAPIBaseURL,
		c.config.PageID,
		url.QueryEscape(imageURL),
		url.QueryEscape(caption),
		url.QueryEscape(c.config.AccessToken),
	)

	resp, err := c.httpClient.Post(createURL, "application/json", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if json.Unmarshal(body, &errResp) == nil {
			return "", fmt.Errorf("Instagram API error: %s (code: %d)", errResp.Error.Message, errResp.Error.Code)
		}
		return "", fmt.Errorf("create container failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse container response: %w", err)
	}

	return result.ID, nil
}

// publishMediaContainer publishes a media container
func (c *Client) publishMediaContainer(containerID string) (string, error) {
	publishURL := fmt.Sprintf("%s/%s/media_publish?creation_id=%s&access_token=%s",
		GraphAPIBaseURL,
		c.config.PageID,
		url.QueryEscape(containerID),
		url.QueryEscape(c.config.AccessToken),
	)

	resp, err := c.httpClient.Post(publishURL, "application/json", nil)
	if err != nil {
		return "", fmt.Errorf("failed to publish: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if json.Unmarshal(body, &errResp) == nil {
			return "", fmt.Errorf("Instagram API error: %s (code: %d)", errResp.Error.Message, errResp.Error.Code)
		}
		return "", fmt.Errorf("publish failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse publish response: %w", err)
	}

	return result.ID, nil
}

// saveImageLocally saves the generated tournament image to a local folder
func (c *Client) saveImageLocally(imagePath string, tournament TournamentImage) error {
	// Create instagram-images directory if it doesn't exist
	imagesDir := "./instagram-images"
	if err := os.MkdirAll(imagesDir, 0755); err != nil {
		return fmt.Errorf("failed to create images directory: %w", err)
	}

	// Generate filename with tournament ID and timestamp
	timestamp := time.Now().Format("20060102-150405")
	destPath := filepath.Join(imagesDir, fmt.Sprintf("tournament_%d_%s.png",
		tournament.TournamentID, timestamp))

	// Read source file
	sourceData, err := os.ReadFile(imagePath)
	if err != nil {
		return fmt.Errorf("failed to read source image: %w", err)
	}

	// Write to destination
	if err := os.WriteFile(destPath, sourceData, 0644); err != nil {
		return fmt.Errorf("failed to write image: %w", err)
	}

	log.Printf("Saved tournament image to: %s", destPath)
	return nil
}

// formatDate formats a date string for display
func formatDate(dateStr string) string {
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return dateStr
	}
	return t.Format("02/01/2006")
}

// TestConnection tests the Instagram API connection
func (c *Client) TestConnection() error {
	if !c.config.Enabled {
		return fmt.Errorf("Instagram integration is disabled")
	}

	// Validate credentials are present
	if c.config.AccessToken == "" {
		return fmt.Errorf("Instagram access token is not configured")
	}
	if c.config.PageID == "" {
		return fmt.Errorf("Instagram page ID is not configured")
	}

	// Test by getting account info
	testURL := fmt.Sprintf("%s/%s?fields=id,username&access_token=%s",
		GraphAPIBaseURL, c.config.PageID, c.config.AccessToken)

	resp, err := c.httpClient.Get(testURL)
	if err != nil {
		return fmt.Errorf("failed to connect to Instagram API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Instagram API test failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// isTokenError checks if the error is related to an invalid/expired token
func isTokenError(statusCode int, errResp *ErrorResponse) bool {
	if statusCode == http.StatusUnauthorized {
		return true
	}

	if errResp != nil {
		// Check for token-related error codes and messages
		tokenErrorCodes := []int{190, 102, 104} // OAuthException codes
		for _, code := range tokenErrorCodes {
			if errResp.Error.Code == code {
				return true
			}
		}

		// Check error message for token-related keywords
		msg := strings.ToLower(errResp.Error.Message)
		if strings.Contains(msg, "token") ||
			strings.Contains(msg, "expired") ||
			strings.Contains(msg, "invalid") ||
			strings.Contains(msg, "authentication") {
			return true
		}
	}

	return false
}

// isRateLimitError checks if the error is a rate limit error
func isRateLimitError(statusCode int, errResp *ErrorResponse) bool {
	if statusCode == http.StatusTooManyRequests {
		return true
	}

	if errResp != nil && errResp.Error.Code == 4 { // Rate limit error code
		return true
	}

	return false
}

// isRetriableError checks if the error can be retried
func isRetriableError(statusCode int) bool {
	// Network errors, server errors, and service unavailable are retriable
	return statusCode >= 500 || statusCode == 408 || statusCode == 429
}

// retryWithBackoff executes a function with exponential backoff retry logic
func (c *Client) retryWithBackoff(operation string, fn func() error) error {
	var lastErr error
	backoff := initialBackoff

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("Retrying %s (attempt %d/%d) after %v...", operation, attempt, maxRetries, backoff)
			time.Sleep(backoff)

			// Exponential backoff with max limit
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}

		err := fn()
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// If it's not a retriable error, don't retry
		if !strings.Contains(err.Error(), "status 5") &&
			!strings.Contains(err.Error(), "status 408") &&
			!strings.Contains(err.Error(), "status 429") {
			return err
		}

		log.Printf("Attempt %d failed for %s: %v", attempt+1, operation, err)
	}

	return fmt.Errorf("failed after %d attempts: %w", maxRetries+1, lastErr)
}

// handleTokenError attempts to refresh the token and update the client config
func (c *Client) handleTokenError() error {
	log.Println("WARNING: Instagram API returned token error - attempting to refresh token...")

	// Try to refresh the token
	if err := ForceRefreshToken(config.InstagramAccessToken); err != nil {
		return fmt.Errorf("failed to refresh token after error: %w", err)
	}

	// Update the client's token from config (it should have been updated by ForceRefreshToken)
	// Note: The config package variables are global, so they should be updated
	c.config.AccessToken = config.InstagramAccessToken

	log.Println("SUCCESS: Token refreshed successfully after error")
	return nil
}
