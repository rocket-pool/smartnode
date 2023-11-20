package state

import (
	"math/big"
	"time"
)

const (
	threadLimit int = 10
)

// Global constants
var zero = big.NewInt(0)

// Converts a time on the chain (as Unix time in seconds) to a time.Time struct
func convertToTime(value *big.Int) time.Time {
	return time.Unix(value.Int64(), 0)
}

// Converts a duration on the chain (as a number of seconds) to a time.Duration struct
func convertToDuration(value *big.Int) time.Duration {
	return time.Duration(value.Uint64()) * time.Second
}
