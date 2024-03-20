package utils

import (
	"context"
	"time"
)

// Sleeps for the specified time, but can break out if the provided context is cancelled.
// Returns true if the context is cancelled, false if it's not and the full period was slept.
// TODO: move to NMC
func SleepWithCancel(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	select {
	case <-ctx.Done():
		// Cancel occurred
		timer.Stop()
		return true

	case <-timer.C:
		// Duration has passed without a cancel
		return false
	}
}
