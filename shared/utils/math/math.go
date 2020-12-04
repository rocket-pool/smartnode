package math

import (
    "math"
)


// Round a float64 down to a number of places
func RoundDown(val float64, places int) float64 {
    return math.Floor(val * math.Pow10(places)) / math.Pow10(places)
}

