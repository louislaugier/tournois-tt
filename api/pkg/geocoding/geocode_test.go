package geocoding

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// mockProvider implements the Provider interface for testing
type mockProvider struct {
	name         string
	shouldFail   bool
	errorMessage string
}

func (p *mockProvider) GetCoordinates(address Address) (Location, error) {
	if p.shouldFail {
		return Location{Failed: true}, errors.New(p.errorMessage)
	}

	// Mock successful geocoding for testing
	return Location{
		Lat:    48.8566,
		Lon:    2.3522,
		Failed: false,
	}, nil
}

func (p *mockProvider) Name() string {
	return p.name
}

// TestGeocodingProviderSingletons verifies that the geocoding providers are initialized as singletons
func TestGeocodingProviderSingletons(t *testing.T) {
	// Since nominatimProvider and googleProvider are package variables
	// initialized in init(), they should be the same instances across multiple calls

	// First verification
	nominatim1 := nominatimProvider
	google1 := googleProvider

	assert.NotNil(t, nominatim1, "Nominatim provider should be initialized")
	assert.NotNil(t, google1, "Google provider should be initialized")

	// Second verification (should be the same instances)
	nominatim2 := nominatimProvider
	google2 := googleProvider

	// Verify exact same instances (pointer equality)
	assert.Equal(t, nominatim1, nominatim2, "Nominatim provider should be a singleton")
	assert.Equal(t, google1, google2, "Google provider should be a singleton")

	// Verify the provider names
	assert.Equal(t, "Nominatim", nominatim1.Name(), "Wrong provider name")
	assert.Equal(t, "Google", google1.Name(), "Wrong provider name")
}

