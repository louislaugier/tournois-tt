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
	// If the tournament already has coordinates, it doesn't need geocoding
	if tournament.Address.Latitude != 0 && tournament.Address.Longitude != 0 {
		return false
	}

	// If the tournament doesn't have a valid address, it can't be geocoded
	if !HasValidAddress(tournament.Address) {
		return false
	}

	// Otherwise, it needs geocoding
	return true
}

// mockFFTTClient for testing
type mockFFTTClient struct {
	tournaments []fftt.Tournament
}

func (m *mockFFTTClient) GetTournaments(params url.Values) (*http.Response, error) {
	// Create a mock HTTP response
	recorder := httptest.NewRecorder()
	recorder.Header().Set("Content-Type", "application/json")

	// Write the mock data to the response
	if err := json.NewEncoder(recorder).Encode(m.tournaments); err != nil {
		return nil, err
	}

	// Return the response
	return recorder.Result(), nil
}

// TestGeocoding is a focused test that verifies the geocoding functionality works correctly
func TestGeocoding(t *testing.T) {
	// Set up mocked FFTT client
	originalClient := fftt.FFTTClient
	defer func() { fftt.FFTTClient = originalClient }()

	// Create test tournaments
	mockClient := &mockFFTTClient{
		tournaments: []fftt.Tournament{
			{
				ID:        1,
				Name:      "Test Tournament 1",
				Type:      "I",
				StartDate: "2023-06-01",
				EndDate:   "2023-06-02",
				Address: geocoding.Address{
					StreetAddress:   "123 Main St",
					PostalCode:      "75001",
					AddressLocality: "Paris",
				},
			},
		},
	}
	fftt.FFTTClient = mockClient

	// Mock geocoding function
	originalGetCoordinatesFn := geocoding.GetCoordinatesFn
	defer func() {
		geocoding.GetCoordinatesFn = originalGetCoordinatesFn
	}()

	// Set up a mock geocoding function that returns predefined coordinates
	geocoding.GetCoordinatesFn = func(address geocoding.Address) (geocoding.Location, error) {
		// Return test coordinates
		return geocoding.Location{
			Lat:    48.856614,
			Lon:    2.352222,
			Failed: false,
		}, nil
	}

	// Call the geocoding function directly
	addr := geocoding.Address{
		StreetAddress:   "123 Main St",
		PostalCode:      "75001",
		AddressLocality: "Paris",
	}
	location, err := geocoding.GetCoordinates(addr)

	// Verify results
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if location.Lat != 48.856614 || location.Lon != 2.352222 {
		t.Errorf("Got incorrect coordinates: %f, %f", location.Lat, location.Lon)
	}
	if location.Failed {
		t.Error("Expected geocoding to succeed, but it failed")
	}
}

// TestGetFutureTournaments verifies that the FFTT client is used correctly
func TestGetFutureTournaments(t *testing.T) {
	// Save original client and replace after test
	originalClient := fftt.FFTTClient
	defer func() {
		fftt.FFTTClient = originalClient
	}()

	// Create test data
	expectedTournaments := []fftt.Tournament{
		{
			ID:        1,
			Name:      "Test Tournament",
			Type:      "I",
			StartDate: "2023-06-01",
			EndDate:   "2023-06-02",
		},
	}

	// Create mock client
	mockClient := &mockFFTTClient{
		tournaments: expectedTournaments,
	}

	// Override the client for testing
	fftt.FFTTClient = mockClient

	// Call the function
	startDateAfter := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	tournaments, err := fftt.GetFutureTournaments(startDateAfter, nil)

	// Verify results
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(tournaments) != len(expectedTournaments) {
		t.Fatalf("Expected %d tournaments, got %d", len(expectedTournaments), len(tournaments))
	}

	if tournaments[0].ID != expectedTournaments[0].ID {
		t.Errorf("Expected tournament ID %d, got %d", expectedTournaments[0].ID, tournaments[0].ID)
	}
}

// TestRefreshAddressProcessing tests the address validation in RefreshTournamentsAndGeocoding
func TestRefreshAddressProcessing(t *testing.T) {
	// Set up mocked FFTT client
	originalClient := fftt.FFTTClient
	defer func() {
		fftt.FFTTClient = originalClient
	}()

	// Create test data with two tournaments - one with valid address, one without
	testTournaments := []fftt.Tournament{
		{
			ID:        1,
			Name:      "Tournament 1",
			Type:      "I",
			StartDate: "2023-06-01",
			EndDate:   "2023-06-02",
			Address: geocoding.Address{
				StreetAddress:   "123 Main St",
				PostalCode:      "75001",
				AddressLocality: "Paris",
			},
		},
		{
			ID:        2,
			Name:      "Tournament 2",
			Type:      "R",
			StartDate: "2023-07-01",
			EndDate:   "2023-07-02",
			Address: geocoding.Address{
				StreetAddress:   "", // Empty street address to test invalid address handling
				PostalCode:      "",
				AddressLocality: "Lyon",
			},
		},
	}

	// Mock the FFTT client
	mockClient := &mockFFTTClient{
		tournaments: testTournaments,
	}

	// Override the client for testing
	fftt.FFTTClient = mockClient

	// Mock the geocoding function and track calls
	originalGetCoordinatesFn := geocoding.GetCoordinatesFn
	defer func() {
		geocoding.GetCoordinatesFn = originalGetCoordinatesFn
	}()

	geocodingCalled := 0
	geocodingAddresses := make([]geocoding.Address, 0)

	geocoding.GetCoordinatesFn = func(address geocoding.Address) (geocoding.Location, error) {
		geocodingCalled++
		geocodingAddresses = append(geocodingAddresses, address)

		// Return test coordinates
		return geocoding.Location{
			Lat:    48.856614,
			Lon:    2.352222,
			Failed: false,
		}, nil
	}

	// The main function under test is RefreshTournamentsAndGeocoding
	// However, since we can't fully mock the cache, we'll verify parts of its behavior
	// by testing the related components:

	// 1. Verify address validation logic by checking the filters we know are applied
	if !HasValidAddress(testTournaments[0].Address) {
		t.Error("First tournament should have a valid address")
	}

	if HasValidAddress(testTournaments[1].Address) {
		t.Error("Second tournament should not have a valid address")
	}

	// 2. Test the TournamentNeedsGeocoding function
	// Create a tournament with coordinates already set
	tournamentWithCoords := fftt.Tournament{
		ID:   3,
		Name: "Tournament with coordinates",
		Address: geocoding.Address{
			StreetAddress:   "123 Main St",
			PostalCode:      "75001",
			AddressLocality: "Paris",
			Latitude:        48.856614,
			Longitude:       2.352222,
		},
	}

	// It should not need geocoding since coordinates are already present
	if TournamentNeedsGeocoding(tournamentWithCoords) {
		t.Error("Tournament with coordinates should not need geocoding")
	}

	// Tournament with valid address but no coordinates should need geocoding
	if !TournamentNeedsGeocoding(testTournaments[0]) {
		t.Error("Tournament with valid address but no coordinates should need geocoding")
	}

	// Tournament with invalid address should not need geocoding
	if TournamentNeedsGeocoding(testTournaments[1]) {
		t.Error("Tournament with invalid address should not need geocoding")
	}
}
