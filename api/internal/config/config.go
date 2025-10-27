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
	InstagramAppID       string
	InstagramAppSecret   string
)

// Threads configuration
var (
	ThreadsEnabled     bool
	ThreadsAccessToken string
	ThreadsUserID      string
)

// Facebook configuration
var (
	FacebookEnabled     bool
	FacebookAccessToken string
	FacebookPageID      string
)

func init() {
	if err := godotenv.Load("../.env"); err != nil {
		godotenv.Load("./.env")
	}

	// Load Instagram configuration
	InstagramEnabled, _ = strconv.ParseBool(os.Getenv("INSTAGRAM_ENABLED"))
	InstagramAccessToken = os.Getenv("INSTAGRAM_ACCESS_TOKEN")
	InstagramPageID = os.Getenv("INSTAGRAM_PAGE_ID")
	InstagramAppID = os.Getenv("INSTAGRAM_APP_ID")
	InstagramAppSecret = os.Getenv("INSTAGRAM_APP_SECRET")

	// Load Threads configuration
	ThreadsEnabled, _ = strconv.ParseBool(os.Getenv("THREADS_ENABLED"))
	ThreadsAccessToken = os.Getenv("THREADS_ACCESS_TOKEN")
	ThreadsUserID = os.Getenv("THREADS_USER_ID")

	// Load Facebook configuration
	FacebookEnabled, _ = strconv.ParseBool(os.Getenv("FACEBOOK_ENABLED"))
	FacebookAccessToken = os.Getenv("FACEBOOK_ACCESS_TOKEN")
	FacebookPageID = os.Getenv("FACEBOOK_PAGE_ID")
}
