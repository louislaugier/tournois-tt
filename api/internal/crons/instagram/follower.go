package instagram

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"tournois-tt/api/pkg/instagram/bot"
)

const (
	ffttAccountsFile       = "./fftt-instagram-accounts.json"
	instagramBlacklistFile = "./instagram_blacklist.json"
)

type SourceAccounts struct {
	Accounts []string `json:"source_accounts"`
}

type Blacklist struct {
	Usernames []string `json:"usernames"`
}

func RunFollowerBotOnStartup() {
	// Wait a bit for the app to be ready
	time.Sleep(10 * time.Second)

	log.Println("🤖 Starting Instagram follower bot on startup...")

	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	for {
		location, err := time.LoadLocation("Europe/Paris")
		if err != nil {
			log.Printf("⚠️  Failed to load Europe/Paris timezone: %v", err)
			time.Sleep(5 * time.Minute) // Wait before retrying
			continue
		}

		now := time.Now().In(location)
		hour := now.Hour()

		// Use hardcoded time window from bot package
		log.Printf("🕐 Current Paris time: %s (hour: %d, window: %d-%d)", now.Format("15:04"), hour, bot.MinHour, bot.MaxHour)

		// Only run during configured hours (11 AM to 9 PM Paris time)
		if hour >= bot.MinHour && hour < bot.MaxHour {
			if err := RunFollowerBot(); err != nil {
				log.Printf("❌ ERROR: Follower bot run failed: %v", err)
			}
		} else {
			log.Printf("⏰ Outside of time window (current: %d, window: %d-%d). Skipping this run.", hour, bot.MinHour, bot.MaxHour)
		}

		// Sleep for a random duration before the next run (30 to 60 minutes)
		sleepDuration := time.Duration(rand.Intn(30)+30) * time.Minute
		log.Printf("⏰ Follower bot will run again in %v (at approx. %s)", sleepDuration, time.Now().Add(sleepDuration).In(location).Format("15:04"))
		time.Sleep(sleepDuration)
	}
}

func RunFollowerBot() error {
	log.Println("🤖 Starting new follower bot session...")

	username := os.Getenv("INSTAGRAM_BOT_USERNAME")
	password := os.Getenv("INSTAGRAM_BOT_PASSWORD")
	totpSecret := os.Getenv("INSTAGRAM_BOT_TOTP_SECRET")
	headless := getEnvAsBool("INSTAGRAM_BOT_HEADLESS", true)

	if username == "" || password == "" {
		return fmt.Errorf("INSTAGRAM_BOT_USERNAME and INSTAGRAM_BOT_PASSWORD must be set")
	}

	// Build config - all randomization is now done internally in the bot
	config := bot.BotConfig{
		DataDir: "./tmp/bot_data",
	}

	followerBot, err := bot.NewFollowerBot(username, password, totpSecret, config, headless)
	if err != nil {
		return fmt.Errorf("failed to create follower bot: %w", err)
	}
	defer followerBot.Close()

	// Load blacklist
	blacklist, err := loadBlacklist()
	if err != nil {
		return fmt.Errorf("failed to load blacklist: %w", err)
	}
	log.Printf("📋 Loaded %d users in blacklist.", len(blacklist))

	// Filter blacklist to only users we're currently following
	myFollowing, err := followerBot.GetFollowing()
	if err != nil {
		return fmt.Errorf("failed to get my following list: %w", err)
	}
	myFollowingMap := listToMap(myFollowing)
	log.Printf("📋 Currently following %d users.", len(myFollowing))

	// Check if unfollowing is enabled
	unfollowEnabled := getEnvAsBool("INSTAGRAM_BOT_UNFOLLOW_ENABLED", true)
	
	if unfollowEnabled {
		// Get users from blacklist that we are actually following
		var accountsToUnfollow []string
		for username := range blacklist {
			if myFollowingMap[username] {
				accountsToUnfollow = append(accountsToUnfollow, username)
			}
		}

		// Unfollow blacklisted users first
		if len(accountsToUnfollow) > 0 {
			log.Printf("🚫 Found %d blacklisted users to unfollow.", len(accountsToUnfollow))
			if err := followerBot.UnfollowUsers(accountsToUnfollow); err != nil {
				log.Printf("⚠️  Unfollow session failed: %v", err)
			}
		} else {
			log.Println("✅ No blacklisted users to unfollow.")
		}
	} else {
		log.Println("⚠️  Unfollowing is disabled via INSTAGRAM_BOT_UNFOLLOW_ENABLED.")
	}

	// Load source accounts
	sourceAccounts, err := loadSourceAccounts()
	if err != nil {
		return fmt.Errorf("failed to load source accounts: %w", err)
	}
	log.Printf("📋 Loaded %d source accounts.", len(sourceAccounts))

	// Randomly select one source account for this iteration
	if len(sourceAccounts) == 0 {
		return fmt.Errorf("no source accounts available")
	}
	randomIndex := rand.Intn(len(sourceAccounts))
	selectedSource := sourceAccounts[randomIndex]
	log.Printf("🎲 Randomly selected source account: %s", selectedSource)

	// Get followers from the selected source account
	var accountsToFollow []string
	followers, err := followerBot.GetFollowers(selectedSource)
	if err != nil {
		return fmt.Errorf("failed to get followers for %s: %w", selectedSource, err)
	}

	for _, follower := range followers {
		// Don't follow if already following or in blacklist
		if !myFollowingMap[follower] && !blacklist[follower] {
			accountsToFollow = append(accountsToFollow, follower)
		}
	}

	if len(accountsToFollow) == 0 {
		log.Println("✅ No new accounts to follow in this session.")
		return nil
	}

	log.Printf("📋 Found %d new accounts to follow.", len(accountsToFollow))

	// Follow new accounts
	if err := followerBot.FollowUsers(accountsToFollow); err != nil {
		return fmt.Errorf("follow session failed: %w", err)
	}

	log.Println("✅ Follower bot session complete.")
	return nil
}

