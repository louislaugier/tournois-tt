package nominatim

import (
	"net/http"
	"time"
	"tournois-tt/api/internal/types"
)

func NewClient() *types.NominatimClient {
	return &types.NominatimClient{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		BaseURL:    "https://nominatim.openstreetmap.org/search",
	}
}
