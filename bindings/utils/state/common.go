package state

import (
	"math/big"
	"time"
)

const (
	threadLimit int = 10
)

// BigTime is a variable-sized big.Int from an evm 256-bit integer that represents a Unix time in seconds
type bigTime struct {
	big.Int
}

// BigDuration is a variable-sized big.Int from an evm 256-bit integer that represents a duration in seconds
type bigDuration struct {
	big.Int
}

func (b *bigTime) toTime() time.Time {
	return time.Unix(b.Int64(), 0)
}

func (b *bigDuration) toDuration() time.Duration {
	return time.Duration(b.Uint64()) * time.Second
}
