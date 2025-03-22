package fftt

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

// mockClient is a wrapper for testing
type mockClient struct {
	mockGetTournamentsFn func(params url.Values) (*http.Response, error)
}

func (m *mockClient) GetTournaments(params url.Values) (*http.Response, error) {
	return m.mockGetTournamentsFn(params)
}

func TestFetchTournaments(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method and path
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/tournament_requests" {
			t.Errorf("Expected path /api/tournament_requests, got %s", r.URL.Path)
		}

		// Check query parameters
		q := r.URL.Query()
		if q.Get("itemsPerPage") != "10" {
			t.Errorf("Expected itemsPerPage=10, got %s", q.Get("itemsPerPage"))
		}

		// Return a mock response
		w.Header().Set("Content-Type", "application/json")
		tournaments := []Tournament{
			{
				ID:        1,
				Name:      "Tournament 1",
				Type:      "I",
				StartDate: "2023-01-01",
				EndDate:   "2023-01-02",
				Address: struct {
					StreetAddress             string  `json:"streetAddress"`
					PostalCode                string  `json:"postalCode"`
					AddressLocality           string  `json:"addressLocality"`
					DisambiguatingDescription string  `json:"disambiguatingDescription,omitempty"`
					Latitude                  float64 `json:"latitude,omitempty"`
					Longitude                 float64 `json:"longitude,omitempty"`
					Failed                    bool    `json:"failed,omitempty"`
				}{
					StreetAddress:   "123 Main St",
					PostalCode:      "75001",
					AddressLocality: "Paris",
				},
				Club: Club{
					ID:   1,
					Name: "Club 1",
					Code: "C1",
				},
				Endowment: 1000,
			},
		}
		json.NewEncoder(w).Encode(tournaments)
	}))
	defer server.Close()

	// Save original client and replace it after test
	originalClient := FFTTClient
	defer func() { FFTTClient = originalClient }()

	// Create mock client
	mockFFTT := &mockClient{
		mockGetTournamentsFn: func(params url.Values) (*http.Response, error) {
			req, err := http.NewRequest("GET", server.URL+"/api/tournament_requests", nil)
			if err != nil {
				return nil, err
			}
			req.URL.RawQuery = params.Encode()
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Referer", FFTT_REFERER_URL)
			return http.DefaultClient.Do(req)
		},
	}

	// Override the client for testing
	FFTTClient = mockFFTT

	// Test the fetch tournaments function
	params := url.Values{}
	params.Set("itemsPerPage", "10")

	// Important: Call FetchTournaments with the explicit client parameter
	// instead of relying on GetClient() to get the globally set client
	tournaments, err := FetchTournaments(params)

	// Check for errors
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check the results
	if len(tournaments) != 1 {
		t.Fatalf("Expected 1 tournament, got %d", len(tournaments))
	}

	if tournaments[0].ID != 1 {
		t.Errorf("Expected tournament ID 1, got %d", tournaments[0].ID)
	}

	if tournaments[0].Name != "Tournament 1" {
		t.Errorf("Expected tournament name 'Tournament 1', got %s", tournaments[0].Name)
	}
}

func TestGetFutureTournaments(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check query parameters
		q := r.URL.Query()

		// Verify the date format in the query parameters
		startDateAfter := q.Get("startDate[after]")
		if startDateAfter == "" {
			t.Errorf("Expected startDate[after] parameter, got none")
		}

		startDateBefore := q.Get("startDate[before]")
		if startDateBefore == "" {
			// Optional parameter, only check if it exists
			t.Logf("startDate[before] parameter not provided")
		} else {
			t.Logf("startDate[before] parameter: %s", startDateBefore)
		}

		// Return a mock response
		w.Header().Set("Content-Type", "application/json")
		tournaments := []Tournament{
			{
				ID:        2,
				Name:      "Future Tournament",
				Type:      "N",
				StartDate: "2023-06-01",
				EndDate:   "2023-06-02",
			},
		}
		json.NewEncoder(w).Encode(tournaments)
	}))
	defer server.Close()

	// Save original client and replace it after test
	originalClient := FFTTClient
	defer func() { FFTTClient = originalClient }()

	// Create mock client
	mockFFTT := &mockClient{
		mockGetTournamentsFn: func(params url.Values) (*http.Response, error) {
			req, err := http.NewRequest("GET", server.URL+"/api/tournament_requests", nil)
			if err != nil {
				return nil, err
			}
			req.URL.RawQuery = params.Encode()
			req.Header.Set("Accept", "application/json")
			req.Header.Set("Referer", FFTT_REFERER_URL)
			return http.DefaultClient.Do(req)
		},
	}

	// Override the client for testing
	FFTTClient = mockFFTT

	// Test GetFutureTournaments
	startDateAfter := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	startDateBefore := time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)

	tournaments, err := GetFutureTournaments(startDateAfter, &startDateBefore)

	// Check for errors
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check the results
	if len(tournaments) != 1 {
		t.Fatalf("Expected 1 tournament, got %d", len(tournaments))
	}

	if tournaments[0].ID != 2 {
		t.Errorf("Expected tournament ID 2, got %d", tournaments[0].ID)
	}

	if tournaments[0].Name != "Future Tournament" {
		t.Errorf("Expected tournament name 'Future Tournament', got %s", tournaments[0].Name)
	}
}
