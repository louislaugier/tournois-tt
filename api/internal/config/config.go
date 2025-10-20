package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var FrontendURL = "http://frontend:3000"

// Instagram configuration
var (
	InstagramEnabled     bool
	InstagramAccessToken string
	InstagramPageID      string
	InstagramRecipientID string
	InstagramAppID       string
	InstagramAppSecret   string
)

func init() {
	if err := godotenv.Load("../.env"); err != nil {
		godotenv.Load("./.env")
	}

	// Load Instagram configuration
	InstagramEnabled, _ = strconv.ParseBool(os.Getenv("INSTAGRAM_ENABLED"))
	InstagramAccessToken = os.Getenv("INSTAGRAM_ACCESS_TOKEN")
	InstagramPageID = os.Getenv("INSTAGRAM_PAGE_ID")
	InstagramRecipientID = os.Getenv("INSTAGRAM_RECIPIENT_ID")
	InstagramAppID = os.Getenv("INSTAGRAM_APP_ID")
	InstagramAppSecret = os.Getenv("INSTAGRAM_APP_SECRET")
}
