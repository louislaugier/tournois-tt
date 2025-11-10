package bot

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/pquerna/otp/totp"
)

const (
	MinHour = 11
	MaxHour = 21
)

type BotConfig struct {
	DataDir string // Directory to store session and state files
}

type BotState struct {
	LastRunDate    string `json:"last_run_date"`
	FollowsToday   int    `json:"follows_today"`
	UnfollowsToday int    `json:"unfollows_today"`
}

type FollowerBot struct {
	username   string
	password   string
	totpSecret string
	config     BotConfig
	state      *BotState
	pw         *playwright.Playwright
	browser    playwright.Browser
	context    playwright.BrowserContext
	page       playwright.Page
}

func NewFollowerBot(username, password, totpSecret string, config BotConfig, headless bool) (*FollowerBot, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("could not start playwright: %w", err)
	}

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(headless),
		SlowMo:   playwright.Float(1000),
		Args: []string{
			"--disable-blink-features=AutomationControlled",
			"--disable-dev-shm-usage",
			"--no-sandbox",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("could not launch browser: %w", err)
	}

	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: playwright.String("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	})
	if err != nil {
		return nil, fmt.Errorf("could not create context: %w", err)
	}

	page, err := context.NewPage()
	if err != nil {
		return nil, fmt.Errorf("could not create page: %w", err)
	}

	if config.DataDir == "" {
		config.DataDir = "."
	}
	if err := os.MkdirAll(config.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	state := loadBotState(config.DataDir)
	bot := &FollowerBot{
		username:   username,
		password:   password,
		totpSecret: totpSecret,
		config:     config,
		state:      state,
		pw:         pw,
		browser:    browser,
		context:    context,
		page:       page,
	}

	if err := bot.ensureLoggedIn(); err != nil {
		screenshotPath := filepath.Join(bot.config.DataDir, "login_error.png")
		bot.page.Screenshot(playwright.PageScreenshotOptions{Path: playwright.String(screenshotPath)})
		html, _ := bot.page.Content()
		log.Printf("HTML content at time of error: %s", html)
		return nil, fmt.Errorf("login failed: %w. Screenshot saved to %s", err, screenshotPath)
	}

	return bot, nil
}

func (bot *FollowerBot) isWithinTimeWindow() bool {
	location, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		log.Printf("⚠️ Failed to load Europe/Paris timezone: %v", err)
		return false
	}
	
	now := time.Now().In(location)
	currentHour := now.Hour()
	
	log.Printf("🕐 Current Paris time: %s (hour: %d, window: %d-%d)", now.Format("15:04"), currentHour, MinHour, MaxHour)
	
	return currentHour >= MinHour && currentHour < MaxHour
}

func (bot *FollowerBot) Close() {
	bot.browser.Close()
	bot.pw.Stop()
	if err := saveBotState(bot.state, bot.config.DataDir); err != nil {
		log.Printf("⚠️ Failed to save bot state: %v", err)
	}
}

