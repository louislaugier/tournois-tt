package api

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
    "strings"
    "time"

    igimage "tournois-tt/api/pkg/image"
    "tournois-tt/api/pkg/instagram"
    "tournois-tt/api/pkg/utils"
)

const (
    GraphAPIBaseURL    = "https://graph.instagram.com/v18.0"
    ThreadsAPIBaseURL  = "https://graph.threads.net/v1.0"
)

type Client struct {
    config     Config
    httpClient *http.Client
}

func NewClient(config Config) *Client {
    if token, err := LoadToken(); err == nil && token != "" {
        config.AccessToken = token
    }
    if token, err := LoadThreadsToken(); err == nil && token != "" {
        config.ThreadsAccessToken = token
    }
    return &Client{config: config, httpClient: &http.Client{Timeout: 30 * time.Second}}
}

func (c *Client) PostTournament(tournament igimage.TournamentImage) (*TournamentNotification, error) {
    return c.postTournament(tournament, false)
}

func (c *Client) PostTournamentStoryOnly(tournament igimage.TournamentImage) (*TournamentNotification, error) {
    return c.postTournament(tournament, true)
}

func (c *Client) postTournament(tournament igimage.TournamentImage, storyOnly bool) (*TournamentNotification, error) {
    if !c.config.Enabled {
        return nil, fmt.Errorf("instagram posting is disabled in configuration")
    }
    if c.config.AccessToken == "" {
        return nil, fmt.Errorf("instagram access token is not configured")
    }
    if c.config.PageID == "" {
        return nil, fmt.Errorf("instagram page ID is not configured")
    }

    // Skip duplicate check if story-only (allow reposting as story)
    if !storyOnly {
        // Check cache first (fast, no API calls)
        cache := instagram.GetPostedCache()
        if posted, record := cache.IsPosted(tournament.TournamentID); posted {
            log.Printf("⚠️  Tournament %d already posted on %s - skipping", tournament.TournamentID, record.PostedAt.Format("2006-01-02 15:04:05"))
            log.Printf("   Posted to: Feed=%v, Story=%v, Threads=%v", record.InstagramFeed, record.InstagramStory, record.Threads)
            return nil, fmt.Errorf("tournament %d already posted (from cache)", tournament.TournamentID)
        }
        
        // Double-check with API (in case cache is outdated)
        log.Printf("🔍 Checking Instagram/Threads APIs for tournament %d...", tournament.TournamentID)
        alreadyPosted, platform, err := c.isTournamentAlreadyPosted(tournament)
        if err != nil {
            log.Printf("⚠️  Warning: Could not check APIs for duplicates: %v", err)
            // Continue anyway - cache check already passed
        } else if alreadyPosted {
            log.Printf("⚠️  Tournament %d found on %s (updating cache) - skipping", tournament.TournamentID, platform)
            // Update cache with this finding
            cache.MarkPosted(&instagram.PostedRecord{
                TournamentID:   tournament.TournamentID,
                TournamentName: tournament.Name,
                PostedAt:       time.Now(),
                InstagramFeed:  true,
                InstagramStory: false,
                Threads:        platform == "Threads",
            })
            return nil, fmt.Errorf("tournament %d already posted on %s", tournament.TournamentID, platform)
        }
        log.Printf("✅ Tournament %d not yet posted - proceeding", tournament.TournamentID)
    } else {
        log.Printf("📱 Posting story only for tournament %d", tournament.TournamentID)
    }

    notification := &TournamentNotification{Tournament: tournament, SentAt: time.Now(), Success: false}

    var imagePath, storyImagePath string
    var postID string
    var err error

    if !storyOnly {
        // Generate feed image (1080x1080)
        imagePath, err = igimage.GenerateTournamentImage(tournament)
        if err != nil {
            notification.Error = fmt.Sprintf("failed to generate feed image: %v", err)
            return notification, err
        }
    }

    // Generate story image (1080x1920)
    storyImagePath, err = igimage.GenerateTournamentStoryImage(tournament)
    if err != nil {
        if imagePath != "" {
            _ = os.Remove(imagePath)
        }
        notification.Error = fmt.Sprintf("failed to generate story image: %v", err)
        return notification, err
    }

    // Post to feed only if not story-only
    if !storyOnly {
        postID, err = c.postImage(imagePath, tournament)
        if err != nil {
            _ = os.Remove(imagePath)
            _ = os.Remove(storyImagePath)
            notification.Error = fmt.Sprintf("failed to post image to feed: %v", err)
            return notification, err
        }
        log.Printf("✅ Posted to feed - Post ID: %s", postID)
    }

    // Post to story
    storyID, storyErr := c.postStory(storyImagePath, tournament.TournamentURL)
    if storyErr != nil {
        if !storyOnly {
            notification.Error = fmt.Sprintf("failed to post to story (feed succeeded): %v", storyErr)
            log.Printf("⚠️  Warning: Story posting failed but feed post succeeded: %v", storyErr)
        } else {
            notification.Error = fmt.Sprintf("failed to post to story: %v", storyErr)
            return notification, storyErr
        }
    } else {
        log.Printf("✅ Posted to story - Story ID: %s", storyID)
        if storyOnly {
            postID = storyID // Use story ID as the main ID for story-only posts
        }
    }

    // Post to Threads only if not story-only
    if !storyOnly && c.config.ThreadsEnabled {
        if threadID, err := c.postThread(imagePath, tournament); err != nil {
            log.Printf("⚠️  Warning: Threads posting failed: %v", err)
        } else {
            log.Printf("✅ Posted to Threads - Thread ID: %s", threadID)
        }
    }

    log.Printf("⏳ Waiting 30 seconds for platforms to finalize...")
    time.Sleep(30 * time.Second)

    // Cleanup images
    if imagePath != "" {
        if err := os.Remove(imagePath); err != nil {
            log.Printf("⚠️  Warning: failed to cleanup feed image %s: %v", imagePath, err)
        } else {
            log.Printf("🗑️  Cleaned up feed image: %s", imagePath)
        }
    }
    
    if err := os.Remove(storyImagePath); err != nil {
        log.Printf("⚠️  Warning: failed to cleanup story image %s: %v", storyImagePath, err)
    } else {
        log.Printf("🗑️  Cleaned up story image: %s", storyImagePath)
    }

    notification.MessageID = postID
    notification.Success = true
    
    // Save to cache
    cache := instagram.GetPostedCache()
    record := &instagram.PostedRecord{
        TournamentID:     tournament.TournamentID,
        TournamentName:   tournament.Name,
        PostedAt:         time.Now(),
        InstagramFeed:    !storyOnly && postID != "",
        InstagramStory:   storyErr == nil,
        Threads:          !storyOnly && c.config.ThreadsEnabled,
        InstagramPostID:  postID,
        InstagramStoryID: storyID,
    }
    
    if err := cache.MarkPosted(record); err != nil {
        log.Printf("⚠️  Warning: Failed to save to cache: %v", err)
    }
    
    if storyOnly {
        log.Printf("✅ Successfully posted tournament %d (%s) to Instagram Story", tournament.TournamentID, tournament.Name)
    } else {
        log.Printf("✅ Successfully posted tournament %d (%s) to Instagram and Threads", tournament.TournamentID, tournament.Name)
    }
    return notification, nil
}

