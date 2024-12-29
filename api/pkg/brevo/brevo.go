package brevo

import (
	"context"
	"io"
	"log"

	brevo "github.com/getbrevo/brevo-go/lib"
)

func NewBrevoClient(apiKey string) *brevo.APIClient {
	cfg := brevo.NewConfiguration()
	cfg.AddDefaultHeader("api-key", apiKey)

	br := brevo.NewAPIClient(cfg)

	return br
}

func SendCampaign(b *brevo.APIClient, campaignID int) error {
	resp, err := b.EmailCampaignsApi.SendEmailCampaignNow(context.Background(), int64(campaignID))
	if err != nil {
		if resp != nil {
			// Log status code and status
			log.Printf("%v %v\n", resp.StatusCode, resp.Status)
			// Log each header key and value
			for h, v := range resp.Header {
				log.Printf("Header: %v: %v\n", h, v)
			}
			// If the body is not empty, read it
			bodyBytes, _ := io.ReadAll(resp.Body)
			// Close the body to prevent resource leaks
			defer resp.Body.Close()
			if len(bodyBytes) > 0 {
				log.Printf("Body: %s\n", string(bodyBytes))
			} else {
				log.Println("Body is empty")
			}
		}
		return err
	}

	return nil
}