func TestConstructFullAddress(t *testing.T) {
	testCases := []struct {
		name     string
		address  Address
		expected string
	}{
		{
			name: "Complete address",
			address: Address{
				StreetAddress:   "123 Rue de la Paix",
				PostalCode:      "75000",
				AddressLocality: "Paris",
			},
			expected: "123 Rue de la Paix, 75000 Paris, France",
		},
		{
			name: "No street address",
			address: Address{
				StreetAddress:             "",
				PostalCode:                "75000",
				AddressLocality:           "Paris",
				DisambiguatingDescription: "Eiffel Tower",
			},
			expected: "Eiffel Tower, 75000 Paris, France",
		},
		{
			name: "Street address without number",
			address: Address{
				StreetAddress:             "Avenue",
				PostalCode:                "75000",
				AddressLocality:           "Paris",
				DisambiguatingDescription: "Near Louvre",
			},
			expected: "Near Louvre Avenue, 75000 Paris, France",
		},
		{
			name: "Extra spaces in fields",
			address: Address{
				StreetAddress:   "  123 Rue de la Paix  ",
				PostalCode:      "  75000  ",
				AddressLocality: "  Paris  ",
			},
			expected: "123 Rue de la Paix, 75000 Paris, France",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ConstructFullAddress(tc.address)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCreateGeocodeResult(t *testing.T) {
	address := Address{
		StreetAddress:   "123 Rue de la Paix",
		PostalCode:      "75000",
		AddressLocality: "Paris",
	}

	t.Run("Success case", func(t *testing.T) {
		location := Location{
			Lat:    48.856614,
			Lon:    2.3522219,
			Failed: false,
		}

		result := CreateGeocodeResult(address, location, nil)

		assert.Equal(t, address, result.Address)
		assert.Equal(t, location.Lat, result.Latitude)
		assert.Equal(t, location.Lon, result.Longitude)
		assert.False(t, result.Failed)
		assert.NotZero(t, result.Timestamp)
		assert.WithinDuration(t, time.Now(), result.Timestamp, 2*time.Second)
	})

	t.Run("Error case", func(t *testing.T) {
		result := CreateGeocodeResult(address, Location{}, errors.New("geocoding failed"))

		assert.Equal(t, address, result.Address)
		assert.Equal(t, 0.0, result.Latitude)
		assert.Equal(t, 0.0, result.Longitude)
		assert.True(t, result.Failed)
		assert.NotZero(t, result.Timestamp)
		assert.WithinDuration(t, time.Now(), result.Timestamp, 2*time.Second)
	})

	t.Run("Failed location case", func(t *testing.T) {
		location := Location{
			Lat:    48.856614,
			Lon:    2.3522219,
			Failed: true,
		}

		result := CreateGeocodeResult(address, location, nil)

		assert.Equal(t, address, result.Address)
		assert.Equal(t, location.Lat, result.Latitude)
		assert.Equal(t, location.Lon, result.Longitude)
		assert.True(t, result.Failed)
		assert.NotZero(t, result.Timestamp)
		assert.WithinDuration(t, time.Now(), result.Timestamp, 2*time.Second)
	})
}

func TestIsAddressValid(t *testing.T) {
	testCases := []struct {
		name     string
		address  Address
		expected bool
	}{
		{
			name: "Valid address",
			address: Address{
				StreetAddress:   "123 Rue de la Paix",
				PostalCode:      "75000",
				AddressLocality: "Paris",
			},
			expected: true,
		},
		{
			name: "Missing postal code",
			address: Address{
				StreetAddress:   "123 Rue de la Paix",
				PostalCode:      "",
				AddressLocality: "Paris",
			},
			expected: false,
		},
		{
			name: "Missing locality",
			address: Address{
				StreetAddress:   "123 Rue de la Paix",
				PostalCode:      "75000",
				AddressLocality: "",
			},
			expected: false,
		},
		{
			name: "Missing street address",
			address: Address{
				StreetAddress:   "",
				PostalCode:      "75000",
				AddressLocality: "Paris",
			},
			expected: true, // Still valid because postal code and locality are present
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsAddressValid(tc.address)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetCoordinatesWithMockProvider(t *testing.T) {
	// Save original providers
	origNominatimProvider := nominatimProvider
	origGoogleProvider := googleProvider
	defer func() {
		// Restore original providers
		nominatimProvider = origNominatimProvider
		googleProvider = origGoogleProvider
	}()

	// Test case: Both providers succeed
	t.Run("Nominatim success", func(t *testing.T) {
		// Set up mock providers
		nominatimProvider = &mockProvider{name: "Nominatim", shouldFail: false}
		googleProvider = &mockProvider{name: "Google", shouldFail: false}

		address := Address{
			StreetAddress:   "123 Rue de la Paix",
			PostalCode:      "75000",
			AddressLocality: "Paris",
		}

		location, err := GetCoordinates(address)

		assert.NoError(t, err)
		assert.Equal(t, 48.856614, location.Lat)
		assert.Equal(t, 2.3522219, location.Lon)
		assert.False(t, location.Failed)
	})

	// Test case: Nominatim fails, Google succeeds
	t.Run("Nominatim fails, Google succeeds", func(t *testing.T) {
		// Set up mock providers
		nominatimProvider = &mockProvider{name: "Nominatim", shouldFail: true, errorMessage: "Nominatim error"}
		googleProvider = &mockProvider{name: "Google", shouldFail: false}

		address := Address{
			StreetAddress:   "123 Rue de la Paix",
			PostalCode:      "75000",
			AddressLocality: "Paris",
		}

		location, err := GetCoordinates(address)

		assert.NoError(t, err)
		assert.Equal(t, 48.856614, location.Lat)
		assert.Equal(t, 2.3522219, location.Lon)
		assert.False(t, location.Failed)
	})

	// Test case: Both providers fail
	t.Run("Both providers fail", func(t *testing.T) {
		// Set up mock providers
		nominatimProvider = &mockProvider{name: "Nominatim", shouldFail: true, errorMessage: "Nominatim error"}
		googleProvider = &mockProvider{name: "Google", shouldFail: true, errorMessage: "Google error"}

		address := Address{
			StreetAddress:   "123 Rue de la Paix",
			PostalCode:      "75000",
			AddressLocality: "Paris",
		}

		location, err := GetCoordinates(address)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Google error")
		assert.True(t, location.Failed)
	})
}
