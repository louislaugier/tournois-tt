package geocoding

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// GetCoordinates attempts to geocode an address using Nominatim
func GetCoordinates(address Address) (Location, error) {
	// Construct full address string
	fullAddress := address.StreetAddress

	// Add disambiguating description if street number is not in the address
	if address.StreetAddress == "" || !strings.Contains(address.StreetAddress, " ") {
		if address.DisambiguatingDescription != "" {
			fullAddress = address.DisambiguatingDescription + " " + fullAddress
		}
	}

	// Append postal code and locality
	fullAddress = fmt.Sprintf("%s, %s %s, France",
		strings.TrimSpace(fullAddress),
		strings.TrimSpace(address.PostalCode),
		strings.TrimSpace(address.AddressLocality))

	// Prepare Nominatim request
	params := url.Values{}
	params.Add("q", fullAddress)
	params.Add("format", "json")
	params.Add("limit", "1")

	resp, err := http.Get("https://nominatim.openstreetmap.org/search?" + params.Encode())
	if err != nil {
		return Location{Failed: true}, fmt.Errorf("failed to query Nominatim: %v", err)
	}
	defer resp.Body.Close()

	var results []struct {
		Lat string `json:"lat"`
		Lon string `json:"lon"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return Location{Failed: true}, fmt.Errorf("failed to parse Nominatim response: %v", err)
	}

	if len(results) == 0 {
		log.Printf("No coordinates found for address: %s", fullAddress)
		return Location{Failed: true}, fmt.Errorf("no coordinates found")
	}

	// Parse coordinates
	var lat, lon float64
	_, err = fmt.Sscanf(results[0].Lat, "%f", &lat)
	if err != nil {
		return Location{Failed: true}, fmt.Errorf("invalid latitude: %v", err)
	}
	_, err = fmt.Sscanf(results[0].Lon, "%f", &lon)
	if err != nil {
		return Location{Failed: true}, fmt.Errorf("invalid longitude: %v", err)
	}

	log.Printf("Geocoded address: %s -> (%.6f, %.6f)", fullAddress, lat, lon)

	return Location{
		Lat:    lat,
		Lon:    lon,
		Failed: false,
	}, nil
}
