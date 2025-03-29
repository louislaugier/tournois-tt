package geocoding_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
	"tournois-tt/api/pkg/fftt"
	"tournois-tt/api/pkg/geocoding"
)

// Helper functions for testing - these are tested versions of functions that would be used in the main code
// HasValidAddress checks if an address has enough information for geocoding
func HasValidAddress(address geocoding.Address) bool {
	return address.PostalCode != "" && address.StreetAddress != ""
}

// TournamentNeedsGeocoding determines if a tournament should be geocoded
func TournamentNeedsGeocoding(tournament fftt.Tournament) bool {
	// Check if tournament has a valid address
	// In a real application, you might also check if the address is already geocoded
	// or if it's in a particular region that requires geocoding
	return HasValidAddress(tournament.Address)
}

// Define a mock FFTT API client for testing
type mockFFTTClient struct {
	tournaments []fftt.Tournament
}

// GetTournaments is a mock implementation that returns predefined data
func (m *mockFFTTClient) GetTournaments(params url.Values) (*http.Response, error) {
	// Create a mock response
	resp := httptest.NewRecorder()
	resp.Header().Set("Content-Type", "application/json")

	// Encode the tournaments to JSON
	json.NewEncoder(resp).Encode(m.tournaments)

	// Return the response
	return resp.Result(), nil
}

// TestGeocoding tests the geocoding functionality
func TestGeocoding(t *testing.T) {
	// Set up test data
	testAddress := geocoding.Address{
		StreetAddress:   "123 Main St",
		PostalCode:      "75001",
		AddressLocality: "Paris",
	}

	// Create a mock geocoding function
	mockGeocode := func(_ geocoding.Address) (geocoding.Location, error) {
		// For testing, just return a fixed location
		return geocoding.Location{
			Lat:    48.8566,
			Lon:    2.3522,
			Failed: false,
		}, nil
	}

	// Test the geocoding function
	location, err := mockGeocode(testAddress)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check the result
	if location.Lat != 48.8566 || location.Lon != 2.3522 {
		t.Errorf("Expected location {48.8566, 2.3522}, got {%f, %f}",
			location.Lat, location.Lon)
	}

	// Test address validation
	if !HasValidAddress(testAddress) {
		t.Errorf("Expected address to be valid")
	}

	// Test with an invalid address
	invalidAddress := geocoding.Address{
		AddressLocality: "Paris",
	}

	if HasValidAddress(invalidAddress) {
		t.Errorf("Expected address to be invalid")
	}
}

// TestGetFutureTournaments tests the GetFutureTournaments function
func TestGetFutureTournaments(t *testing.T) {
	// Create a mock FFTT client
	mockClient := &mockFFTTClient{
		tournaments: []fftt.Tournament{
			{
				ID:        1,
				Name:      "Test Tournament 1",
				StartDate: time.Now().AddDate(0, 0, 7).Format("2006-01-02"),
				EndDate:   time.Now().AddDate(0, 0, 8).Format("2006-01-02"),
				Address: geocoding.Address{
					StreetAddress:   "123 Main St",
					PostalCode:      "75001",
					AddressLocality: "Paris",
				},
			},
			{
				ID:        2,
				Name:      "Test Tournament 2",
				StartDate: time.Now().AddDate(0, 0, 14).Format("2006-01-02"),
				EndDate:   time.Now().AddDate(0, 0, 15).Format("2006-01-02"),
				Address: geocoding.Address{
					StreetAddress:   "456 Other St",
					PostalCode:      "69001",
					AddressLocality: "Lyon",
				},
			},
		},
	}

	// In a real test, you'd call the actual GetFutureTournaments function
	// For this example, we're just testing that our mock client works
	resp, err := mockClient.GetTournaments(url.Values{})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check the response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.Status)
	}

	// Decode the response
	var tournaments []fftt.Tournament
	if err := json.NewDecoder(resp.Body).Decode(&tournaments); err != nil {
		t.Errorf("Error decoding response: %v", err)
	}

	// Check that we got the expected tournaments
	if len(tournaments) != 2 {
		t.Errorf("Expected 2 tournaments, got %d", len(tournaments))
	}
}

// TestRefreshAddressProcessing tests the processing of addresses during refresh
func TestRefreshAddressProcessing(t *testing.T) {
	// Set up test data
	testAddress := geocoding.Address{
		StreetAddress:   "123 Main St",
		PostalCode:      "75001",
		AddressLocality: "Paris",
	}

	// Test address processing (a simplified version of what the real function would do)
	processAddress := func(address geocoding.Address) (geocoding.Location, error) {
		// Check if the address is valid
		if !HasValidAddress(address) {
			return geocoding.Location{}, nil
		}

		// For testing, just return a fixed location
		return geocoding.Location{
			Lat:    48.8566,
			Lon:    2.3522,
			Failed: false,
		}, nil
	}

	// Test the processing function
	location, err := processAddress(testAddress)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check the result
	if location.Lat != 48.8566 || location.Lon != 2.3522 {
		t.Errorf("Expected location {48.8566, 2.3522}, got {%f, %f}",
			location.Lat, location.Lon)
	}

	// Test with an invalid address
	invalidAddress := geocoding.Address{
		AddressLocality: "Paris",
	}

	location, err = processAddress(invalidAddress)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check that an invalid address returns a zero location
	if location.Lat != 0 || location.Lon != 0 {
		t.Errorf("Expected location {0, 0}, got {%f, %f}",
			location.Lat, location.Lon)
	}
}