func loadSourceAccounts() ([]string, error) {
	// Try multiple possible paths for the JSON file
	possiblePaths := []string{
		ffttAccountsFile,
		"/go/src/tournois-tt/api/fftt-instagram-accounts.json",
		"./api/fftt-instagram-accounts.json",
	}
	
	var data []byte
	var err error
	var foundPath string
	for _, path := range possiblePaths {
		data, err = os.ReadFile(path)
		if err == nil {
			foundPath = path
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read source accounts file: %w", err)
	}
	log.Printf("✅ Loaded source accounts from: %s", foundPath)
	var sources SourceAccounts
	if err := json.Unmarshal(data, &sources); err != nil {
		return nil, fmt.Errorf("failed to parse source accounts file: %w", err)
	}
	return sources.Accounts, nil
}

func loadBlacklist() (map[string]bool, error) {
	blacklist := make(map[string]bool)
	
	// Try multiple possible paths for the JSON file
	possiblePaths := []string{
		instagramBlacklistFile,
		"/go/src/tournois-tt/api/instagram_blacklist.json",
		"./api/instagram_blacklist.json",
	}
	
	var data []byte
	var err error
	var foundPath string
	for _, path := range possiblePaths {
		data, err = os.ReadFile(path)
		if err == nil {
			foundPath = path
			break
		}
	}
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("ℹ️  No blacklist file found, starting with empty blacklist.")
			return blacklist, nil // No blacklist is not an error
		}
		return nil, fmt.Errorf("failed to read blacklist file: %w", err)
	}
	log.Printf("✅ Loaded blacklist from: %s", foundPath)
	var bl Blacklist
	if err := json.Unmarshal(data, &bl); err != nil {
		return nil, fmt.Errorf("failed to parse blacklist file: %w", err)
	}
	for _, username := range bl.Usernames {
		blacklist[username] = true
	}
	return blacklist, nil
}

func listToMap(list []string) map[string]bool {
	m := make(map[string]bool)
	for _, item := range list {
		m[item] = true
	}
	return m
}

func getEnvAsInt(name string, defaultValue int) int {
	valueStr := os.Getenv(name)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("⚠️  Failed to parse %s as int: %v. Using default: %d", name, err, defaultValue)
		return defaultValue
	}
	return value
}

func getEnvAsBool(name string, defaultValue bool) bool {
	valueStr := os.Getenv(name)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		log.Printf("⚠️  Failed to parse %s as bool: %v. Using default: %t", name, err, defaultValue)
		return defaultValue
	}
	return value
}
