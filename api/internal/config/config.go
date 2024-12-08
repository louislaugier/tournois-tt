package config

import (
	"github.com/joho/godotenv"
)

var FrontendURL = "http://frontend:3000"

func init() {
	if err := godotenv.Load("../.env"); err != nil {
		godotenv.Load("./.env")
	}
}
