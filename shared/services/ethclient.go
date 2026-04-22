package services

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

type EthClient struct {
	*ethclient.Client
}

func NewEthClient(url string) (*EthClient, error) {
	ec, err := ethclient.Dial(url)
	if err != nil {
		return nil, err
	}
	return &EthClient{ec}, nil
}

func (c *EthClient) LatestBlockTime(ctx context.Context) (time.Time, error) {
	header, err := c.HeaderByNumber(ctx, nil)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(int64(header.Time), 0), nil
}
