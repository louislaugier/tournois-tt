package geocoding

import (
	"tournois-tt/api/pkg/models"
)

// IsAddressValid checks if an address has enough data to be geocoded
func IsAddressValid(address models.Address) bool {
	if address.PostalCode == "" {
		return false
	}
	if address.AddressLocality == "" {
		return false
	}
	return true
}

// Note: The ConstructFullAddress function has been moved to geocoding.go to avoid duplication