func (c *Client) postImage(imagePath string, tournament igimage.TournamentImage) (string, error) {
    containerID, err := c.createMediaContainer(imagePath, tournament)
    if err != nil { return "", fmt.Errorf("failed to create media container: %w", err) }
    log.Printf("✅ Media container created: %s", containerID)
    log.Printf("⏳ Waiting for Instagram to process the image...")
    if err := c.waitForContainerReady(containerID); err != nil { return "", fmt.Errorf("failed waiting for container: %w", err) }
    log.Printf("✅ Container ready for publishing")
    postID, err := c.publishMediaContainer(containerID)
    if err != nil { return "", fmt.Errorf("failed to publish media: %w", err) }
    log.Printf("✅ Post published: %s", postID)
    return postID, nil
}

func (c *Client) postStory(imagePath string, tournamentURL string) (string, error) {
    var imageURL string
    ginMode := os.Getenv("GIN_MODE")
    if ginMode == "release" {
        filename := filepath.Base(imagePath)
        imageURL = fmt.Sprintf("https://tournois-tt.fr/instagram-images/%s", filename)
    } else {
        uploadedURL, err := uploadToImgBB(imagePath)
        if err != nil { return "", fmt.Errorf("failed to upload image for story: %w", err) }
        imageURL = uploadedURL
    }
    log.Printf("📸 Story image URL: %s", imageURL)
    // Note: Swipe-up links require 10k+ followers or verified account
    // URL is displayed in the image itself instead
    createURL := fmt.Sprintf("%s/%s/media?image_url=%s&media_type=STORIES&access_token=%s", GraphAPIBaseURL, c.config.PageID, url.QueryEscape(imageURL), url.QueryEscape(c.config.AccessToken))
    resp, err := c.httpClient.Post(createURL, "application/json", nil)
    if err != nil { return "", fmt.Errorf("failed to create story container: %w", err) }
    defer resp.Body.Close()
    body, _ := io.ReadAll(resp.Body)
    if resp.StatusCode != http.StatusOK { var errResp ErrorResponse; if json.Unmarshal(body, &errResp) == nil { return "", fmt.Errorf("instagram API error: %s (code: %d)", errResp.Error.Message, errResp.Error.Code) }; return "", fmt.Errorf("create story container failed with status %d: %s", resp.StatusCode, string(body)) }
    var result struct{ ID string `json:"id"` }
    if err := json.Unmarshal(body, &result); err != nil { return "", fmt.Errorf("failed to parse story container response: %w", err) }
    containerID := result.ID
    log.Printf("✅ Story container created: %s", containerID)
    log.Printf("⏳ Waiting for Instagram to process the story image...")
    if err := c.waitForContainerReady(containerID); err != nil { return "", fmt.Errorf("failed waiting for story container: %w", err) }
    log.Printf("✅ Story container ready for publishing")
    storyID, err := c.publishMediaContainer(containerID)
    if err != nil { return "", fmt.Errorf("failed to publish story: %w", err) }
    return storyID, nil
}

