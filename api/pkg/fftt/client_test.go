package fftt

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetClient(t *testing.T) {
	// Get the client
	client := GetClient()
	assert.NotNil(t, client, "GetClient() should not return nil")

	// Call again to test singleton behavior
	client2 := GetClient()
	assert.Equal(t, client, client2, "GetClient() should return the same instance on subsequent calls")

	// Make sure the HTTPClient is initialized
	clientImpl, ok := client.(*Client)
	assert.True(t, ok, "Client should be of type *Client")
	assert.NotNil(t, clientImpl.HTTPClient, "The HTTPClient should be initialized")
}

func TestGetTournaments(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the request URL and path
		assert.Equal(t, "/api/tournament_requests", r.URL.Path, "Unexpected request path")

		// Extract query parameters
		query := r.URL.Query()
		assert.Equal(t, "test", query.Get("param"), "Missing expected query parameter")

		// Verify request headers
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"), "Content-Type header should be set to application/json")
		assert.Equal(t, FFTT_REFERER_URL, r.Header.Get("Referer"), "Referer header should be set")

		// Send a mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `[{"id": 1, "name": "Test Tournament"}]`)
	}))
	defer server.Close()

	// Create a client and a transport that redirects to our test server
	client := &Client{
		HTTPClient: &http.Client{
			Transport: &mockTransport{server: server},
		},
	}

	// Create URL parameters
	params := url.Values{}
	params.Add("param", "test")

	// Call the method
	resp, err := client.GetTournaments(params)

	// Verify the response
	assert.NoError(t, err, "GetTournaments should not return an error")
	assert.NotNil(t, resp, "Response should not be nil")

	// Check response body
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "Reading response body should not fail")

	// Verify the content
	assert.Contains(t, string(body), `"id": 1`)
	assert.Contains(t, string(body), `"name": "Test Tournament"`)
}

func TestGetTournamentsError(t *testing.T) {
	// Create a client with a transport that always returns an error
	client := &Client{
		HTTPClient: &http.Client{
			Transport: &errorTransport{},
		},
	}

	// Call the method with empty parameters
	resp, err := client.GetTournaments(url.Values{})

	// Verify the response
	assert.Error(t, err, "GetTournaments should return an error")
	assert.Nil(t, resp, "Response should be nil when there's an error")
}

// mockTransport is a custom transport that redirects requests to our test server
type mockTransport struct {
	server *httptest.Server
}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Replace the URL with our test server's URL
	url := t.server.URL + req.URL.Path
	if req.URL.RawQuery != "" {
		url += "?" + req.URL.RawQuery
	}

	// Create a new request to our test server
	newReq, err := http.NewRequest(req.Method, url, req.Body)
	if err != nil {
		return nil, err
	}

	// Copy headers
	newReq.Header = req.Header

	// Send the request to our test server
	return http.DefaultTransport.RoundTrip(newReq)
}

// errorTransport is a custom transport that always returns an error
type errorTransport struct{}

func (t *errorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, &url.Error{
		Op:  "Get",
		URL: req.URL.String(),
		Err: error(errRequestFailed),
	}
}

type errorString string

func (e errorString) Error() string { return string(e) }

var errRequestFailed = errorString("request failed")
