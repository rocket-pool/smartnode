package services

import (
	"context"
	"time"

	"github.com/rocket-pool/rocketpool-go/core"
)

// Confirm the EC's latest block is within the threshold of the current system clock
func IsSyncWithinThreshold(ec core.ExecutionClient) (bool, time.Time, error) {
	// Get latest block
	header, err := ec.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return false, time.Time{}, err
	}

	// Return true if the latest block is under the threshold
	blockTime := time.Unix(int64(header.Time), 0)
	isWithinThreshold := time.Since(blockTime) < ethClientRecentBlockThreshold
	return isWithinThreshold, blockTime, nil
}