func (c *Client) postThread(imagePath string, tournament igimage.TournamentImage) (string, error) {
    if !c.config.ThreadsEnabled { return "", fmt.Errorf("threads posting is disabled") }
    if c.config.ThreadsUserID == "" { return "", fmt.Errorf("threads user ID is not configured") }
    var imageURL string
    ginMode := os.Getenv("GIN_MODE")
    if ginMode == "release" {
        filename := filepath.Base(imagePath)
        imageURL = fmt.Sprintf("https://tournois-tt.fr/instagram-images/%s", filename)
    } else {
        uploadedURL, err := uploadToImgBB(imagePath)
        if err != nil { return "", fmt.Errorf("failed to upload image for thread: %w", err) }
        imageURL = uploadedURL
    }
    threadText := fmt.Sprintf(`🏓 %s

🏆 Type: %s
🏓 Club: %s
💰 Dotation: %d €
📅 %s

📍 %s

%s

🗺️ Découvrez d'autres tournois sur la carte : https://tournois-tt.fr

#TennisDeTable #PingPong #FFTT`,
        tournament.Name,
        utils.MapTournamentType(tournament.Type),
        tournament.Club,
        tournament.Endowment/100,
        formatDatesLocal(tournament.StartDate, tournament.EndDate),
        tournament.Address,
        tournament.TournamentURL,
    )
    if tournament.Page != "" { threadText = fmt.Sprintf(`%s

✍️ Inscription : %s`, threadText, tournament.Page) }
    log.Printf("📸 Thread image URL: %s", imageURL)
    log.Printf("📝 Thread text length: %d characters", len(threadText))
    createURL := fmt.Sprintf("%s/%s/threads?media_type=IMAGE&image_url=%s&text=%s&access_token=%s", ThreadsAPIBaseURL, c.config.ThreadsUserID, url.QueryEscape(imageURL), url.QueryEscape(threadText), url.QueryEscape(c.config.ThreadsAccessToken))
    resp, err := c.httpClient.Post(createURL, "application/json", nil)
    if err != nil { return "", fmt.Errorf("failed to create thread container: %w", err) }
    defer resp.Body.Close()
    body, _ := io.ReadAll(resp.Body)
    if resp.StatusCode != http.StatusOK { var errResp ErrorResponse; if json.Unmarshal(body, &errResp) == nil { return "", fmt.Errorf("threads API error: %s (code: %d)", errResp.Error.Message, errResp.Error.Code) }; return "", fmt.Errorf("create thread container failed with status %d: %s", resp.StatusCode, string(body)) }
    var result struct{ ID string `json:"id"` }
    if err := json.Unmarshal(body, &result); err != nil { return "", fmt.Errorf("failed to parse thread container response: %w", err) }
    containerID := result.ID
    log.Printf("✅ Thread container created: %s", containerID)
    log.Printf("⏳ Waiting for Threads to process the image...")
    if err := c.waitForThreadContainerReady(containerID); err != nil { return "", fmt.Errorf("failed waiting for thread container: %w", err) }
    log.Printf("✅ Thread container ready for publishing")
    publishURL := fmt.Sprintf("%s/%s/threads_publish?creation_id=%s&access_token=%s", ThreadsAPIBaseURL, c.config.ThreadsUserID, url.QueryEscape(containerID), url.QueryEscape(c.config.ThreadsAccessToken))
    resp2, err := c.httpClient.Post(publishURL, "application/json", nil)
    if err != nil { return "", fmt.Errorf("failed to publish thread: %w", err) }
    defer resp2.Body.Close()
    body2, _ := io.ReadAll(resp2.Body)
    if resp2.StatusCode != http.StatusOK { var errResp ErrorResponse; if json.Unmarshal(body2, &errResp) == nil { return "", fmt.Errorf("threads API error: %s (code: %d)", errResp.Error.Message, errResp.Error.Code) }; return "", fmt.Errorf("publish thread failed with status %d: %s", resp2.StatusCode, string(body2)) }
    var publishResult struct{ ID string `json:"id"` }
    if err := json.Unmarshal(body2, &publishResult); err != nil { return "", fmt.Errorf("failed to parse thread publish response: %w", err) }
    return publishResult.ID, nil
}

