package services

import (
	"context"
	"github.com/urfave/cli"
	"math/big"
)

func GetEthClientLatestBlockTimestamp(c *cli.Context) (uint64, error) {
	// Get eth client
	var err error
	ec, err := GetEthClientProxy(c)
	if err != nil {
		return 0, err
	}

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
