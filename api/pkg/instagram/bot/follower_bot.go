package bot

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
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
	LastRunDate       string    `json:"last_run_date"`
	FollowsToday      int       `json:"follows_today"`
	UnfollowsToday    int       `json:"unfollows_today"`
	RateLimitedUntil  time.Time `json:"rate_limited_until,omitempty"`
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
		UserAgent: playwright.String("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
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
		if _, err := bot.page.WaitForSelector("a[href='/'] > svg[aria-label='Home']", playwright.PageWaitForSelectorOptions{Timeout: playwright.Float(5000)}); err == nil {
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
	cookieButtonSelector := "button:has-text('Allow all cookies')"
	cookieButton, err := bot.page.QuerySelector(cookieButtonSelector)
	if err == nil && cookieButton != nil {
		log.Println("🍪 Cookie consent dialog found. Clicking 'Allow all cookies'.")
		if err := cookieButton.Click(); err != nil {
			log.Println("⚠️ Could not click 'Allow all cookies' button, but continuing.")
		}
		// Wait for the dialog to disappear by waiting for the selector to be hidden
		if _, err := bot.page.WaitForSelector(cookieButtonSelector, playwright.PageWaitForSelectorOptions{State: playwright.WaitForSelectorStateHidden, Timeout: playwright.Float(5000)}); err != nil {
			log.Println("⚠️ Cookie dialog did not disappear as expected.")
		}
	} else {
		log.Println("ℹ️ No cookie consent dialog found, or it timed out.")
	}

	if _, err := bot.page.WaitForSelector("input[name='username']"); err != nil {
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

	if err := bot.page.Click("button[type='submit']"); err != nil {
		return fmt.Errorf("failed to click login button: %w", err)
	}

	// Wait for 2FA, home page, or an error message
	if err := bot.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{State: playwright.LoadStateNetworkidle}); err != nil {
		log.Printf("⚠️ Error waiting for network idle after login click: %v", err)
	}

	// Check for multiple possible outcomes after login attempt
	homeNavSelector := "a[aria-label='Home']"
	twoFactorSelector := "input[name='verificationCode']"
	loginErrorSelector := "[data-testid='login-error-message']"

	// Wait for any of the selectors to appear
	_, err = bot.page.WaitForSelector(
		fmt.Sprintf("%s, %s, %s", homeNavSelector, twoFactorSelector, loginErrorSelector),
		playwright.PageWaitForSelectorOptions{Timeout: playwright.Float(120000)},
	)
	if err != nil {
		return fmt.Errorf("timed out waiting for login result (home, 2FA, or error): %w", err)
	}

	// Outcome 1: Login Error
	errorMessage, _ := bot.page.QuerySelector(loginErrorSelector)
	if errorMessage != nil {
		text, _ := errorMessage.InnerText()
		return fmt.Errorf("login failed with message: %s", text)
	}

	// Outcome 2: Two-Factor Authentication
	twoFactorInput, _ := bot.page.QuerySelector(twoFactorSelector)
	if twoFactorInput != nil {
		html, _ := bot.page.Content()
		log.Printf("HTML content at 2FA stage: %s", html)
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
		log.Println("Attempting to click 2FA confirm button")
		time.Sleep(2 * time.Second) // Add a small delay
		confirmButton, err := bot.page.QuerySelector("form div[role='button']:has-text('Confirm')")
		if err != nil || confirmButton == nil {
			// Fallback to the old selector just in case
			confirmButton, err = bot.page.QuerySelector("button:has-text('Confirm')")
			if err != nil || confirmButton == nil {
				return fmt.Errorf("2FA confirm button not found: %w", err)
			}
		}
		if err := confirmButton.DispatchEvent("click"); err != nil {
			return fmt.Errorf("failed to click 2FA confirm button: %w", err)
		}
	}

	// Wait for final navigation to home page
	if _, err := bot.page.WaitForSelector(homeNavSelector, playwright.PageWaitForSelectorOptions{Timeout: playwright.Float(60000)}); err != nil {
		return fmt.Errorf("login failed, home icon not found after potential 2FA: %w", err)
	}

	log.Println("✅ Login successful.")

	// Handle "Save your login info?" dialog
	var saveInfoButton playwright.ElementHandle
	saveInfoButton, err = bot.page.QuerySelector("button:has-text('Save Login Info')")
	if err == nil && saveInfoButton != nil {
		if err := saveInfoButton.Click(); err != nil {
			log.Println("⚠️ Could not click 'Save info' button, but continuing.")
		}
	}

	// Handle "Turn on Notifications" dialog
	notNowButton, err := bot.page.QuerySelector("button:has-text('Not Now'), button:has-text('Turn on notifications')")
	if err == nil && notNowButton != nil {
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
		// Get all links in the list dialog that are not the user's own profile link
		handles, err := bot.page.QuerySelectorAll(listSelector + " a[role='link']:not([href*='/?next='])")
		if err != nil {
			log.Printf("⚠️ Failed to get user handles, but continuing: %v", err)
		}

		for _, handle := range handles {
			// Extract username from the href attribute for reliability
			href, err := handle.GetAttribute("href")
			if err != nil || href == "" {
				continue
			}
			// Expected href format: /username/
			parts := strings.Split(strings.Trim(href, "/"), "/")
			if len(parts) > 0 {
				username := parts[0]
				if username != "" {
					usernames[username] = true
				}
			}
		}

		// Scroll the list
		if _, err = bot.page.Evaluate(fmt.Sprintf("() => { const el = document.querySelector('%s'); el.scrollTop = el.scrollHeight; }", listSelector)); err != nil {
			log.Printf("⚠️ Failed to scroll, but continuing: %v", err)
		}

		// Wait for a moment to let new content load
		time.Sleep(time.Duration(randomDelay(1, 2)) * time.Second)

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

		if consecutiveNoChange > 3 { // Reduced threshold for faster completion
			log.Println("   -> List seems to have reached the end.")
			break
		}
		lastHeight = currentHeight

		// Safety break
		if len(usernames) > 3000 { // Increased limit slightly
			log.Println("   -> Reached 3000 usernames, stopping scrape.")
			break
		}
	}

	result := make([]string, 0, len(usernames))
	for u := range usernames {
		result = append(result, u)
	}

	log.Printf("✅ Scraped %d usernames.", len(result))
	// Close the dialog
	if err := bot.page.Click("div[role='dialog'] button[aria-label='Close']"); err != nil {
		log.Printf("⚠️ Could not close the %s list dialog, but continuing.", listType)
		// As a fallback, we can try to navigate away
		bot.page.Goto(fmt.Sprintf("https://www.instagram.com/%s/", bot.username))
	}
	return result, nil
}

func (bot *FollowerBot) FollowUsers(accounts []string) error {
	if !bot.isWithinTimeWindow() {
		log.Println("🚫 Outside of allowed time window. Skipping follow session.")
		return nil
	}
	
	// Check if we're currently rate limited
	if !bot.state.RateLimitedUntil.IsZero() && time.Now().Before(bot.state.RateLimitedUntil) {
		waitDuration := time.Until(bot.state.RateLimitedUntil)
		log.Printf("🚫 Rate limited until %s (%.0f minutes remaining). Skipping follow session.", 
			bot.state.RateLimitedUntil.Format("15:04"), waitDuration.Minutes())
		return nil
	}
	
	// Clear rate limit if we've passed the time
	if !bot.state.RateLimitedUntil.IsZero() && time.Now().After(bot.state.RateLimitedUntil) {
		log.Println("✅ Rate limit period has passed. Resuming normal operation.")
		bot.state.RateLimitedUntil = time.Time{}
		if err := saveBotState(bot.state, bot.config.DataDir); err != nil {
			log.Printf("⚠️ Failed to save state after clearing rate limit: %v", err)
		}
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
				if err.Error() == "rate limit detected" {
					log.Println("⏳ Rate limit detected. Pausing for 2 hours.")
					bot.state.RateLimitedUntil = time.Now().Add(2 * time.Hour)
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
	if _, err := bot.page.Goto(url, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateNetworkidle}); err != nil {
		return fmt.Errorf("failed to navigate to profile page: %w", err)
	}

	// More robust selector for the follow button, targeting the header area
	followButtonSelector := "header button:has-text('Follow'), header button:has-text('Follow back')"
	followButton, err := bot.page.QuerySelector(followButtonSelector)
	if err != nil {
		return fmt.Errorf("could not query for follow button: %w", err)
	}

	if followButton == nil {
		// Check if already following, requested, or if it's our own profile
		alreadyFollowingSelector := "header button:has-text('Following'), header button:has-text('Requested'), a[href='/accounts/edit/']"
		alreadyFollowing, _ := bot.page.QuerySelector(alreadyFollowingSelector)
		if alreadyFollowing != nil {
			log.Printf("   ℹ️  Already following, requested, or own profile: %s", username)
			return nil
		}
		return fmt.Errorf("follow button not found for %s", username)
	}

	// Wait for the button to be enabled
	if err := followButton.WaitForElementState("visible"); err != nil {
		return fmt.Errorf("follow button not visible: %w", err)
	}
	time.Sleep(time.Duration(randomDelay(1, 2)) * time.Second) // Small random delay before click

	if err := followButton.Click(); err != nil {
		// Handle cases where the button might be obscured
		if _, err := bot.page.Evaluate("el => el.click()", followButton); err != nil {
			return fmt.Errorf("failed to click follow button with JS: %w", err)
		}
	}

	// Check for rate limit alert modal
	time.Sleep(2 * time.Second) // Wait for potential modal to appear
	alertModalSelector := "div[role='dialog']:has-text('Try Again Later'), div[role='dialog']:has-text('Action Blocked'), div[role='dialog']:has-text('temporarily blocked')"
	alertModal, err := bot.page.QuerySelector(alertModalSelector)
	if err == nil && alertModal != nil {
		log.Printf("🚫 Rate limit alert detected for %s. Closing modal and setting 2-hour wait.", username)
		// Try to close the modal by clicking OK or Close button
		closeButton, _ := bot.page.QuerySelector("div[role='dialog'] button:has-text('OK'), div[role='dialog'] button:has-text('Close')")
		if closeButton != nil {
			closeButton.Click()
			time.Sleep(1 * time.Second)
		}
		return fmt.Errorf("rate limit detected")
	}

	// Verify the button text changes to 'Following' or 'Requested'
	if _, err := bot.page.WaitForSelector("header button:has-text('Following'), header button:has-text('Requested')", playwright.PageWaitForSelectorOptions{Timeout: playwright.Float(5000)}); err != nil {
		log.Printf("⚠️  Follow confirmation for %s not found, but proceeding.", username)
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
	if _, err := bot.page.Goto(url, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateNetworkidle}); err != nil {
		return fmt.Errorf("failed to navigate to profile page: %w", err)
	}

	// More robust selector for the unfollow button in the header
	unfollowButtonSelector := "header button:has-text('Following'), header button:has-text('Requested')"
	unfollowButton, err := bot.page.QuerySelector(unfollowButtonSelector)
	if err != nil {
		return fmt.Errorf("could not query for unfollow button: %w", err)
	}

	if unfollowButton == nil {
		log.Printf("   ℹ️  Not following %s (or button not found)", username)
		return nil
	}

	// Wait for the button to be visible and then click
	if err := unfollowButton.WaitForElementState("visible"); err != nil {
		return fmt.Errorf("unfollow button not visible: %w", err)
	}
	time.Sleep(time.Duration(randomDelay(1, 2)) * time.Second)
	if err := unfollowButton.Click(); err != nil {
		return fmt.Errorf("failed to click unfollow button: %w", err)
	}

	// Confirm unfollow in the dialog
	confirmButtonSelector := "div[role='dialog'] button:has-text('Unfollow')"
	confirmButton, err := bot.page.QuerySelector(confirmButtonSelector)
	if err != nil {
		return fmt.Errorf("could not query for confirm unfollow button: %w", err)
	}
	if confirmButton == nil {
		return fmt.Errorf("confirm unfollow button not found in dialog")
	}

	// Wait for the confirm button to be visible and then click
	if err := confirmButton.WaitForElementState("visible"); err != nil {
		return fmt.Errorf("confirm unfollow button not visible: %w", err)
	}
	time.Sleep(time.Duration(randomDelay(1, 2)) * time.Second)
	if err := confirmButton.Click(); err != nil {
		return fmt.Errorf("failed to click confirm unfollow button: %w", err)
	}

	// Wait a moment and check if the unfollow actually succeeded
	time.Sleep(2 * time.Second)
	
	// Check if button changed back to 'Follow' to confirm unfollow succeeded
	followButton, err := bot.page.QuerySelector("header button:has-text('Follow')")
	if err == nil && followButton != nil {
		bot.state.UnfollowsToday++
		log.Printf("   ✅ Successfully unfollowed %s (%d unfollows today)", username, bot.state.UnfollowsToday)
		return nil
	}
	
	// If button didn't change, unfollow likely didn't work (rate limited silently)
	log.Printf("   ⚠️  Unfollow button didn't change for %s - likely rate limited. Continuing without incrementing counter.", username)
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
