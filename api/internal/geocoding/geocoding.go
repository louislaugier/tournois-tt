package geocoding

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

var cache *GeocodingCache

func init() {
	var err error
	cache, err = LoadCacheFromFile()
	if err != nil {
		fmt.Printf("Error loading geocoding cache: %v\n", err)
		cache = &GeocodingCache{
			Addresses: make(map[string]Coordinates),
		}
	}
}

func geocodeAddressWithRetry(address string) (Coordinates, error) {
	baseURL := "https://nominatim.openstreetmap.org/search"
	client := &http.Client{Timeout: 10 * time.Second}

	params := url.Values{}
	params.Add("q", address)
	params.Add("format", "json")
	params.Add("limit", "1")

	req, err := http.NewRequest("GET", baseURL+"?"+params.Encode(), nil)
	if err != nil {
		return Coordinates{}, err
	}

	req.Header.Set("User-Agent", "TournoisTT/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return Coordinates{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Coordinates{}, fmt.Errorf("geocoding request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Coordinates{}, err
	}

	var result nominatimResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return Coordinates{}, err
	}

	if len(result) == 0 {
		return Coordinates{Failed: true}, nil
	}

	var lat, lon float64
	fmt.Sscanf(result[0].Lat, "%f", &lat)
	fmt.Sscanf(result[0].Lon, "%f", &lon)

	return Coordinates{Lat: lat, Lon: lon}, nil
}

// GetCoordinates returns the coordinates for a given address, using cache if available
func GetCoordinates(addr Address) (Coordinates, error) {
	if !hasValidAddress(addr) {
		return Coordinates{Failed: true}, fmt.Errorf("invalid address")
	}
	return GeocodeAddress(addr)
}

func GeocodeAddress(addr Address) (Coordinates, error) {
	variants := generateAddressVariants(addr)

	// Check cache first for all variants
	for _, variant := range variants {
		if coords, exists := cache.Get(variant); exists {
			return coords, nil
		}
	}

	// Try geocoding each variant
	for _, variant := range variants {
		coords, err := geocodeAddressWithRetry(variant)
		if err != nil {
			continue
		}

		// Cache the result for all variants
		for _, v := range variants {
			cache.Set(v, coords)
		}

		// Save cache periodically
		if time.Since(cache.LastSave) > 5*time.Minute {
			if err := cache.SaveToFile(); err != nil {
				fmt.Printf("Error saving geocoding cache: %v\n", err)
			}
		}

		return coords, nil
	}

	// If all variants fail, mark as failed in cache
	failedCoords := Coordinates{Failed: true}
	for _, variant := range variants {
		cache.Set(variant, failedCoords)
	}

	return failedCoords, nil
}