func (c *Client) waitForThreadContainerReady(containerID string) error {
    maxAttempts := 30
    for attempt := 1; attempt <= maxAttempts; attempt++ {
        statusURL := fmt.Sprintf("%s/%s?fields=status&access_token=%s", ThreadsAPIBaseURL, containerID, url.QueryEscape(c.config.ThreadsAccessToken))
        resp, err := c.httpClient.Get(statusURL)
        if err != nil { log.Printf("⚠️  Attempt %d/%d: Failed to check thread status: %v", attempt, maxAttempts, err); time.Sleep(2*time.Second); continue }
        body, _ := io.ReadAll(resp.Body)
        resp.Body.Close()
        if resp.StatusCode != http.StatusOK { log.Printf("⚠️  Attempt %d/%d: Thread status check returned %d: %s", attempt, maxAttempts, resp.StatusCode, string(body)); time.Sleep(2*time.Second); continue }
        var result struct{ Status string `json:"status"` }
        if err := json.Unmarshal(body, &result); err != nil { log.Printf("⚠️  Attempt %d/%d: Failed to parse thread status: %v", attempt, maxAttempts, err); time.Sleep(2*time.Second); continue }
        switch result.Status { case "FINISHED": return nil; case "ERROR": return fmt.Errorf("thread container processing failed with ERROR status"); case "EXPIRED": return fmt.Errorf("thread container expired before it could be published"); case "IN_PROGRESS": time.Sleep(2*time.Second); default: time.Sleep(2*time.Second) }
    }
    return fmt.Errorf("timeout waiting for thread container to be ready after %d attempts", maxAttempts)
}

