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
			Locations: make(map[string]Location),
			Aliases:   make(map[string]string),
		}
	}
}

func geocodeAddressWithRetry(address string) (Location, error) {
	baseURL := "https://nominatim.openstreetmap.org/search"
	client := &http.Client{Timeout: 10 * time.Second}

	params := url.Values{}
	params.Add("q", address)
	params.Add("format", "json")
	params.Add("limit", "1")

	req, err := http.NewRequest("GET", baseURL+"?"+params.Encode(), nil)
	if err != nil {
		return Location{}, err
	}

	req.Header.Set("User-Agent", "TournoisTT/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return Location{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Location{}, fmt.Errorf("geocoding request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Location{}, err
	}

	var result nominatimResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return Location{}, err
	}

	if len(result) == 0 {
		return Location{Failed: true, LastUpdated: time.Now()}, nil
	}

	var lat, lon float64
	fmt.Sscanf(result[0].Lat, "%f", &lat)
	fmt.Sscanf(result[0].Lon, "%f", &lon)

	// Check if location is approximate (city-level only)
	isApproximate := result[0].Type == "city" ||
		result[0].Type == "administrative" ||
		result[0].Class == "boundary" ||
		result[0].Category == "boundary"

	return Location{
		Lat:         lat,
		Lon:         lon,
		Failed:      false,
		Approximate: isApproximate,
		LastUpdated: time.Now(),
	}, nil
}

// GetCoordinates returns the coordinates for a given address, using cache if available
func GetCoordinates(addr Address) (Location, error) {
	if !hasValidAddress(addr) {
		return Location{Failed: true, LastUpdated: time.Now()}, fmt.Errorf("invalid address")
	}
	return GeocodeAddress(addr)
}

func GeocodeAddress(addr Address) (Location, error) {
	variants := generateAddressVariants(addr)

	// Check cache first for all variants
	for _, variant := range variants {
		if loc, exists := cache.Get(variant); exists {
			return loc, nil
		}
	}

	// Try geocoding each variant
	for _, variant := range variants {
		loc, err := geocodeAddressWithRetry(variant)
		if err != nil {
			continue
		}

		// Cache the result for all variants
		for _, v := range variants {
			cache.Set(v, loc.Lat, loc.Lon, loc.Failed, loc.Approximate)
		}

		// Save cache periodically
		if time.Since(cache.LastSave) > 5*time.Minute {
			if err := cache.SaveToFile(); err != nil {
				fmt.Printf("Error saving geocoding cache: %v\n", err)
			}
		}

		return loc, nil
	}

	// If all variants fail, mark as failed in cache
	now := time.Now()
	failedLoc := Location{Failed: true, LastUpdated: now}
	for _, variant := range variants {
		cache.Set(variant, 0, 0, true, false)
	}

	return failedLoc, nil
}