func (bot *FollowerBot) ensureLoggedIn() error {
	if err := bot.loadSession(); err == nil {
		log.Println("✅ Session loaded from file.")
		if _, err := bot.page.Goto("https://www.instagram.com/"); err != nil {
			return fmt.Errorf("failed to navigate to instagram: %w", err)
		}
		// More robust check for login status
		if _, err := bot.page.WaitForSelector("a[aria-label='Home']", playwright.PageWaitForSelectorOptions{Timeout: playwright.Float(5000)}); err == nil {
			log.Println("✅ Session is valid.")
			return nil
		}
		log.Println("⚠️ Session is invalid or expired. Logging in again.")
	} else {
		log.Printf("⚠️ Failed to load session: %v", err)
	}

	log.Println("🤖 Logging in to Instagram...")
	if _, err := bot.page.Goto("https://www.instagram.com/accounts/login/"); err != nil {
		return fmt.Errorf("failed to navigate to login page: %w", err)
	}

	// Handle cookie consent dialog
	cookieButton, err := bot.page.QuerySelector("button:has-text('Allow all cookies')")
	if err != nil {
		log.Printf("⚠️ Error checking for cookie button: %v", err)
	} else if cookieButton != nil {
		log.Println("🍪 Cookie consent dialog found. Clicking 'Allow all cookies'.")
		if err := cookieButton.Click(); err != nil {
			log.Println("⚠️ Could not click 'Allow all cookies' button, but continuing.")
		}
		// Wait for the dialog to disappear
		time.Sleep(time.Duration(randomDelay(2, 4)) * time.Second)
	}

	if _, err := bot.page.WaitForSelector("input[name='username']", playwright.PageWaitForSelectorOptions{Timeout: playwright.Float(60000)}); err != nil {
		return fmt.Errorf("login form not found: %w", err)
	}

	time.Sleep(time.Duration(randomDelay(1, 3)) * time.Second)
	if err := bot.page.Fill("input[name='username']", bot.username); err != nil {
		return fmt.Errorf("failed to fill username: %w", err)
	}
	time.Sleep(time.Duration(randomDelay(1, 3)) * time.Second)
	if err := bot.page.Fill("input[name='password']", bot.password); err != nil {
		return fmt.Errorf("failed to fill password: %w", err)
	}
	time.Sleep(time.Duration(randomDelay(1, 3)) * time.Second)

	if err := bot.page.Click("button[type='submit']", playwright.PageClickOptions{Timeout: playwright.Float(60000)}); err != nil {
		return fmt.Errorf("failed to click login button: %w", err)
	}

	// Wait for 2FA or home page
	time.Sleep(time.Duration(randomDelay(3, 5)) * time.Second)
	
	twoFactorInput, err := bot.page.QuerySelector("input[name='verificationCode']")
	if err != nil {
		log.Printf("⚠️ Error checking for 2FA: %v", err)
	} else if twoFactorInput != nil {
		if bot.totpSecret == "" {
			return fmt.Errorf("2FA required, but TOTP secret not provided")
		}
		otp, err := totp.GenerateCode(bot.totpSecret, time.Now())
		if err != nil {
			return fmt.Errorf("failed to generate OTP: %w", err)
		}
		log.Printf("🔐 Entering 2FA code: %s", otp)
		if err := twoFactorInput.Fill(otp); err != nil {
			return fmt.Errorf("failed to fill 2FA code: %w", err)
		}
		if err := bot.page.Click("button:has-text('Confirm')"); err != nil {
			return fmt.Errorf("failed to click 2FA confirm button: %w", err)
		}
	}

	if _, err := bot.page.WaitForSelector("a[aria-label='Home']", playwright.PageWaitForSelectorOptions{Timeout: playwright.Float(60000)}); err != nil {
		errorMessage, errQuery := bot.page.QuerySelector("[data-testid='login-error-message']")
		if errQuery == nil && errorMessage != nil {
			text, _ := errorMessage.InnerText()
			return fmt.Errorf("login failed: %s", text)
		}
		return fmt.Errorf("login failed, home icon not found: %w", err)
	}

	log.Println("✅ Login successful.")

	// Handle "Save your login info?" dialog
	time.Sleep(time.Duration(randomDelay(2, 4)) * time.Second)
	saveInfoButton, err := bot.page.QuerySelector("button:has-text('Save info')")
	if err != nil {
		log.Printf("⚠️ Error checking for save info button: %v", err)
	} else if saveInfoButton != nil {
		if err := saveInfoButton.Click(); err != nil {
			log.Println("⚠️ Could not click 'Save info' button, but continuing.")
		}
	}

	// Handle "Turn on Notifications" dialog
	time.Sleep(time.Duration(randomDelay(2, 4)) * time.Second)
	notNowButton, err := bot.page.QuerySelector("button:has-text('Not Now')")
	if err != nil {
		log.Printf("⚠️ Error checking for notifications button: %v", err)
	} else if notNowButton != nil {
		if err := notNowButton.Click(); err != nil {
			log.Println("⚠️ Could not click 'Not Now' button, but continuing.")
		}
	}

	if err := bot.saveSession(); err != nil {
		log.Printf("⚠️ Failed to save session: %v", err)
	}

	return nil
}

func (bot *FollowerBot) sessionFilePath() string {
	return filepath.Join(bot.config.DataDir, "session.json")
}

func (bot *FollowerBot) loadSession() error {
	cookiesData, err := os.ReadFile(bot.sessionFilePath())
	if err != nil {
		return err
	}
	var cookies []playwright.Cookie
	if err := json.Unmarshal(cookiesData, &cookies); err != nil {
		return err
	}

	var optionalCookies []playwright.OptionalCookie
	for _, cookie := range cookies {
		optionalCookies = append(optionalCookies, playwright.OptionalCookie{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Domain:   &cookie.Domain,
			Path:     &cookie.Path,
			Expires:  &cookie.Expires,
			HttpOnly: &cookie.HttpOnly,
			Secure:   &cookie.Secure,
			SameSite: cookie.SameSite,
		})
	}

	if err := bot.context.AddCookies(optionalCookies); err != nil {
		return err
	}
	return nil
}

