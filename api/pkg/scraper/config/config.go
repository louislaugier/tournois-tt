package config

// BrowserConfig holds the configuration for browser setup
type BrowserConfig struct {
	Headless  bool
	UserAgent string
}

// DefaultConfig returns the default browser configuration
func DefaultConfig() BrowserConfig {
	return BrowserConfig{
		Headless:  true,
		UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
	}
}

// BrowserArgs returns the default browser launch arguments
func BrowserArgs() []string {
	return []string{
		"--disable-dev-shm-usage",
		"--no-sandbox",
		"--disable-setuid-sandbox",
		"--disable-gpu",
		"--no-first-run",
		"--no-zygote",
		"--single-process",
		"--disable-extensions",
	}
}
