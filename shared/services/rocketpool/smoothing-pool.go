package rocketpool

import (
	"context"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func GetSmoothingPoolBalance(rp rocketpool.RocketPool, ec *services.ExecutionClientManager) (*api.SmoothingRewardsResponse, error) {
	smoothingPoolContract, err := rp.GetContract("rocketSmoothingPool", nil)
	if err != nil {
		return nil, err
	}

	response := api.SmoothingRewardsResponse{}

	balanceWei, err := ec.BalanceAt(context.Background(), *smoothingPoolContract.Address, nil)
	if err != nil {
		return nil, err
	}
	response.EthBalance = balanceWei

	return &response, nil
}
