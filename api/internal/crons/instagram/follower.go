package instagram

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"tournois-tt/api/internal/config"
	instabot "tournois-tt/api/pkg/instagram/bot"
)

const (
	ffttAccountsFile       = "./fftt-instagram-accounts.json"
	instagramBlacklistFile = "./instagram_blacklist.json"
)

// Blacklist represents a list of blacklisted Instagram usernames
type Blacklist struct {
	Usernames []string `json:"usernames"`
}

// FFTTAccount represents an Instagram account to follow
type FFTTAccount struct {
	Username   string    `json:"username"`
	Type       string    `json:"type"`
	Region     string    `json:"region,omitempty"`
	Followed   bool      `json:"followed"`
	FollowedAt time.Time `json:"followed_at,omitempty"`
	Notes      string    `json:"notes,omitempty"`
}


// AccountList manages a list of FFTT accounts
type AccountList struct {
	Accounts    []FFTTAccount `json:"accounts"`
	LastUpdated time.Time     `json:"last_updated"`
}

// RunFollowerBot runs the Instagram follower bot during daytime hours
func RunFollowerBot() {
	// Check if bot is enabled
	if !config.InstagramBotEnabled {
		log.Println("🤖 Instagram follower bot is disabled (INSTAGRAM_BOT_ENABLED not set or false)")
		return
	}

	location, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		log.Printf("⚠️  Failed to load Europe/Paris timezone: %v", err)
		return
	}

	now := time.Now().In(location)
	hour := now.Hour()

	// Only run during daytime (11 AM to 9 PM Paris time)
	if hour < 11 || hour >= 21 {
		log.Printf("⏰ Outside daytime hours (current: %d:00 Paris time). Follower bot skipping.", hour)
		return
	}

	// Weekend check - be more conservative on weekends
	if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
		log.Printf("📅 Weekend detected - taking it easy like a human would")
		// Only run at specific hours on weekends
		if hour != 13 && hour != 17 {
			log.Printf("⏰ Skipping non-prime weekend hours")
			return
		}
	}

	log.Println("🤖 Starting Instagram follower bot...")

	// Get credentials from environment
	username := os.Getenv("INSTAGRAM_BOT_USERNAME")
	password := os.Getenv("INSTAGRAM_BOT_PASSWORD")
	totpSecret := os.Getenv("INSTAGRAM_BOT_TOTP_SECRET")

	if username == "" || password == "" {
		log.Println("ERROR: INSTAGRAM_BOT_USERNAME and INSTAGRAM_BOT_PASSWORD must be set")
		return
	}

	if totpSecret == "" {
		log.Println("⚠️  WARNING: INSTAGRAM_BOT_TOTP_SECRET not set. If account has 2FA, login will fail.")
	}

	// Load blacklist
	blacklist, err := loadBlacklist()
	if err != nil {
		log.Printf("ERROR: Failed to load blacklist: %v", err)
		return
	}

	// Load accounts to follow
	accountsToFollow, err := loadUnfollowedAccounts(blacklist)
	if err != nil {
		log.Printf("ERROR: Failed to load accounts: %v", err)
		return
	}

	if len(accountsToFollow) == 0 {
		log.Println("✅ No accounts to follow. All done!")
		return
	}

	log.Printf("📋 Found %d accounts to follow", len(accountsToFollow))

	// Create bot instance
	bot := instabot.NewFollowerBot(username, password, totpSecret)

	// Follow accounts
	if err := bot.Follow(accountsToFollow); err != nil {
		log.Printf("ERROR: Follower bot failed: %v", err)
		return
	}

	// Update the accounts file to mark as followed
	if err := markAccountsAsFollowed(accountsToFollow); err != nil {
		log.Printf("WARNING: Failed to update accounts file: %v", err)
	}

	log.Println("✅ Follower bot session complete")
}

