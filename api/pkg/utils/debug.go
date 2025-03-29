package utils

import (
	"log"
	"os"
)

// IsDebugMode returns true if DEBUG environment variable is set
func IsDebugMode() bool {
	return os.Getenv("DEBUG") != ""
}

// DebugLog logs a message when in debug mode
func DebugLog(format string, args ...interface{}) {
	if IsDebugMode() {
		log.Printf("[DEBUG] "+format, args...)
	}
}
