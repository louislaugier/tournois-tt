package config

import (
	"os"
)

// Config holds all configuration for the application
type Config struct {
	ServerAddress string
	Environment   string
	// Add more configuration fields as needed
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		ServerAddress: getEnv("SERVER_ADDRESS", ":8000"),
		Environment:   getEnv("GIN_MODE", "debug"),
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
