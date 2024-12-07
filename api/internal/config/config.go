package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment   string
	ServerAddress string
	FrontendURL   string
}

var FrontendURL string

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or error loading it: %v", err)
	}

	environment := getEnv("GIN_MODE", "debug")
	domain := getEnv("DOMAIN", "localhost:3000") // default value to local frontend

	// Construct frontend URL
	FrontendURL = fmt.Sprintf("http://%s", domain)
	if environment != "debug" {
		FrontendURL = fmt.Sprintf("https://%s", domain)
	}

	// Infer server address
	serverAddress := ":80"
	if environment != "debug" && domain != "localhost:3000" {
		cleanDomain := strings.Split(domain, ":")[0]
		serverAddress = fmt.Sprintf("https://api.%s", cleanDomain)
	}

	return &Config{
		Environment:   environment,
		ServerAddress: serverAddress,
		FrontendURL:   FrontendURL,
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
