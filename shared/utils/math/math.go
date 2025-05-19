package math

import (
	"math"
	"math/bits"
)

// Round a float64 down to a number of places
func RoundDown(val float64, places int) float64 {
	return math.Floor(val*math.Pow10(places)) / math.Pow10(places)
}

// Round a float64 up to a number of places
func RoundUp(val float64, places int) float64 {
	return math.Ceil(val*math.Pow10(places)) / math.Pow10(places)
}

func GetPowerOfTwoCeil(x uint64) uint64 {
	// Base case
	if x <= 1 {
		return 1
	}

	// Check if already a power of two
	if x&(x-1) == 0 {
		return x
	}

	// Find the most significant bit
	msb := bits.Len64(x) - 1
	return 1 << (msb + 1)
}
