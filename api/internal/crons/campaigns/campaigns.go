package campaigns

import (
	"log"
	"os"
	"strconv"
	"tournois-tt/api/pkg/brevo"
)

func sendCurrentCampaign() {
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
