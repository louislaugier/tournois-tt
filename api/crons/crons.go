package crons

import (
	"log"
	"os"
	"strconv"
	"time"
	"tournois-tt/api/pkg/brevo"

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
	_, err = c.AddFunc("0 13 * * *", sendCampaign)
	if err != nil {
		log.Fatal("Error adding cron job:", err)
	}

	// Start the cron scheduler in a separate goroutine
	go func() {
		c.Start()
	}()
}

func sendCampaign() {
	log.Println("Sending campaign now.")

	cl := brevo.NewBrevoClient(os.Getenv("BREVO_API_KEY"))

	campaignID, err := strconv.Atoi(os.Getenv("BREVO_CAMPAIGN_ID"))
	if err != nil {
		log.Println(err)
		return
	}

	err = brevo.SendCampaign(cl, campaignID)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("Campaign sent successfully.")
}
