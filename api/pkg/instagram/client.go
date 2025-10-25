package instagram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

const (
	// Instagram Graph API base URL
	GraphAPIBaseURL = "https://graph.instagram.com/v18.0"
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

	// Step 3: Wait for Instagram to download the image, then cleanup
	log.Printf("⏳ Waiting 30 seconds for Instagram to download the image...")
	time.Sleep(30 * time.Second)

	// Delete the local image to save disk space
	if err := os.Remove(imagePath); err != nil {
		log.Printf("⚠️  Warning: failed to cleanup image %s: %v", imagePath, err)
	} else {
		log.Printf("🗑️  Cleaned up local image: %s", imagePath)
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

	// Determine image URL based on environment
	var imageURL string
	ginMode := os.Getenv("GIN_MODE")

	if ginMode == "release" {
		// Production: use our server
		filename := filepath.Base(imagePath)
		imageURL = fmt.Sprintf("https://tournois-tt.fr/instagram-images/%s", filename)
		log.Printf("📸 [PROD] Using server URL: %s", imageURL)
	} else {
		// Development: upload to free image hosting (Catbox.moe)
		log.Println("📸 [DEV] Uploading image to Catbox.moe for testing...")
		uploadedURL, err := uploadToImgBB(imagePath)
		if err != nil {
			return "", fmt.Errorf("failed to upload image to Catbox.moe: %w", err)
		}
		imageURL = uploadedURL
		log.Printf("✅ [DEV] Image uploaded to Catbox.moe: %s", imageURL)
	}

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

// uploadToImgBB uploads an image to Catbox.moe (free image hosting, no API key needed) and returns the URL
// This is used for local development testing when GIN_MODE != "release"
func uploadToImgBB(imagePath string) (string, error) {
	// Open the image file
	file, err := os.Open(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	// Prepare multipart form data
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add file field
	part, err := writer.CreateFormFile("fileToUpload", filepath.Base(imagePath))
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	// Copy file content
	if _, err := io.Copy(part, file); err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	// Add reqtype field (required by Catbox)
	if err := writer.WriteField("reqtype", "fileupload"); err != nil {
		return "", fmt.Errorf("failed to write reqtype field: %w", err)
	}

	// Close the writer
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	// Catbox.moe API - completely free, no API key required
	uploadURL := "https://catbox.moe/user/api.php"

	// Create request
	req, err := http.NewRequest("POST", uploadURL, &requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload: %w", err)
	}
	defer resp.Body.Close()

	// Read response (Catbox returns just the URL as plain text)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Trim whitespace and validate URL
	imageURL := string(bytes.TrimSpace(body))
	if imageURL == "" {
		return "", fmt.Errorf("upload succeeded but no URL returned")
	}

	return imageURL, nil
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