func (c *Client) createMediaContainer(imagePath string, tournament igimage.TournamentImage) (string, error) {
    caption := fmt.Sprintf(`🏓 %s

🏆 Type: %s
🏓 Club: %s
💰 Dotation: %d €
📅 %s
📍 %s

🔗 Règlement: %s

#TennisDeTable #PingPong #FFTT #Tournoi`,
        tournament.Name,
        utils.MapTournamentType(tournament.Type),
        tournament.Club,
        tournament.Endowment/100,
        formatDatesLocal(tournament.StartDate, tournament.EndDate),
        tournament.Address,
        tournament.TournamentURL,
    )
    var imageURL string
    ginMode := os.Getenv("GIN_MODE")
    if ginMode == "release" { filename := filepath.Base(imagePath); imageURL = fmt.Sprintf("https://tournois-tt.fr/instagram-images/%s", filename) } else { log.Println("📸 [DEV] Uploading image to Catbox.moe for testing..."); uploadedURL, err := uploadToImgBB(imagePath); if err != nil { return "", fmt.Errorf("failed to upload image to Catbox.moe: %w", err) }; imageURL = uploadedURL }
    createURL := fmt.Sprintf("%s/%s/media?image_url=%s&caption=%s&access_token=%s", GraphAPIBaseURL, c.config.PageID, url.QueryEscape(imageURL), url.QueryEscape(caption), url.QueryEscape(c.config.AccessToken))
    resp, err := c.httpClient.Post(createURL, "application/json", nil); if err != nil { return "", fmt.Errorf("failed to create container: %w", err) }
    defer resp.Body.Close()
    body, _ := io.ReadAll(resp.Body)
    if resp.StatusCode != http.StatusOK { var errResp ErrorResponse; if json.Unmarshal(body, &errResp) == nil { return "", fmt.Errorf("instagram API error: %s (code: %d)", errResp.Error.Message, errResp.Error.Code) }; return "", fmt.Errorf("create container failed with status %d: %s", resp.StatusCode, string(body)) }
    var result struct{ ID string `json:"id"` }
    if err := json.Unmarshal(body, &result); err != nil { return "", fmt.Errorf("failed to parse container response: %w", err) }
    return result.ID, nil
}

func (c *Client) publishMediaContainer(containerID string) (string, error) {
    publishURL := fmt.Sprintf("%s/%s/media_publish?creation_id=%s&access_token=%s", GraphAPIBaseURL, c.config.PageID, url.QueryEscape(containerID), url.QueryEscape(c.config.AccessToken))
    resp, err := c.httpClient.Post(publishURL, "application/json", nil); if err != nil { return "", fmt.Errorf("failed to publish: %w", err) }
    defer resp.Body.Close()
    body, _ := io.ReadAll(resp.Body)
    if resp.StatusCode != http.StatusOK { var errResp ErrorResponse; if json.Unmarshal(body, &errResp) == nil { return "", fmt.Errorf("instagram API error: %s (code: %d)", errResp.Error.Message, errResp.Error.Code) }; return "", fmt.Errorf("publish failed with status %d: %s", resp.StatusCode, string(body)) }
    var result struct{ ID string `json:"id"` }
    if err := json.Unmarshal(body, &result); err != nil { return "", fmt.Errorf("failed to parse publish response: %w", err) }
    return result.ID, nil
}