// loadUnfollowedAccounts loads accounts that haven't been followed yet
func loadUnfollowedAccounts(blacklist map[string]bool) ([]string, error) {
	data, err := os.ReadFile(ffttAccountsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read accounts file: %w", err)
	}

	var accountList AccountList
	if err := json.Unmarshal(data, &accountList); err != nil {
		return nil, fmt.Errorf("failed to parse accounts file: %w", err)
	}

	// Filter unfollowed accounts, prioritize by type
	var unfollowed []string

	// Helper to check blacklist and append
	addToFollowList := func(username string) {
		if blacklist[username] {
			log.Printf("🚫 Account %s is in the blacklist, skipping", username)
			return
		}
		unfollowed = append(unfollowed, username)
	}

	// Priority 1: Federation accounts
	for _, account := range accountList.Accounts {
		if !account.Followed && account.Type == "federation" {
			addToFollowList(account.Username)
		}
	}

	// Priority 2: Ligue accounts
	for _, account := range accountList.Accounts {
		if !account.Followed && account.Type == "ligue" {
			addToFollowList(account.Username)
		}
	}

	// Priority 3: Comité accounts
	for _, account := range accountList.Accounts {
		if !account.Followed && account.Type == "comite" {
			addToFollowList(account.Username)
		}
	}

	// Priority 4: Club accounts
	for _, account := range accountList.Accounts {
		if !account.Followed && account.Type == "club" {
			addToFollowList(account.Username)
		}
	}

	// Limit to a reasonable batch per run (will get more in next run)
	if len(unfollowed) > 10 {
		unfollowed = unfollowed[:10]
	}

	return unfollowed, nil
}

// markAccountsAsFollowed updates the accounts file to mark accounts as followed
func markAccountsAsFollowed(usernames []string) error {
	data, err := os.ReadFile(ffttAccountsFile)
	if err != nil {
		return fmt.Errorf("failed to read accounts file: %w", err)
	}

	var accountList AccountList
	if err := json.Unmarshal(data, &accountList); err != nil {
		return fmt.Errorf("failed to parse accounts file: %w", err)
	}

	// Create a map for quick lookup
	followedMap := make(map[string]bool)
	for _, username := range usernames {
		followedMap[username] = true
	}

	// Update accounts
	now := time.Now()
	for i := range accountList.Accounts {
		if followedMap[accountList.Accounts[i].Username] {
			accountList.Accounts[i].Followed = true
			accountList.Accounts[i].FollowedAt = now
		}
	}

	accountList.LastUpdated = now

	// Save updated list
	updatedData, err := json.MarshalIndent(accountList, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal accounts: %w", err)
	}

	if err := os.WriteFile(ffttAccountsFile, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write accounts file: %w", err)
	}

	log.Printf("✅ Updated accounts file - marked %d accounts as followed", len(usernames))
	return nil
}

// loadBlacklist loads the Instagram blacklist from a JSON file
func loadBlacklist() (map[string]bool, error) {
	blacklist := make(map[string]bool)

	data, err := os.ReadFile(instagramBlacklistFile)
	if err != nil {
		// If the file doesn't exist, it's not an error, just return an empty blacklist
		if os.IsNotExist(err) {
			log.Println("⚠️  Blacklist file not found, proceeding without a blacklist.")
			return blacklist, nil
		}
		return nil, fmt.Errorf("failed to read blacklist file: %w", err)
	}

	var blacklistData Blacklist
	if err := json.Unmarshal(data, &blacklistData); err != nil {
		return nil, fmt.Errorf("failed to parse blacklist file: %w", err)
	}

	for _, username := range blacklistData.Usernames {
		blacklist[username] = true
	}

	log.Printf("✅ Loaded %d usernames from the blacklist", len(blacklist))
	return blacklist, nil
}


