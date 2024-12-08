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
	if err := godotenv.Load("../.env"); err != nil {
		if err = godotenv.Load("./.env"); err != nil {
			log.Fatal("godotenv.Load: ", err)
		}
	}

	domain := getEnv("DOMAIN", "localhost:3000") // default value to local frontend
	isLocal := domain == "localhost:3000"

	// Infer environment based on domain
	environment := "debug"
	if !isLocal {
		environment = "release"
	}

	isRelease := environment == "release"

	// Construct frontend URL
	FrontendURL = fmt.Sprintf("http://%s", domain)
	if isRelease {
		FrontendURL = fmt.Sprintf("https://%s", domain)
	}

	// Infer server address
	serverAddress := ":80"
	if isRelease && !isLocal {
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