func (c *Client) waitForContainerReady(containerID string) error {
    maxAttempts := 30
    for attempt := 1; attempt <= maxAttempts; attempt++ {
        statusURL := fmt.Sprintf("%s/%s?fields=status_code&access_token=%s", GraphAPIBaseURL, containerID, url.QueryEscape(c.config.AccessToken))
        resp, err := c.httpClient.Get(statusURL)
        if err != nil { log.Printf("⚠️  Attempt %d/%d: Failed to check status: %v", attempt, maxAttempts, err); time.Sleep(2*time.Second); continue }
        body, _ := io.ReadAll(resp.Body); resp.Body.Close()
        if resp.StatusCode != http.StatusOK { log.Printf("⚠️  Attempt %d/%d: Status check returned %d: %s", attempt, maxAttempts, resp.StatusCode, string(body)); time.Sleep(2*time.Second); continue }
        var result struct{ StatusCode string `json:"status_code"` }
        if err := json.Unmarshal(body, &result); err != nil { log.Printf("⚠️  Attempt %d/%d: Failed to parse status: %v", attempt, maxAttempts, err); time.Sleep(2*time.Second); continue }
        switch result.StatusCode { case "FINISHED": return nil; case "ERROR": return fmt.Errorf("container processing failed with ERROR status"); case "EXPIRED": return fmt.Errorf("container expired before it could be published"); case "IN_PROGRESS": time.Sleep(2*time.Second); default: time.Sleep(2*time.Second) }
    }
    return fmt.Errorf("timeout waiting for container to be ready after %d attempts", maxAttempts)
}

func uploadToImgBB(imagePath string) (string, error) {
    file, err := os.Open(imagePath); if err != nil { return "", fmt.Errorf("failed to open image: %w", err) }
    defer file.Close()
    var requestBody bytes.Buffer
    writer := multipart.NewWriter(&requestBody)
    part, err := writer.CreateFormFile("fileToUpload", filepath.Base(imagePath)); if err != nil { return "", fmt.Errorf("failed to create form file: %w", err) }
    if _, err := io.Copy(part, file); err != nil { return "", fmt.Errorf("failed to copy file: %w", err) }
    if err := writer.WriteField("reqtype", "fileupload"); err != nil { return "", fmt.Errorf("failed to write reqtype field: %w", err) }
    if err := writer.Close(); err != nil { return "", fmt.Errorf("failed to close writer: %w", err) }
    uploadURL := "https://catbox.moe/user/api.php"
    req, err := http.NewRequest("POST", uploadURL, &requestBody); if err != nil { return "", fmt.Errorf("failed to create request: %w", err) }
    req.Header.Set("Content-Type", writer.FormDataContentType())
    client := &http.Client{Timeout: 30 * time.Second}
    resp, err := client.Do(req); if err != nil { return "", fmt.Errorf("failed to upload: %w", err) }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body); if err != nil { return "", fmt.Errorf("failed to read response: %w", err) }
    if resp.StatusCode != http.StatusOK { return "", fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body)) }
    imageURL := string(bytes.TrimSpace(body))
    if imageURL == "" { return "", fmt.Errorf("upload succeeded but no URL returned") }
    return imageURL, nil
}

func formatDate(dateStr string) string {
    t, err := time.Parse(time.RFC3339, dateStr)
    if err != nil { return dateStr }
    return t.Format("02/01/2006")
}

// formatDatesLocal mirrors the date formatting used for captions/threads
func formatDatesLocal(startDate, endDate string) string {
    if startDate == "" { return "Date non disponible" }
    parse := func(layouts []string, s string) (time.Time, bool) {
        for _, l := range layouts { if t, err := time.Parse(l, s); err == nil { return t, true } }
        return time.Time{}, false
    }
    layouts := []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02"}
    start, ok := parse(layouts, startDate); if !ok { return startDate }
    if endDate == "" || endDate == startDate { return start.Format("02/01/2006") }
    end, ok := parse(layouts, endDate); if !ok { return start.Format("02/01/2006") }
    if start.Month() == end.Month() && start.Year() == end.Year() { months := []string{"janvier","février","mars","avril","mai","juin","juillet","août","septembre","octobre","novembre","décembre"}; return fmt.Sprintf("%d-%d %s %d", start.Day(), end.Day(), months[start.Month()-1], start.Year()) }
    return fmt.Sprintf("%s - %s", start.Format("02/01/2006"), end.Format("02/01/2006"))
}

