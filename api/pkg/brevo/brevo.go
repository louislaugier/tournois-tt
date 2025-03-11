package brevo

import (
	"context"

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
			// Close the body to prevent resource leaks
			defer resp.Body.Close()
		}
		return err
	}

	return nil
}
