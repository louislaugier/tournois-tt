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
	// Threads Graph API base URL
	ThreadsAPIBaseURL = "https://graph.threads.net/v1.0"
)

// Client represents an Instagram API client
type Client struct {
	config     Config
	httpClient *http.Client
}

// NewClient creates a new Instagram API client
func NewClient(config Config) *Client {
	// Try to load the latest Instagram token from storage
	if token, err := LoadToken(); err == nil && token != "" {
		config.AccessToken = token
	}

	// Try to load the latest Threads token from storage
	if token, err := LoadThreadsToken(); err == nil && token != "" {
		config.ThreadsAccessToken = token
	}

	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// PostTournament posts a tournament image to Instagram feed and Threads
func (c *Client) PostTournament(tournament TournamentImage) (*TournamentNotification, error) {
	if !c.config.Enabled {
		return nil, fmt.Errorf("instagram posting is disabled in configuration")
	}

	// Validate credentials are present
	if c.config.AccessToken == "" {
		return nil, fmt.Errorf("instagram access token is not configured")
	}
	if c.config.PageID == "" {
		return nil, fmt.Errorf("instagram page ID is not configured")
	}

	notification := &TournamentNotification{
		Tournament: tournament,
		SentAt:     time.Now(),
		Success:    false,
	}

	// Generate the tournament image
	imagePath, err := GenerateTournamentImage(tournament)
	if err != nil {
		notification.Error = fmt.Sprintf("failed to generate image: %v", err)
		return notification, err
	}

	// Post to feed (without cleanup yet - story needs the same image)
	postID, err := c.postImage(imagePath, tournament)
	if err != nil {
		CleanupImage(imagePath) // Cleanup on error
		notification.Error = fmt.Sprintf("failed to post image to feed: %v", err)
		return notification, err
	}

	log.Printf("✅ Posted to feed - Post ID: %s", postID)

	// Post to story (reusing same image, with clickable link)
	storyID, err := c.postStory(imagePath, tournament.TournamentURL)
	if err != nil {
		notification.Error = fmt.Sprintf("failed to post to story (feed succeeded): %v", err)
		log.Printf("⚠️  Warning: Story posting failed but feed post succeeded: %v", err)
		// Don't return error - feed post succeeded
	} else {
		log.Printf("✅ Posted to story - Story ID: %s", storyID)
	}

	// Post to Threads
	if c.config.ThreadsEnabled {
		threadID, err := c.postThread(imagePath, tournament)
		if err != nil {
			log.Printf("⚠️  Warning: Threads posting failed: %v", err)
			// Don't fail - Instagram posts succeeded
		} else {
			log.Printf("✅ Posted to Threads - Thread ID: %s", threadID)
		}
	}

	// Cleanup after ALL posts are done
	log.Printf("⏳ Waiting 30 seconds for platforms to finalize...")
	time.Sleep(30 * time.Second)

	if err := os.Remove(imagePath); err != nil {
		log.Printf("⚠️  Warning: failed to cleanup image %s: %v", imagePath, err)
	} else {
		log.Printf("🗑️  Cleaned up local image: %s", imagePath)
	}

	notification.MessageID = postID
	notification.Success = true

	log.Printf("Successfully posted tournament %d (%s) to Instagram and Threads",
		tournament.TournamentID, tournament.Name)

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

	// Step 2: Wait for container to be ready (Instagram needs time to process)
	log.Printf("⏳ Waiting for Instagram to process the image...")
	if err := c.waitForContainerReady(containerID); err != nil {
		return "", fmt.Errorf("failed waiting for container: %w", err)
	}

	log.Printf("✅ Container ready for publishing")

	// Step 3: Publish the container
	postID, err := c.publishMediaContainer(containerID)
	if err != nil {
		return "", fmt.Errorf("failed to publish media: %w", err)
	}

	log.Printf("✅ Post published: %s", postID)

	return postID, nil
}

// postStory posts an image to Instagram Story with a clickable link
func (c *Client) postStory(imagePath string, tournamentURL string) (string, error) {
	// Get the image URL (same logic as feed posts)
	var imageURL string
	ginMode := os.Getenv("GIN_MODE")

	if ginMode == "release" {
		// Production: use our server
		filename := filepath.Base(imagePath)
		imageURL = fmt.Sprintf("https://tournois-tt.fr/instagram-images/%s", filename)
	} else {
		// Development: upload to Catbox.moe
		uploadedURL, err := uploadToImgBB(imagePath)
		if err != nil {
			return "", fmt.Errorf("failed to upload image for story: %w", err)
		}
		imageURL = uploadedURL
	}

	log.Printf("📸 Story image URL: %s", imageURL)
	log.Printf("🔗 Story link: %s", tournamentURL)

	// Step 1: Create story container with link sticker
	createURL := fmt.Sprintf("%s/%s/media?image_url=%s&media_type=STORIES&link=%s&access_token=%s",
		GraphAPIBaseURL,
		c.config.PageID,
		url.QueryEscape(imageURL),
		url.QueryEscape(tournamentURL),
		url.QueryEscape(c.config.AccessToken),
	)

	resp, err := c.httpClient.Post(createURL, "application/json", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create story container: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if json.Unmarshal(body, &errResp) == nil {
			return "", fmt.Errorf("instagram API error: %s (code: %d)", errResp.Error.Message, errResp.Error.Code)
		}
		return "", fmt.Errorf("create story container failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse story container response: %w", err)
	}

	containerID := result.ID
	log.Printf("✅ Story container created: %s", containerID)

	// Step 2: Wait for container to be ready
	log.Printf("⏳ Waiting for Instagram to process the story image...")
	if err := c.waitForContainerReady(containerID); err != nil {
		return "", fmt.Errorf("failed waiting for story container: %w", err)
	}

	log.Printf("✅ Story container ready for publishing")

	// Step 3: Publish the story
	storyID, err := c.publishMediaContainer(containerID)
	if err != nil {
		return "", fmt.Errorf("failed to publish story: %w", err)
	}

	return storyID, nil
}

// postThread posts a tournament image and info to Threads
func (c *Client) postThread(imagePath string, tournament TournamentImage) (string, error) {
	if !c.config.ThreadsEnabled {
		return "", fmt.Errorf("threads posting is disabled")
	}

	if c.config.ThreadsUserID == "" {
		return "", fmt.Errorf("threads user ID is not configured")
	}

	// Get the image URL (same logic as Instagram posts)
	var imageURL string
	ginMode := os.Getenv("GIN_MODE")

	if ginMode == "release" {
		// Production: use our server
		filename := filepath.Base(imagePath)
		imageURL = fmt.Sprintf("https://tournois-tt.fr/instagram-images/%s", filename)
	} else {
		// Development: upload to Catbox.moe
		uploadedURL, err := uploadToImgBB(imagePath)
		if err != nil {
			return "", fmt.Errorf("failed to upload image for thread: %w", err)
		}
		imageURL = uploadedURL
	}

	// Prepare thread text (in French)
	threadText := fmt.Sprintf(`🏓 %s

🏆 Type: %s
💰 Dotation: %d €
📅 %s

📍 %s

%s

🗺️ Découvrez d'autres tournois sur la carte : https://tournois-tt.fr

#TennisDeTable #PingPong #FFTT`,
		tournament.Name,
		tournament.Type,
		tournament.Endowment/100,
		formatDates(tournament.StartDate, tournament.EndDate),
		tournament.Address,
		tournament.TournamentURL,
	)

	// Add inscription link if available
	if tournament.Page != "" {
		threadText = fmt.Sprintf(`%s

✍️ Inscription : %s`, threadText, tournament.Page)
	}

	log.Printf("📸 Thread image URL: %s", imageURL)
	log.Printf("📝 Thread text length: %d characters", len(threadText))

	// Step 1: Create thread container
	createURL := fmt.Sprintf("%s/%s/threads?media_type=IMAGE&image_url=%s&text=%s&access_token=%s",
		ThreadsAPIBaseURL,
		c.config.ThreadsUserID,
		url.QueryEscape(imageURL),
		url.QueryEscape(threadText),
		url.QueryEscape(c.config.ThreadsAccessToken),
	)

	resp, err := c.httpClient.Post(createURL, "application/json", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create thread container: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if json.Unmarshal(body, &errResp) == nil {
			return "", fmt.Errorf("threads API error: %s (code: %d)", errResp.Error.Message, errResp.Error.Code)
		}
		return "", fmt.Errorf("create thread container failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse thread container response: %w", err)
	}

	containerID := result.ID
	log.Printf("✅ Thread container created: %s", containerID)

	// Step 2: Wait for container to be ready
	log.Printf("⏳ Waiting for Threads to process the image...")
	if err := c.waitForThreadContainerReady(containerID); err != nil {
		return "", fmt.Errorf("failed waiting for thread container: %w", err)
	}

	log.Printf("✅ Thread container ready for publishing")

	// Step 3: Publish the thread
	publishURL := fmt.Sprintf("%s/%s/threads_publish?creation_id=%s&access_token=%s",
		ThreadsAPIBaseURL,
		c.config.ThreadsUserID,
		url.QueryEscape(containerID),
		url.QueryEscape(c.config.ThreadsAccessToken),
	)

	resp2, err := c.httpClient.Post(publishURL, "application/json", nil)
	if err != nil {
		return "", fmt.Errorf("failed to publish thread: %w", err)
	}
	defer resp2.Body.Close()

	body2, _ := io.ReadAll(resp2.Body)

	if resp2.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if json.Unmarshal(body2, &errResp) == nil {
			return "", fmt.Errorf("threads API error: %s (code: %d)", errResp.Error.Message, errResp.Error.Code)
		}
		return "", fmt.Errorf("publish thread failed with status %d: %s", resp2.StatusCode, string(body2))
	}

	var publishResult struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body2, &publishResult); err != nil {
		return "", fmt.Errorf("failed to parse thread publish response: %w", err)
	}

	return publishResult.ID, nil
}