func (bot *FollowerBot) saveSession() error {
	cookies, err := bot.context.Cookies()
	if err != nil {
		return err
	}
	cookiesData, err := json.Marshal(cookies)
	if err != nil {
		return err
	}
	return os.WriteFile(bot.sessionFilePath(), cookiesData, 0644)
}

func (bot *FollowerBot) GetFollowing() ([]string, error) {
	log.Println("🤖 Getting following list...")
	return bot.scrapeList(bot.username, "following")
}

func (bot *FollowerBot) GetFollowers(username string) ([]string, error) {
	log.Printf("🤖 Getting followers of %s...", username)
	return bot.scrapeList(username, "followers")
}

func (bot *FollowerBot) scrapeList(username string, listType string) ([]string, error) {
	url := fmt.Sprintf("https://www.instagram.com/%s/", username)
	if _, err := bot.page.Goto(url); err != nil {
		return nil, fmt.Errorf("failed to navigate to profile page: %w", err)
	}

	linkSelector := fmt.Sprintf("a[href='/%s/%s/']", username, listType)
	if err := bot.page.Click(linkSelector); err != nil {
		// Fallback for private accounts or other issues
		if listType == "followers" {
			followersCountEl, errQuery := bot.page.QuerySelector(fmt.Sprintf("a[href='/%s/followers/'] > span", username))
			if errQuery == nil && followersCountEl != nil {
				followersCount, _ := followersCountEl.GetAttribute("title")
				log.Printf("   -> Could not click followers link, but found count: %s", followersCount)
				return nil, fmt.Errorf("could not click %s link, maybe profile is private", listType)
			}
		}
		return nil, fmt.Errorf("failed to click %s link: %w", listType, err)
	}

	// A more robust selector for the scrollable list container
	listSelector := "div[role='dialog'] ._aano"
	if _, err := bot.page.WaitForSelector(listSelector, playwright.PageWaitForSelectorOptions{Timeout: playwright.Float(10000)}); err != nil {
		return nil, fmt.Errorf("%s list dialog not found: %w", listType, err)
	}

	log.Println("   -> Scrolling through the list...")
	usernames := make(map[string]bool)
	lastHeight := -1
	consecutiveNoChange := 0

	for {
		handles, err := bot.page.QuerySelectorAll(listSelector + " a[role='link']")
		if err != nil {
			log.Printf("⚠️ Failed to get user handles, but continuing: %v", err)
		}

		for _, handle := range handles {
			username, err := handle.InnerText()
			if err == nil && username != "" {
				usernames[username] = true
			}
		}

		// Scroll the list
		if _, err = bot.page.Evaluate(fmt.Sprintf("() => { const el = document.querySelector('%s'); el.scrollTop = el.scrollHeight; }", listSelector)); err != nil {
			log.Printf("⚠️ Failed to scroll, but continuing: %v", err)
		}

		time.Sleep(time.Duration(randomDelay(1, 3)) * time.Second)

		newHeight, err := bot.page.Evaluate(fmt.Sprintf("() => document.querySelector('%s').scrollHeight", listSelector))
		if err != nil {
			log.Printf("⚠️ Failed to get scroll height, but continuing: %v", err)
			break
		}

		currentHeight := int(newHeight.(float64))
		if currentHeight == lastHeight {
			consecutiveNoChange++
		} else {
			consecutiveNoChange = 0
		}

		if consecutiveNoChange > 5 {
			log.Println("   -> List seems to have reached the end.")
			break
		}
		lastHeight = currentHeight

		// Safety break
		if len(usernames) > 2500 {
			log.Println("   -> Reached 2500 usernames, stopping scrape.")
			break
		}
	}

	result := make([]string, 0, len(usernames))
	for u := range usernames {
		result = append(result, u)
	}

	log.Printf("✅ Scraped %d usernames.", len(result))
	return result, nil
}

