package geocoding

import (
	"fmt"
	"tournois-tt/api/internal/types"
	gcache "tournois-tt/api/pkg/geocoding/cache"
)

var Cache types.Cache

func init() {
	var err error
	Cache, err = gcache.LoadFromFile()
	if err != nil {
		fmt.Printf("Error loading geocoding cache: %v\n", err)
		Cache = gcache.NewCache()
	}
}
