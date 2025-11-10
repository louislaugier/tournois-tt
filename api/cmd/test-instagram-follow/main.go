package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"strconv"

	"tournois-tt/api/pkg/instagram/bot"
)

type SourceAccounts struct {
	Accounts []string `json:"source_accounts"`
}

type BlacklistAccounts struct {
	Usernames []string `json:"usernames"`
}

func main() {
	log.Println("🚀 Instagram Follow/Unfollow Bot Test")
	log.Println("======================================")

	instaUsername := os.Getenv("INSTAGRAM_BOT_USERNAME")
	instaPassword := os.Getenv("INSTAGRAM_BOT_PASSWORD")
	instaTotpSecret := os.Getenv("INSTAGRAM_BOT_TOTP_SECRET")
	headless := getEnvAsBool("INSTAGRAM_BOT_HEADLESS", true)

	if instaUsername == "" || instaPassword == "" {
		log.Fatal("❌ Missing required environment variables: INSTAGRAM_BOT_USERNAME, INSTAGRAM_BOT_PASSWORD")
	}

	log.Printf("✅ Bot username: %s", instaUsername)
	log.Printf("✅ Headless mode: %t", headless)

	config := bot.BotConfig{
		DataDir: "tmp/bot_data",
	}

	log.Println("\n📋 Bot Configuration:")
	log.Printf("  - Time window: %d:00 - %d:00", bot.MinHour, bot.MaxHour)
	log.Printf("  - Max follows/day: 25-35 (randomized)")
	log.Printf("  - Max unfollows/day: 25-35 (randomized)")
	log.Printf("  - Pause between follows: 5-15 seconds")
	log.Printf("  - Pause between batches: 1-3 minutes")
	log.Printf("  - Follow batch size: 3-7")
	log.Printf("  - Unfollow batch size: 3-7")
	log.Printf("  - Data directory: %s\n", config.DataDir)

	followerBot, err := bot.NewFollowerBot(instaUsername, instaPassword, instaTotpSecret, config, headless)
	if err != nil {
		log.Fatalf("❌ Failed to create follower bot: %v", err)
	}
	defer followerBot.Close()

	log.Println("✅ Bot initialized successfully")

	// Get my following list first
	log.Println("\n🔍 Getting your following list...")
	following, err := followerBot.GetFollowing()
	if err != nil {
		log.Fatalf("❌ Failed to get following list: %v", err)
	}
	log.Printf("✅ Currently following %d users", len(following))

	followingMap := make(map[string]bool)
	for _, user := range following {
		followingMap[user] = true
	}

	// Load and process blacklist
	log.Println("\n🚫 Processing blacklist...")
	blacklistAccounts, err := getAccountsToUnfollow()
	if err != nil {
		log.Printf("⚠️  Failed to get blacklist accounts: %v", err)
	} else if len(blacklistAccounts.Usernames) > 0 {
		log.Printf("📋 Found %d accounts in blacklist", len(blacklistAccounts.Usernames))

		// Filter to only users we're actually following
		var accountsToUnfollow []string
		for _, username := range blacklistAccounts.Usernames {
			if followingMap[username] {
				accountsToUnfollow = append(accountsToUnfollow, username)
			}
		}

		if len(accountsToUnfollow) > 0 {
			log.Printf("🚫 Will unfollow %d blacklisted users", len(accountsToUnfollow))
			if err := followerBot.UnfollowUsers(accountsToUnfollow); err != nil {
				log.Printf("⚠️  Failed to unfollow accounts: %v", err)
			}
		} else {
			log.Println("✅ No blacklisted users are currently followed")
		}
	} else {
		log.Println("ℹ️  No accounts in blacklist")
	}

	// Load source accounts and gather followers
	log.Println("\n📋 Loading source accounts...")
	sourceAccounts, err := getSourceAccounts()
	if err != nil {
		log.Fatalf("❌ Failed to get source accounts: %v", err)
	}
	log.Printf("✅ Found %d source accounts: %v", len(sourceAccounts.Accounts), sourceAccounts.Accounts)

	allFollowers := []string{}
	for _, account := range sourceAccounts.Accounts {
		log.Printf("\n🔍 Getting followers of %s...", account)
		followers, err := followerBot.GetFollowers(account)
		if err != nil {
			log.Printf("⚠️  Failed to get followers of %s: %v", account, err)
			continue
		}
		log.Printf("✅ Found %d followers", len(followers))
		allFollowers = append(allFollowers, followers...)
	}

	// Create blacklist map for quick lookup
	blacklistMap := make(map[string]bool)
	if blacklistAccounts != nil {
		for _, username := range blacklistAccounts.Usernames {
			blacklistMap[username] = true
		}
	}

	// Find accounts to follow (not following and not in blacklist)
	accountsToFollow := []string{}
	followersSeen := make(map[string]bool)

	for _, follower := range allFollowers {
		// Skip duplicates
		if followersSeen[follower] {
			continue
		}
		followersSeen[follower] = true

		// Add if not following and not in blacklist
		if !followingMap[follower] && !blacklistMap[follower] {
			accountsToFollow = append(accountsToFollow, follower)
		}
	}

	if len(accountsToFollow) == 0 {
		log.Println("\n✅ No new accounts to follow")
		log.Println("🎉 Test completed successfully!")
		return
	}

	log.Printf("\n📋 Found %d unique new accounts to follow", len(accountsToFollow))

	if err := followerBot.FollowUsers(accountsToFollow); err != nil {
		log.Fatalf("❌ Failed to follow accounts: %v", err)
	}

	log.Println("\n🎉 Test completed successfully!")
}

func getSourceAccounts() (*SourceAccounts, error) {
	jsonFile, err := os.Open("fftt-instagram-accounts.json")
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	var sourceAccounts SourceAccounts
	if err := json.Unmarshal(byteValue, &sourceAccounts); err != nil {
		return nil, err
	}

	return &sourceAccounts, nil
}

func getAccountsToUnfollow() (*BlacklistAccounts, error) {
	jsonFile, err := os.Open("instagram_blacklist.json")
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	var blacklistAccounts BlacklistAccounts
	if err := json.Unmarshal(byteValue, &blacklistAccounts); err != nil {
		return nil, err
	}

	return &blacklistAccounts, nil
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
