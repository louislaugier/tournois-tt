package config

import (
	"log"

	"github.com/joho/godotenv"
)

var FrontendURL = "http://frontend:3000"

func init() {
	if err := godotenv.Load("../.env"); err != nil {
		if err = godotenv.Load("./.env"); err != nil {
			log.Printf("No .env file found")
		}
	}
}
