package api

import (
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "time"

    "tournois-tt/api/internal/config"
)

const (
    tokenStoragePath        = "./instagram-token.json"
    threadsTokenStoragePath = "./threads-token.json"
    refreshThresholdDays    = 7
    tokenValidityDays       = 60
)

type TokenStorage struct {
    AccessToken string    `json:"access_token"`
    ExpiresAt   time.Time `json:"expires_at"`
    LastRefresh time.Time `json:"last_refresh"`
    Version     int       `json:"version"`
}

type TokenRefreshResponse struct {
    AccessToken string `json:"access_token"`
    TokenType   string `json:"token_type"`
    ExpiresIn   int    `json:"expires_in"`
}

func LoadToken() (string, error) {
    storage, err := loadTokenStorage()
    if err == nil && storage.AccessToken != "" {
        if shouldRefreshToken(storage) {
            log.Println("Token needs refresh, attempting refresh...")
            if err := RefreshToken(); err != nil { log.Printf("Failed to refresh token: %v", err) } else {
                storage, err = loadTokenStorage(); if err != nil { return "", fmt.Errorf("failed to reload token after refresh: %w", err) }
            }
        }
        return storage.AccessToken, nil
    }
    envToken := config.InstagramAccessToken
    if envToken == "" { return "", fmt.Errorf("no Instagram token found in storage or environment") }
    log.Println("Using token from environment variable (first time setup)")
    storage = &TokenStorage{AccessToken: envToken, ExpiresAt: time.Now().Add(tokenValidityDays*24*time.Hour), LastRefresh: time.Now(), Version: 1}
    _ = saveTokenStorage(storage)
    return envToken, nil
}

func RefreshToken() error {
    storage, err := loadTokenStorage()
    if err != nil || storage.AccessToken == "" {
        envToken := config.InstagramAccessToken
        if envToken == "" { return fmt.Errorf("no token available to refresh") }
        storage = &TokenStorage{AccessToken: envToken, ExpiresAt: time.Now().Add(tokenValidityDays*24*time.Hour), LastRefresh: time.Now(), Version: 1}
    }
    if time.Now().After(storage.ExpiresAt) { return fmt.Errorf("token has expired, manual regeneration required") }
    refreshURL := fmt.Sprintf("https://graph.instagram.com/refresh_access_token?grant_type=ig_refresh_token&access_token=%s", storage.AccessToken)
    resp, err := http.Get(refreshURL); if err != nil { return fmt.Errorf("failed to call refresh API: %w", err) }
    defer resp.Body.Close()
    body, _ := io.ReadAll(resp.Body)
    if resp.StatusCode != http.StatusOK { return fmt.Errorf("refresh failed with status %d: %s", resp.StatusCode, string(body)) }
    var refreshResp TokenRefreshResponse
    if err := json.Unmarshal(body, &refreshResp); err != nil { return fmt.Errorf("failed to parse refresh response: %w", err) }
    storage.AccessToken = refreshResp.AccessToken
    storage.ExpiresAt = time.Now().Add(time.Duration(refreshResp.ExpiresIn) * time.Second)
    storage.LastRefresh = time.Now()
    if err := saveTokenStorage(storage); err != nil { return fmt.Errorf("failed to save refreshed token: %w", err) }
    config.InstagramAccessToken = refreshResp.AccessToken
    log.Printf("Successfully refreshed Instagram token. New expiration: %s", storage.ExpiresAt.Format(time.RFC3339))
    return nil
}

func ForceRefreshToken(currentToken string) error {
    refreshURL := fmt.Sprintf("https://graph.instagram.com/refresh_access_token?grant_type=ig_refresh_token&access_token=%s", currentToken)
    resp, err := http.Get(refreshURL); if err != nil { return fmt.Errorf("failed to call refresh API: %w", err) }
    defer resp.Body.Close()
    body, _ := io.ReadAll(resp.Body)
    if resp.StatusCode != http.StatusOK { return fmt.Errorf("refresh failed with status %d: %s", resp.StatusCode, string(body)) }
    var refreshResp TokenRefreshResponse
    if err := json.Unmarshal(body, &refreshResp); err != nil { return fmt.Errorf("failed to parse refresh response: %w", err) }
    storage := &TokenStorage{AccessToken: refreshResp.AccessToken, ExpiresAt: time.Now().Add(time.Duration(refreshResp.ExpiresIn) * time.Second), LastRefresh: time.Now(), Version: 1}
    if err := saveTokenStorage(storage); err != nil { return fmt.Errorf("failed to save refreshed token: %w", err) }
    config.InstagramAccessToken = refreshResp.AccessToken
    log.Printf("Force refresh successful. New expiration: %s", storage.ExpiresAt.Format(time.RFC3339))
    return nil
}

func shouldRefreshToken(storage *TokenStorage) bool {
    daysUntilExpiry := time.Until(storage.ExpiresAt).Hours() / 24
    return daysUntilExpiry < refreshThresholdDays
}

