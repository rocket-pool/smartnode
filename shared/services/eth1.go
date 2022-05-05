package services

import (
	"context"
	"math/big"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

func GetEthClientLatestBlockTimestamp(ec rocketpool.ExecutionClient) (uint64, error) {
	// Get latest block number
	blockNumber, err := ec.BlockNumber(context.Background())
	if err != nil {
		return 0, err
	}
	blockNumberBig := big.NewInt(0).SetUint64(blockNumber)

	// Get latest block
	header, err := ec.HeaderByNumber(context.Background(), blockNumberBig)
	if err != nil {
		return 0, err
	}

	// Return block timestamp
	return header.Time, nil
}
