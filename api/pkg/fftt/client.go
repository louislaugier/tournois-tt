package fftt

import (
	"net/http"
	"net/url"
	"sync"
)

type Client struct {
	httpClient *http.Client
}

var (
	FFTTClient *Client
	once       sync.Once
)

func GetClient() *Client {
	once.Do(func() {
		FFTTClient = &Client{
			httpClient: &http.Client{},
		}
	})
	return FFTTClient
}

func (c *Client) GetTournaments(params url.Values) (*http.Response, error) {
	req, err := http.NewRequest("GET", "https://apiv2.fftt.com/api/tournament_requests", nil)
	if err != nil {
		return nil, err
	}

	// Add query parameters
	req.URL.RawQuery = params.Encode()

	// Add required headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://monclub.fftt.com/")

	return c.httpClient.Do(req)
}