func (c *Client) TestConnection() error {
    if !c.config.Enabled { return fmt.Errorf("instagram integration is disabled") }
    if c.config.AccessToken == "" { return fmt.Errorf("instagram access token is not configured") }
    if c.config.PageID == "" { return fmt.Errorf("instagram page ID is not configured") }
    testURL := fmt.Sprintf("%s/%s?fields=id,username&access_token=%s", GraphAPIBaseURL, c.config.PageID, c.config.AccessToken)
    resp, err := c.httpClient.Get(testURL); if err != nil { return fmt.Errorf("failed to connect to instagram API: %w", err) }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK { body, _ := io.ReadAll(resp.Body); return fmt.Errorf("instagram API test failed with status %d: %s", resp.StatusCode, string(body)) }
    return nil
}

// isTournamentAlreadyPosted checks if a tournament has already been posted to Instagram or Threads
func (c *Client) isTournamentAlreadyPosted(tournament igimage.TournamentImage) (bool, string, error) {
    // Check Instagram posts
    if posted, err := c.checkInstagramPosts(tournament); err != nil {
        log.Printf("⚠️  Could not check Instagram posts: %v", err)
    } else if posted {
        return true, "Instagram", nil
    }

    // Check Threads posts if enabled
    if c.config.ThreadsEnabled && c.config.ThreadsUserID != "" {
        if posted, err := c.checkThreadsPosts(tournament); err != nil {
            log.Printf("⚠️  Could not check Threads posts: %v", err)
        } else if posted {
            return true, "Threads", nil
        }
    }

    return false, "", nil
}

// checkInstagramPosts fetches recent Instagram posts and checks for tournament
func (c *Client) checkInstagramPosts(tournament igimage.TournamentImage) (bool, error) {
    // Fetch recent media posts from Instagram
    mediaURL := fmt.Sprintf("%s/%s/media?fields=id,caption,timestamp&limit=50&access_token=%s", 
        GraphAPIBaseURL, c.config.PageID, url.QueryEscape(c.config.AccessToken))
    
    resp, err := c.httpClient.Get(mediaURL)
    if err != nil {
        return false, fmt.Errorf("failed to fetch Instagram posts: %w", err)
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)
    if resp.StatusCode != http.StatusOK {
        return false, fmt.Errorf("Instagram API returned status %d: %s", resp.StatusCode, string(body))
    }

    var result struct {
        Data []struct {
            ID      string `json:"id"`
            Caption string `json:"caption"`
        } `json:"data"`
    }

    if err := json.Unmarshal(body, &result); err != nil {
        return false, fmt.Errorf("failed to parse Instagram posts: %w", err)
    }

    // Check if any post contains the tournament name or URL
    return containsTournament(result.Data, tournament), nil
}

// checkThreadsPosts fetches recent Threads posts and checks for tournament
func (c *Client) checkThreadsPosts(tournament igimage.TournamentImage) (bool, error) {
    // Fetch recent threads from Threads API
    threadsURL := fmt.Sprintf("%s/%s/threads?fields=id,text,timestamp&limit=50&access_token=%s",
        ThreadsAPIBaseURL, c.config.ThreadsUserID, url.QueryEscape(c.config.ThreadsAccessToken))
    
    resp, err := c.httpClient.Get(threadsURL)
    if err != nil {
        return false, fmt.Errorf("failed to fetch Threads posts: %w", err)
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)
    if resp.StatusCode != http.StatusOK {
        return false, fmt.Errorf("Threads API returned status %d: %s", resp.StatusCode, string(body))
    }

    var result struct {
        Data []struct {
            ID   string `json:"id"`
            Text string `json:"text"`
        } `json:"data"`
    }

    if err := json.Unmarshal(body, &result); err != nil {
        return false, fmt.Errorf("failed to parse Threads posts: %w", err)
    }

    // Convert to common format for checking
    type Post struct {
        Caption string
    }
    posts := make([]Post, len(result.Data))
    for i, t := range result.Data {
        posts[i] = Post{Caption: t.Text}
    }

    return containsTournament(posts, tournament), nil
}

