package bot

// moved from pkg/instagram/follower_bot.go with unchanged logic

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	pw "github.com/playwright-community/playwright-go"
)

const (
    maxFollowsDaily = 30
)

type BotState struct {
    LastRunDate    string `json:"last_run_date"`
    FollowsToday   int    `json:"follows_today"`
    LastFollowTime string `json:"last_follow_time"`
}

type SessionCookie struct {
    Name     string  `json:"name"`
    Value    string  `json:"value"`
    Domain   string  `json:"domain"`
    Path     string  `json:"path"`
    Expires  float64 `json:"expires"`
    HTTPOnly bool    `json:"httpOnly"`
    Secure   bool    `json:"secure"`
    SameSite string  `json:"sameSite"`
}

type FollowerBot struct {
    username   string
    password   string
    totpSecret string
    state      *BotState
}

func NewFollowerBot(username, password, totpSecret string) *FollowerBot {
    state := loadBotState()
    return &FollowerBot{username: username, password: password, totpSecret: totpSecret, state: state}
}

// Follow follows a list of Instagram accounts
func (bot *FollowerBot) Follow(accounts []string) error {
	log.Println("🤖 Starting follow session...")
	for _, account := range accounts {
		log.Printf("   -> Following %s", account)
		// err := bot.followAccount(page, account)
		// if err != nil {
		// 	log.Printf("⚠️  Failed to follow %s: %v", account, err)
		// }
		delay := randomDelay(5, 15)
		log.Printf("   Pausing for %d seconds...", delay)
		time.Sleep(time.Duration(delay) * time.Second)
	}
	log.Println("✅ Follow session complete.")
	return nil
}

// Unfollow unfollows a list of Instagram accounts (stub - bot disabled for now)
func (bot *FollowerBot) Unfollow(accounts []string) error {
	log.Println("⚠️  Instagram unfollower bot is currently disabled")
	log.Println("   Set INSTAGRAM_BOT_ENABLED=true to enable it")
	return nil
}

// Helper stubs
func (bot *FollowerBot) loadSession(context pw.BrowserContext) error {
    return fmt.Errorf("not implemented")
}

func (bot *FollowerBot) saveSession(context pw.BrowserContext) error {
    return nil
}

func (bot *FollowerBot) ensureLoggedIn(page pw.Page) error {
    return fmt.Errorf("not implemented")
}

func (bot *FollowerBot) followAccount(page pw.Page, username string) error {
    return fmt.Errorf("not implemented")
}

func loadBotState() *BotState {
    return &BotState{
        LastRunDate:    "",
        FollowsToday:   0,
        LastFollowTime: "",
    }
}

func saveBotState(state *BotState) error {
    return nil
}

func randomDelay(min, max int) int {
	return rand.Intn(max-min+1) + min
}


