package services

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rocket-pool/smartnode/shared/units"
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

func (c *EthClient) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (units.Wei, error) {
	balance, err := c.Client.BalanceAt(ctx, account, blockNumber)
	if err != nil {
		return units.Wei{}, err
	}
	return units.NewWei(balance), nil
}

func (c *EthClient) LatestBlockTime(ctx context.Context) (time.Time, error) {
	header, err := c.HeaderByNumber(ctx, nil)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(int64(header.Time), 0), nil
}
