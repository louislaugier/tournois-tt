package crons

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
	"tournois-tt/api/pkg/brevo"
)

var executedToday bool

func Schedule() {
	// Get the France timezone (handles both CET and CEST automatically)
	location, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		fmt.Println("Error loading location:", err)
		return
	}

	for {
		now := time.Now().In(location)

		// Check if it's 12:30 PM local French time
		isTwelveThirty := now.Hour() == 12 && now.Minute() == 30

		if isTwelveThirty && !executedToday {
			executedToday = true
			go sendCampaign()
		}

		// Reset executedToday flag after 12:30 PM
		if now.Hour() != 12 || now.Minute() != 30 {
			executedToday = false
		}

		// Sleep for 55 seconds before checking again
		time.Sleep(55 * time.Second)
	}
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
