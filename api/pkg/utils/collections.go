package utils

// MapToSlice converts a map of values to a slice
// This is a generic helper function for working with maps and slices
func MapToSlice[K comparable, V any](m map[K]V) []V {
	slice := make([]V, 0, len(m))
	for _, v := range m {
		slice = append(slice, v)
	}
	return slice
}
