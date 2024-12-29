package crons

import (
	"log"
	"os"
	"strconv"
	"time"
	"tournois-tt/api/pkg/brevo"
)

var executedToday bool

func Schedule() {
	// location := time.FixedZone("UTC+5", 5*60*60)  // UTC +5 hours
	location := time.FixedZone("UTC+5:30", 5*60*60+30*60) // UTC +5 hours 30 minutes

	for {
		now := time.Now().In(location)
		isEightOClock := now.Hour() == 8 && now.Minute() == 00

		if isEightOClock && !executedToday {
			executedToday = true
			go sendCampaign()
		}
		if now.Hour() != 8 || now.Minute() != 0 {
			executedToday = false
		}

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
