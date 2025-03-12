package logger

import (
	"log"

	pw "github.com/playwright-community/playwright-go"
)

// SetupPageLogging configures logging for a Playwright page
func SetupPageLogging(page pw.Page) {
	// Enable console logging from the page
	page.On("console", func(msg pw.ConsoleMessage) {
		log.Printf("Browser console [%s]: %s", msg.Type(), msg.Text())
	})

	// Enable request/response logging
	page.On("request", func(req pw.Request) {
		log.Printf("Request sent: %s %s", req.Method(), req.URL())
	})
	page.On("response", func(res pw.Response) {
		log.Printf("Response received: %d %s", res.Status(), res.URL())
	})
}

// Info logs an informational message
func Info(format string, v ...interface{}) {
	log.Printf("INFO: "+format, v...)
}

// Error logs an error message
func Error(format string, v ...interface{}) {
	log.Printf("ERROR: "+format, v...)
}

// Debug logs a debug message
func Debug(format string, v ...interface{}) {
	log.Printf("DEBUG: "+format, v...)
}