// waitForThreadContainerReady polls the thread container status until it's ready
func (c *Client) waitForThreadContainerReady(containerID string) error {
	maxAttempts := 30 // 30 attempts = up to 60 seconds

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Check container status
		statusURL := fmt.Sprintf("%s/%s?fields=status&access_token=%s",
			ThreadsAPIBaseURL,
			containerID,
			url.QueryEscape(c.config.ThreadsAccessToken),
		)

		resp, err := c.httpClient.Get(statusURL)
		if err != nil {
			log.Printf("⚠️  Attempt %d/%d: Failed to check thread status: %v", attempt, maxAttempts, err)
			time.Sleep(2 * time.Second)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("⚠️  Attempt %d/%d: Thread status check returned %d: %s", attempt, maxAttempts, resp.StatusCode, string(body))
			time.Sleep(2 * time.Second)
			continue
		}

		// Parse status
		var result struct {
			Status string `json:"status"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			log.Printf("⚠️  Attempt %d/%d: Failed to parse thread status: %v", attempt, maxAttempts, err)
			time.Sleep(2 * time.Second)
			continue
		}

		log.Printf("📊 Thread container status: %s (attempt %d/%d)", result.Status, attempt, maxAttempts)

		switch result.Status {
		case "FINISHED":
			return nil // Ready to publish!
		case "ERROR":
			return fmt.Errorf("thread container processing failed with ERROR status")
		case "EXPIRED":
			return fmt.Errorf("thread container expired before it could be published")
		case "IN_PROGRESS":
			// Continue waiting
			time.Sleep(2 * time.Second)
		default:
			log.Printf("⚠️  Unknown thread status: %s, continuing to wait...", result.Status)
			time.Sleep(2 * time.Second)
		}
	}

	return fmt.Errorf("timeout waiting for thread container to be ready after %d attempts", maxAttempts)
}

// createMediaContainer creates a media container with the image
// Note: Instagram requires the image to be accessible via a public HTTPS URL
func (c *Client) createMediaContainer(imagePath string, tournament TournamentImage) (string, error) {
	// Prepare caption (convert endowment from cents to euros)
	caption := fmt.Sprintf(`🏓 %s

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
			return "", fmt.Errorf("instagram API error: %s (code: %d)", errResp.Error.Message, errResp.Error.Code)
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
			return "", fmt.Errorf("instagram API error: %s (code: %d)", errResp.Error.Message, errResp.Error.Code)
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

// waitForContainerReady polls the container status until it's ready to publish
func (c *Client) waitForContainerReady(containerID string) error {
	maxAttempts := 30 // 30 attempts = up to 60 seconds

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Check container status
		statusURL := fmt.Sprintf("%s/%s?fields=status_code&access_token=%s",
			GraphAPIBaseURL,
			containerID,
			url.QueryEscape(c.config.AccessToken),
		)

		resp, err := c.httpClient.Get(statusURL)
		if err != nil {
			log.Printf("⚠️  Attempt %d/%d: Failed to check status: %v", attempt, maxAttempts, err)
			time.Sleep(2 * time.Second)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("⚠️  Attempt %d/%d: Status check returned %d: %s", attempt, maxAttempts, resp.StatusCode, string(body))
			time.Sleep(2 * time.Second)
			continue
		}

		// Parse status
		var result struct {
			StatusCode string `json:"status_code"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			log.Printf("⚠️  Attempt %d/%d: Failed to parse status: %v", attempt, maxAttempts, err)
			time.Sleep(2 * time.Second)
			continue
		}

		log.Printf("📊 Container status: %s (attempt %d/%d)", result.StatusCode, attempt, maxAttempts)

		switch result.StatusCode {
		case "FINISHED":
			return nil // Ready to publish!
		case "ERROR":
			return fmt.Errorf("container processing failed with ERROR status")
		case "EXPIRED":
			return fmt.Errorf("container expired before it could be published")
		case "IN_PROGRESS":
			// Continue waiting
			time.Sleep(2 * time.Second)
		default:
			log.Printf("⚠️  Unknown status: %s, continuing to wait...", result.StatusCode)
			time.Sleep(2 * time.Second)
		}
	}

	return fmt.Errorf("timeout waiting for container to be ready after %d attempts", maxAttempts)
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
		return fmt.Errorf("instagram integration is disabled")
	}

	// Validate credentials are present
	if c.config.AccessToken == "" {
		return fmt.Errorf("instagram access token is not configured")
	}
	if c.config.PageID == "" {
		return fmt.Errorf("instagram page ID is not configured")
	}

	// Test by getting account info
	testURL := fmt.Sprintf("%s/%s?fields=id,username&access_token=%s",
		GraphAPIBaseURL, c.config.PageID, c.config.AccessToken)

	resp, err := c.httpClient.Get(testURL)
	if err != nil {
		return fmt.Errorf("failed to connect to instagram API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("instagram API test failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
