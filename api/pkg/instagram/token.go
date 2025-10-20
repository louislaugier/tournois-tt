package instagram

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
	// Token storage file path
	tokenStoragePath = "./instagram-token.json"

	// Refresh token when less than this many days remain
	refreshThresholdDays = 7

	// Instagram token validity period (60 days)
	tokenValidityDays = 60
)

// TokenStorage represents the persisted token data
type TokenStorage struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
	LastRefresh time.Time `json:"last_refresh"`
	Version     int       `json:"version"` // For future migrations
}

// TokenRefreshResponse represents Instagram's refresh response
type TokenRefreshResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"` // Seconds until expiration
}

// LoadToken loads the token from persistent storage or falls back to env var
// This is the main function to get the current valid token
func LoadToken() (string, error) {
	// Try to load from persistent storage first
	storage, err := loadTokenStorage()
	if err == nil && storage.AccessToken != "" {
		// Check if token needs refresh
		if shouldRefreshToken(storage) {
			log.Println("Token needs refresh, attempting refresh...")
			if err := RefreshToken(); err != nil {
				log.Printf("Failed to refresh token: %v", err)
				// Still return the current token, it might still be valid
			} else {
				// Reload after refresh
				storage, err = loadTokenStorage()
				if err != nil {
					return "", fmt.Errorf("failed to reload token after refresh: %w", err)
				}
			}
		}
		return storage.AccessToken, nil
	}

	// Fall back to environment variable (initial setup)
	envToken := config.InstagramAccessToken
	if envToken == "" {
		return "", fmt.Errorf("no Instagram token found in storage or environment")
	}

	log.Println("Using token from environment variable (first time setup)")

	// Initialize storage with env token
	// Assume it's a new token with 60 days validity
	storage = &TokenStorage{
		AccessToken: envToken,
		ExpiresAt:   time.Now().Add(tokenValidityDays * 24 * time.Hour),
		LastRefresh: time.Now(),
		Version:     1,
	}

	if err := saveTokenStorage(storage); err != nil {
		log.Printf("Warning: Failed to save initial token to storage: %v", err)
	}

	return envToken, nil
}

// RefreshToken refreshes the Instagram access token and persists it
func RefreshToken() error {
	// Load current token
	storage, err := loadTokenStorage()
	if err != nil || storage.AccessToken == "" {
		// Fall back to env token
		envToken := config.InstagramAccessToken
		if envToken == "" {
			return fmt.Errorf("no token available to refresh")
		}
		storage = &TokenStorage{
			AccessToken: envToken,
			ExpiresAt:   time.Now().Add(tokenValidityDays * 24 * time.Hour),
			LastRefresh: time.Now(),
			Version:     1,
		}
	}

	// Check if token is already expired
	if time.Now().After(storage.ExpiresAt) {
		return fmt.Errorf("token has expired, manual regeneration required")
	}

	// Call Instagram refresh API
	refreshURL := fmt.Sprintf("https://graph.instagram.com/refresh_access_token?grant_type=ig_refresh_token&access_token=%s",
		storage.AccessToken)

	resp, err := http.Get(refreshURL)
	if err != nil {
		return fmt.Errorf("failed to call refresh API: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	var refreshResp TokenRefreshResponse
	if err := json.Unmarshal(body, &refreshResp); err != nil {
		return fmt.Errorf("failed to parse refresh response: %w", err)
	}

	// Update storage with new token
	storage.AccessToken = refreshResp.AccessToken
	storage.ExpiresAt = time.Now().Add(time.Duration(refreshResp.ExpiresIn) * time.Second)
	storage.LastRefresh = time.Now()

	// Persist updated token
	if err := saveTokenStorage(storage); err != nil {
		return fmt.Errorf("failed to save refreshed token: %w", err)
	}

	// Update global config variable for immediate use
	config.InstagramAccessToken = refreshResp.AccessToken

	log.Printf("Successfully refreshed Instagram token. New expiration: %s", storage.ExpiresAt.Format(time.RFC3339))

	return nil
}

// ForceRefreshToken forces a token refresh regardless of expiration
func ForceRefreshToken(currentToken string) error {
	refreshURL := fmt.Sprintf("https://graph.instagram.com/refresh_access_token?grant_type=ig_refresh_token&access_token=%s",
		currentToken)

	resp, err := http.Get(refreshURL)
	if err != nil {
		return fmt.Errorf("failed to call refresh API: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	var refreshResp TokenRefreshResponse
	if err := json.Unmarshal(body, &refreshResp); err != nil {
		return fmt.Errorf("failed to parse refresh response: %w", err)
	}

	storage := &TokenStorage{
		AccessToken: refreshResp.AccessToken,
		ExpiresAt:   time.Now().Add(time.Duration(refreshResp.ExpiresIn) * time.Second),
		LastRefresh: time.Now(),
		Version:     1,
	}

	if err := saveTokenStorage(storage); err != nil {
		return fmt.Errorf("failed to save refreshed token: %w", err)
	}

	config.InstagramAccessToken = refreshResp.AccessToken

	log.Printf("Force refresh successful. New expiration: %s", storage.ExpiresAt.Format(time.RFC3339))

	return nil
}

// shouldRefreshToken checks if the token should be refreshed
func shouldRefreshToken(storage *TokenStorage) bool {
	daysUntilExpiry := time.Until(storage.ExpiresAt).Hours() / 24
	return daysUntilExpiry < refreshThresholdDays
}

// GetTokenInfo returns information about the current token
func GetTokenInfo() (expiresAt time.Time, daysRemaining float64, err error) {
	storage, err := loadTokenStorage()
	if err != nil {
		return time.Time{}, 0, err
	}

	daysRemaining = time.Until(storage.ExpiresAt).Hours() / 24
	return storage.ExpiresAt, daysRemaining, nil
}

// loadTokenStorage loads token data from persistent storage
func loadTokenStorage() (*TokenStorage, error) {
	data, err := os.ReadFile(tokenStoragePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read token storage: %w", err)
	}

	var storage TokenStorage
	if err := json.Unmarshal(data, &storage); err != nil {
		return nil, fmt.Errorf("failed to parse token storage: %w", err)
	}

	return &storage, nil
}

// saveTokenStorage saves token data to persistent storage
func saveTokenStorage(storage *TokenStorage) error {
	// Ensure directory exists
	dir := filepath.Dir(tokenStoragePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}

	data, err := json.MarshalIndent(storage, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token storage: %w", err)
	}

	if err := os.WriteFile(tokenStoragePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write token storage: %w", err)
	}

	return nil
}