func (bot *FollowerBot) FollowUsers(accounts []string) error {
	if !bot.isWithinTimeWindow() {
		log.Println("🚫 Outside of allowed time window. Skipping follow session.")
		return nil
	}
	log.Println("🤖 Starting follow session...")

	// Shuffle accounts randomly
	shuffledAccounts := make([]string, len(accounts))
	copy(shuffledAccounts, accounts)
	rand.Shuffle(len(shuffledAccounts), func(i, j int) {
		shuffledAccounts[i], shuffledAccounts[j] = shuffledAccounts[j], shuffledAccounts[i]
	})

	accountsProcessed := 0
	for len(shuffledAccounts) > 0 {
		batchSize := randomDelay(3, 7)
		if batchSize > len(shuffledAccounts) {
			batchSize = len(shuffledAccounts)
		}
		batch := shuffledAccounts[:batchSize]
		shuffledAccounts = shuffledAccounts[batchSize:]

		log.Printf("📦 Processing a batch of %d users.", batchSize)
		for _, account := range batch {
			accountsProcessed++
			err := bot.followAccount(account)
			if err != nil {
				log.Printf("⚠️  Failed to follow %s: %v", account, err)
				if err.Error() == "daily follow limit reached" {
					log.Println("🚫 Daily follow limit reached. Stopping for today.")
					return saveBotState(bot.state, bot.config.DataDir)
				}
			}
			delay := randomDelay(5, 15)
			log.Printf("   ⏸️  Pausing for %d seconds... (%d/%d processed)", delay, accountsProcessed, len(accounts))
			time.Sleep(time.Duration(delay) * time.Second)
		}

		if len(shuffledAccounts) > 0 {
			pauseDuration := time.Duration(randomDelay(1*60, 3*60)) * time.Second
			log.Printf("   ⏸️  Pausing for %.1f minutes before next batch...", pauseDuration.Minutes())
			time.Sleep(pauseDuration)
		}
	}

	log.Println("✅ Follow session complete.")
	return saveBotState(bot.state, bot.config.DataDir)
}

func (bot *FollowerBot) followAccount(username string) error {
	maxFollows := randomDelay(25, 35)
	if bot.state.FollowsToday >= maxFollows {
		return fmt.Errorf("daily follow limit reached")
	}

	url := fmt.Sprintf("https://www.instagram.com/%s/", username)
	if _, err := bot.page.Goto(url); err != nil {
		return fmt.Errorf("failed to navigate to profile page: %w", err)
	}

	// More robust selector for the follow button
	followButton, err := bot.page.QuerySelector("div[role='main'] button:has-text('Follow'), div[role='main'] button:has-text('Follow back')")
	if err != nil {
		return fmt.Errorf("could not query for follow button: %w", err)
	}

	if followButton == nil {
		// Check if already following or requested
		alreadyFollowing, _ := bot.page.QuerySelector("div[role='main'] button:has-text('Following'), div[role='main'] button:has-text('Requested')")
		if alreadyFollowing != nil {
			log.Printf("   ℹ️  Already following or requested %s", username)
			return nil
		}
		return fmt.Errorf("follow button not found for %s", username)
	}

	time.Sleep(time.Duration(randomDelay(2, 5)) * time.Second)
	if err := followButton.Click(); err != nil {
		// Handle cases where the button might be obscured
		if _, err := bot.page.Evaluate("el => el.click()", followButton); err != nil {
			return fmt.Errorf("failed to click follow button with JS: %w", err)
		}
	}

	bot.state.FollowsToday++
	log.Printf("   ✅ Successfully followed %s (%d follows today)", username, bot.state.FollowsToday)
	return nil
}

