package crons

import (
	"log"
	"time"

	_ "time/tzdata"

	"github.com/robfig/cron/v3"
)

func Schedule() {
	location, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		log.Fatal("Error loading Paris time zone:", err)
	}

	// Initialize a new cron scheduler with the Paris time zone
	c := cron.New(cron.WithLocation(location))

	// Schedule the cron job to run every day at 1 PM
	// _, err = c.AddFunc("0 13 * * *", sendCurrentCampaign)
	// if err != nil {
	// 	log.Fatal("Error adding cron job:", err)
	// }

	// Schedule the cron job to run every day at 1 PM
	_, err = c.AddFunc("0 13 * * *", RefreshTournaments)
	if err != nil {
		log.Fatal("Error adding cron job:", err)
	}

	// Start the cron scheduler in a separate goroutine
	go func() {
		c.Start()
	}()
}
