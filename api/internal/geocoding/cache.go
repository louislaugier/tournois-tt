package geocoding

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const (
	cacheDir      = "cache"
	cacheFileName = "geocoding_cache.json"
)

func (c *GeocodingCache) SaveToFile() error {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}

	c.LastSave = time.Now()
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(cacheDir, cacheFileName), data, 0644)
}

func LoadCacheFromFile() (*GeocodingCache, error) {
	cache := &GeocodingCache{
		Addresses: make(map[string]Coordinates),
	}

	data, err := os.ReadFile(filepath.Join(cacheDir, cacheFileName))
	if err != nil {
		if os.IsNotExist(err) {
			return cache, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cache); err != nil {
		return nil, err
	}

	return cache, nil
}

func (c *GeocodingCache) Get(address string) (Coordinates, bool) {
	coords, exists := c.Addresses[address]
	return coords, exists
}

func (c *GeocodingCache) Set(address string, coords Coordinates) {
	c.Addresses[address] = coords
}