// VerifyPostExists checks if a specific Instagram post still exists
func (c *Client) VerifyPostExists(postID string) (bool, error) {
	if postID == "" {
		return false, fmt.Errorf("empty post ID")
	}

	// Try to fetch the post
	postURL := fmt.Sprintf("%s/%s?fields=id&access_token=%s", 
		GraphAPIBaseURL, postID, url.QueryEscape(c.config.AccessToken))
	
	resp, err := c.httpClient.Get(postURL)
	if err != nil {
		return false, fmt.Errorf("failed to check post: %w", err)
	}
	defer resp.Body.Close()

	// If post exists, we get 200
	// If deleted, we get 404 or error
	if resp.StatusCode == http.StatusOK {
		return true, nil
	}
	
	// Post doesn't exist anymore
	return false, nil
}

// VerifyThreadExists checks if a specific Threads post still exists
func (c *Client) VerifyThreadExists(threadID string) (bool, error) {
	if threadID == "" {
		return false, fmt.Errorf("empty thread ID")
	}

	// Try to fetch the thread
	threadURL := fmt.Sprintf("%s/%s?fields=id&access_token=%s",
		ThreadsAPIBaseURL, threadID, url.QueryEscape(c.config.ThreadsAccessToken))
	
	resp, err := c.httpClient.Get(threadURL)
	if err != nil {
		return false, fmt.Errorf("failed to check thread: %w", err)
	}
	defer resp.Body.Close()

	// If thread exists, we get 200
	// If deleted, we get 404 or error
	if resp.StatusCode == http.StatusOK {
		return true, nil
	}
	
	// Thread doesn't exist anymore
	return false, nil
}

// SyncCacheWithAPI validates cache entries against Instagram/Threads APIs
func (c *Client) SyncCacheWithAPI() error {
	cache := instagram.GetPostedCache()
	
	checkInstagram := func(tournamentID int, postID string) (bool, error) {
		return c.VerifyPostExists(postID)
	}
	
	checkThreads := func(tournamentID int, threadID string) (bool, error) {
		return c.VerifyThreadExists(threadID)
	}
	
	return cache.ValidateWithAPI(checkInstagram, checkThreads)
}

// containsTournament checks if any post contains the tournament name or URL
func containsTournament[T any](posts []T, tournament igimage.TournamentImage) bool {
    // Extract caption from generic post type
    getCaptionFunc := func(post T) string {
        switch p := any(post).(type) {
        case struct {
            ID      string `json:"id"`
            Caption string `json:"caption"`
        }:
            return p.Caption
        case struct {
            Caption string
        }:
            return p.Caption
        default:
            return ""
        }
    }

    tournamentName := strings.ToLower(tournament.Name)
    tournamentURL := strings.ToLower(tournament.TournamentURL)

    for _, post := range posts {
        caption := strings.ToLower(getCaptionFunc(post))
        
        // Check if caption contains tournament name or URL
        if strings.Contains(caption, tournamentName) {
            log.Printf("✓ Found tournament by name in caption: %s", tournamentName)
            return true
        }
        
        if tournamentURL != "" && strings.Contains(caption, tournamentURL) {
            log.Printf("✓ Found tournament by URL in caption: %s", tournamentURL)
            return true
        }
        
        // Also check for tournament ID in URL format
        tournamentIDInURL := fmt.Sprintf("tournois-tt.fr/%d", tournament.TournamentID)
        if strings.Contains(caption, tournamentIDInURL) {
            log.Printf("✓ Found tournament by ID in caption: %d", tournament.TournamentID)
            return true
        }
    }

    return false
}


