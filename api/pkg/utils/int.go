package utils

// Helper function to get absolute value
func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Min returns the minimum of two integers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Helper function to get maximum of two values
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