func (bot *FollowerBot) UnfollowUsers(accounts []string) error {
	if !bot.isWithinTimeWindow() {
		log.Println("🚫 Outside of allowed time window. Skipping unfollow session.")
		return nil
	}
	log.Println("🤖 Starting unfollow session...")

	// Shuffle accounts randomly
	shuffledAccounts := make([]string, len(accounts))
	copy(shuffledAccounts, accounts)
	rand.Shuffle(len(shuffledAccounts), func(i, j int) {
		shuffledAccounts[i], shuffledAccounts[j] = shuffledAccounts[j], shuffledAccounts[i]
	})

	accountsProcessed := 0
	for len(shuffledAccounts) > 0 {
		batchSize := randomDelay(3, 7)
		if batchSize > len(shuffledAccounts) {
			batchSize = len(shuffledAccounts)
		}
		batch := shuffledAccounts[:batchSize]
		shuffledAccounts = shuffledAccounts[batchSize:]

		log.Printf("📦 Processing a batch of %d users to unfollow.", batchSize)
		for _, account := range batch {
			accountsProcessed++
			err := bot.unfollowAccount(account)
			if err != nil {
				log.Printf("⚠️  Failed to unfollow %s: %v", account, err)
				if err.Error() == "daily unfollow limit reached" {
					log.Println("🚫 Daily unfollow limit reached. Stopping for today.")
					return saveBotState(bot.state, bot.config.DataDir)
				}
			}
			delay := randomDelay(5, 15)
			log.Printf("   ⏸️  Pausing for %d seconds... (%d processed)", delay, accountsProcessed)
			time.Sleep(time.Duration(delay) * time.Second)
		}

		if len(shuffledAccounts) > 0 {
			pauseDuration := time.Duration(randomDelay(1*60, 3*60)) * time.Second
			log.Printf("   ⏸️  Pausing for %.1f minutes before next batch...", pauseDuration.Minutes())
			time.Sleep(pauseDuration)
		}
	}

	log.Println("✅ Unfollow session complete.")
	return saveBotState(bot.state, bot.config.DataDir)
}

func (bot *FollowerBot) unfollowAccount(username string) error {
	maxUnfollows := randomDelay(25, 35)
	if bot.state.UnfollowsToday >= maxUnfollows {
		return fmt.Errorf("daily unfollow limit reached")
	}

	url := fmt.Sprintf("https://www.instagram.com/%s/", username)
	if _, err := bot.page.Goto(url); err != nil {
		return fmt.Errorf("failed to navigate to profile page: %w", err)
	}

	// More robust selector for the unfollow button
	unfollowButton, err := bot.page.QuerySelector("div[role='main'] button:has-text('Following')")
	if err != nil {
		return fmt.Errorf("could not query for unfollow button: %w", err)
	}

	if unfollowButton == nil {
		log.Printf("   ℹ️  Not following %s", username)
		return nil
	}

	time.Sleep(time.Duration(randomDelay(2, 5)) * time.Second)
	if err := unfollowButton.Click(); err != nil {
		return fmt.Errorf("failed to click unfollow button: %w", err)
	}

	// Confirm unfollow
	confirmButton, err := bot.page.QuerySelector("button:has-text('Unfollow')")
	if err != nil {
		return fmt.Errorf("could not query for confirm unfollow button: %w", err)
	}
	if confirmButton == nil {
		return fmt.Errorf("confirm unfollow button not found")
	}

	time.Sleep(time.Duration(randomDelay(2, 5)) * time.Second)
	if err := confirmButton.Click(); err != nil {
		return fmt.Errorf("failed to click confirm unfollow button: %w", err)
	}

	bot.state.UnfollowsToday++
	log.Printf("   ✅ Successfully unfollowed %s (%d unfollows today)", username, bot.state.UnfollowsToday)
	return nil
}

func botStateFilePath(dataDir string) string {
	return filepath.Join(dataDir, "bot_state.json")
}

func loadBotState(dataDir string) *BotState {
	state := &BotState{}
	path := botStateFilePath(dataDir)
	data, err := os.ReadFile(path)
	if err != nil {
		log.Println("🤖 No bot state file found, creating a new one.")
		return state
	}

	if err := json.Unmarshal(data, state); err != nil {
		log.Printf("⚠️ Failed to unmarshal bot state: %v. Starting with a fresh state.", err)
		return &BotState{}
	}

	today := time.Now().Format("2006-01-02")
	if state.LastRunDate != today {
		log.Println("🤖 New day, resetting follow/unfollow count.")
		state.FollowsToday = 0
		state.UnfollowsToday = 0
		state.LastRunDate = today
	}

	log.Printf("✅ Bot state loaded. Follows today: %d, Unfollows today: %d", state.FollowsToday, state.UnfollowsToday)
	return state
}

func saveBotState(state *BotState, dataDir string) error {
	state.LastRunDate = time.Now().Format("2006-01-02")
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal bot state: %w", err)
	}
	return os.WriteFile(botStateFilePath(dataDir), data, 0644)
}

func randomDelay(min, max int) int {
	if min > max {
		min = max
	}
	if min == max {
		return min
	}
	return rand.Intn(max-min+1) + min
}
