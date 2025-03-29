package fftt

import (
	"net/http"
	"net/url"
	"sync"
)

// API URL constants
const (
	// FFTT_API_BASE_URL is the base URL for FFTT API
	FFTT_API_BASE_URL = "https://apiv2.fftt.com/api"
	// FFTT_TOURNAMENT_ENDPOINT is the endpoint for tournament requests
	FFTT_TOURNAMENT_ENDPOINT = "/tournament_requests"
	// FFTT_REFERER_URL is the referer URL for FFTT API requests
	FFTT_REFERER_URL = "https://monclub.fftt.com/"
)

// FFTTClientInterface defines the interface for the FFTT client
type FFTTClientInterface interface {
	GetTournaments(params url.Values) (*http.Response, error)
}

// Client implements the FFTTClientInterface
type Client struct {
	HTTPClient *http.Client
}

// Global client instance for the application
var FFTTClient FFTTClientInterface
var initOnce sync.Once

// GetClient returns the singleton instance of the FFTT client
func GetClient() FFTTClientInterface {
	initOnce.Do(func() {
		FFTTClient = &Client{
			HTTPClient: &http.Client{},
		}
	})
	return FFTTClient
}

// GetTournaments fetches tournaments from the FFTT API
func (c *Client) GetTournaments(params url.Values) (*http.Response, error) {
	// Construct the full URL with constants
	requestURL := FFTT_API_BASE_URL + FFTT_TOURNAMENT_ENDPOINT

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, err
	}

	// Set query parameters if any
	if params != nil {
		req.URL.RawQuery = params.Encode()
	}

	// Set required headers
	req.Header.Set("Referer", FFTT_REFERER_URL)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	return c.HTTPClient.Do(req)
}