// RunUnfollowerBot runs the Instagram unfollower bot during daytime hours
func RunUnfollowerBot() {
	// Check if bot is enabled
	if !config.InstagramBotEnabled {
		log.Println("🤖 Instagram unfollower bot is disabled (INSTAGRAM_BOT_ENABLED not set or false)")
		return
	}

	location, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		log.Printf("⚠️  Failed to load Europe/Paris timezone: %v", err)
		return
	}

	now := time.Now().In(location)
	hour := now.Hour()

	// Only run during daytime (11 AM to 9 PM Paris time)
	if hour < 11 || hour >= 21 {
		log.Printf("⏰ Outside daytime hours (current: %d:00 Paris time). Unfollower bot skipping.", hour)
		return
	}

	log.Println("🤖 Starting Instagram unfollower bot...")

	// Get credentials from environment
	username := os.Getenv("INSTAGRAM_BOT_USERNAME")
	password := os.Getenv("INSTAGRAM_BOT_PASSWORD")
	totpSecret := os.Getenv("INSTAGRAM_BOT_TOTP_SECRET")

	if username == "" || password == "" {
		log.Println("ERROR: INSTAGRAM_BOT_USERNAME and INSTAGRAM_BOT_PASSWORD must be set")
		return
	}

	if totpSecret == "" {
		log.Println("⚠️  WARNING: INSTAGRAM_BOT_TOTP_SECRET not set. If account has 2FA, login will fail.")
	}

	// Load blacklist
	blacklist, err := loadBlacklist()
	if err != nil {
		log.Printf("ERROR: Failed to load blacklist: %v", err)
		return
	}

	if len(blacklist) == 0 {
		log.Println("✅ No accounts in the blacklist. All done!")
		return
	}

	// Load all accounts
	accounts, err := loadAllAccounts()
	if err != nil {
		log.Printf("ERROR: Failed to load accounts: %v", err)
		return
	}

	// Find accounts to unfollow
	accountsToUnfollow := findAccountsToUnfollow(accounts, blacklist)
	if len(accountsToUnfollow) == 0 {
		log.Println("✅ No blacklisted accounts to unfollow. All done!")
		return
	}

	log.Printf("📋 Found %d blacklisted accounts to unfollow", len(accountsToUnfollow))

	// Create bot instance
	bot := instabot.NewFollowerBot(username, password, totpSecret)

	// Unfollow accounts
	if err := bot.Unfollow(accountsToUnfollow); err != nil {
		log.Printf("ERROR: Unfollower bot failed: %v", err)
		return
	}

	// Update the accounts file to mark as unfollowed
	if err := markAccountsAsUnfollowed(accountsToUnfollow); err != nil {
		log.Printf("WARNING: Failed to update accounts file: %v", err)
	}

	log.Println("✅ Unfollower bot session complete")
}

// loadAllAccounts loads all accounts from the accounts file
func loadAllAccounts() ([]FFTTAccount, error) {
	data, err := os.ReadFile(ffttAccountsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read accounts file: %w", err)
	}

	var accountList AccountList
	if err := json.Unmarshal(data, &accountList); err != nil {
		return nil, fmt.Errorf("failed to parse accounts file: %w", err)
	}

	return accountList.Accounts, nil
}

// findAccountsToUnfollow finds followed accounts that are in the blacklist
func findAccountsToUnfollow(accounts []FFTTAccount, blacklist map[string]bool) []string {
	var accountsToUnfollow []string
	for _, account := range accounts {
		if account.Followed && blacklist[account.Username] {
			accountsToUnfollow = append(accountsToUnfollow, account.Username)
		}
	}
	return accountsToUnfollow
}

// markAccountsAsUnfollowed updates the accounts file to mark accounts as unfollowed
func markAccountsAsUnfollowed(usernames []string) error {
	data, err := os.ReadFile(ffttAccountsFile)
	if err != nil {
		return fmt.Errorf("failed to read accounts file: %w", err)
	}

	var accountList AccountList
	if err := json.Unmarshal(data, &accountList); err != nil {
		return fmt.Errorf("failed to parse accounts file: %w", err)
	}

	// Create a map for quick lookup
	unfollowedMap := make(map[string]bool)
	for _, username := range usernames {
		unfollowedMap[username] = true
	}

	// Update accounts
	now := time.Now()
	for i := range accountList.Accounts {
		if unfollowedMap[accountList.Accounts[i].Username] {
			accountList.Accounts[i].Followed = false
			accountList.Accounts[i].FollowedAt = time.Time{} // Reset timestamp
		}
	}

	accountList.LastUpdated = now

	// Save updated list
	updatedData, err := json.MarshalIndent(accountList, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal accounts: %w", err)
	}

	if err := os.WriteFile(ffttAccountsFile, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write accounts file: %w", err)
	}

	log.Printf("✅ Updated accounts file - marked %d accounts as unfollowed", len(usernames))
	return nil
}