func GetTokenInfo() (time.Time, float64, error) {
    storage, err := loadTokenStorage(); if err != nil { return time.Time{}, 0, err }
    daysRemaining := time.Until(storage.ExpiresAt).Hours() / 24
    return storage.ExpiresAt, daysRemaining, nil
}

func loadTokenStorage() (*TokenStorage, error) {
    data, err := os.ReadFile(tokenStoragePath)
    if err != nil { return nil, fmt.Errorf("failed to read token storage: %w", err) }
    var storage TokenStorage
    if err := json.Unmarshal(data, &storage); err != nil { return nil, fmt.Errorf("failed to parse token storage: %w", err) }
    return &storage, nil
}

func saveTokenStorage(storage *TokenStorage) error {
    dir := filepath.Dir(tokenStoragePath)
    if err := os.MkdirAll(dir, 0755); err != nil { return fmt.Errorf("failed to create storage directory: %w", err) }
    data, err := json.MarshalIndent(storage, "", "  ")
    if err != nil { return fmt.Errorf("failed to marshal token storage: %w", err) }
    if err := os.WriteFile(tokenStoragePath, data, 0600); err != nil { return fmt.Errorf("failed to write token storage: %w", err) }
    return nil
}

func LoadThreadsToken() (string, error) {
    storage, err := loadThreadsTokenStorage()
    if err == nil && storage.AccessToken != "" {
        if shouldRefreshToken(storage) {
            log.Println("Threads token needs refresh, attempting refresh...")
            if err := RefreshThreadsToken(); err != nil { log.Printf("Failed to refresh Threads token: %v", err) } else {
                storage, err = loadThreadsTokenStorage(); if err != nil { return "", fmt.Errorf("failed to reload Threads token after refresh: %w", err) }
            }
        }
        return storage.AccessToken, nil
    }
    envToken := config.ThreadsAccessToken
    if envToken == "" { return "", fmt.Errorf("no Threads token found in storage or environment") }
    log.Println("Using Threads token from environment variable (first time setup)")
    storage = &TokenStorage{AccessToken: envToken, ExpiresAt: time.Now().Add(tokenValidityDays*24*time.Hour), LastRefresh: time.Now(), Version: 1}
    _ = saveThreadsTokenStorage(storage)
    return envToken, nil
}

func RefreshThreadsToken() error {
    storage, err := loadThreadsTokenStorage()
    if err != nil || storage.AccessToken == "" {
        envToken := config.ThreadsAccessToken
        if envToken == "" { return fmt.Errorf("no Threads token available to refresh") }
        storage = &TokenStorage{AccessToken: envToken, ExpiresAt: time.Now().Add(tokenValidityDays*24*time.Hour), LastRefresh: time.Now(), Version: 1}
    }
    if time.Now().After(storage.ExpiresAt) { return fmt.Errorf("Threads token has expired, manual regeneration required") }
    refreshURL := fmt.Sprintf("https://graph.threads.net/refresh_access_token?grant_type=th_refresh_token&access_token=%s", storage.AccessToken)
    resp, err := http.Get(refreshURL); if err != nil { return fmt.Errorf("failed to call Threads refresh API: %w", err) }
    defer resp.Body.Close()
    body, _ := io.ReadAll(resp.Body)
    if resp.StatusCode != http.StatusOK { return fmt.Errorf("Threads refresh failed with status %d: %s", resp.StatusCode, string(body)) }
    var refreshResp TokenRefreshResponse
    if err := json.Unmarshal(body, &refreshResp); err != nil { return fmt.Errorf("failed to parse Threads refresh response: %w", err) }
    storage.AccessToken = refreshResp.AccessToken
    storage.ExpiresAt = time.Now().Add(time.Duration(refreshResp.ExpiresIn) * time.Second)
    storage.LastRefresh = time.Now()
    if err := saveThreadsTokenStorage(storage); err != nil { return fmt.Errorf("failed to save refreshed Threads token: %w", err) }
    config.ThreadsAccessToken = refreshResp.AccessToken
    log.Printf("Successfully refreshed Threads token. New expiration: %s", storage.ExpiresAt.Format(time.RFC3339))
    return nil
}

func loadThreadsTokenStorage() (*TokenStorage, error) {
    data, err := os.ReadFile(threadsTokenStoragePath)
    if err != nil { return nil, fmt.Errorf("failed to read Threads token storage: %w", err) }
    var storage TokenStorage
    if err := json.Unmarshal(data, &storage); err != nil { return nil, fmt.Errorf("failed to parse Threads token storage: %w", err) }
    return &storage, nil
}

func saveThreadsTokenStorage(storage *TokenStorage) error {
    dir := filepath.Dir(threadsTokenStoragePath)
    if err := os.MkdirAll(dir, 0755); err != nil { return fmt.Errorf("failed to create Threads storage directory: %w", err) }
    data, err := json.MarshalIndent(storage, "", "  ")
    if err != nil { return fmt.Errorf("failed to marshal Threads token storage: %w", err) }
    if err := os.WriteFile(threadsTokenStoragePath, data, 0600); err != nil { return fmt.Errorf("failed to write Threads token storage: %w", err) }
    return nil
}




