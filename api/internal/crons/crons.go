package crons

import (
	"log"
	"time"
	instagramCron "tournois-tt/api/internal/crons/instagram"
	"tournois-tt/api/internal/crons/tournaments"

	_ "time/tzdata"

	"github.com/robfig/cron/v3"
)

func Schedule() {
	location, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		log.Fatal("Error loading Europe/Paris time zone:", err)
	}

	// Initialize a new cron scheduler with the Paris time zone
	c := cron.New(cron.WithLocation(location))

	// Schedule the cron job to run every day at 1 PM
	// _, err = c.AddFunc("0 13 * * *", sendCurrentCampaign)
	// if err != nil {
	// 	log.Fatal("Error adding cron job:", err)
	// }

	// Schedule the cron job to run every 5 minutes
	_, err = c.AddFunc("*/5 * * * *", tournaments.RefreshListWithGeocoding)
	if err != nil {
		log.Fatal("Error adding cron job:", err)
	}

	// Schedule the cron job to run every day at 1 AM
	// _, err = c.AddFunc("0 1 * * *", tournaments.RefreshSignupURLs)
	// if err != nil {
	// 	log.Fatal("Error adding cron job:", err)
	// }

	// Schedule Instagram token refresh to run daily at 3 AM Paris time
	_, err = c.AddFunc("0 3 * * *", instagramCron.CheckAndRefreshToken)
	if err != nil {
		log.Fatal("Error adding Instagram token refresh cron job:", err)
	}

	// Schedule Threads token refresh to run daily at 3:15 AM Paris time
	_, err = c.AddFunc("15 3 * * *", instagramCron.CheckAndRefreshThreadsToken)
	if err != nil {
		log.Fatal("Error adding Threads token refresh cron job:", err)
	}

	// Check tokens on startup (in background)
	go instagramCron.RefreshTokenOnStartup()
	go instagramCron.RefreshThreadsTokenOnStartup()

	// Start the cron scheduler in a separate goroutine
	go func() {
		c.Start()
		log.Println("✅ All cron jobs started successfully")
	}()
}
