package network

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type networkDelegateContextFactory struct {
	handler *NetworkHandler
}

func (f *networkDelegateContextFactory) Create(vars map[string]string) (*networkDelegateContext, error) {
	c := &networkDelegateContext{
		handler: f.handler,
	}
	return c, nil
}

// ===============
// === Context ===
// ===============

type networkDelegateContext struct {
	handler *NetworkHandler
}

func (c *networkDelegateContext) PrepareData(data *api.NetworkLatestDelegateData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	rp := sp.GetRocketPool()

	// Requirements
	err := sp.RequireEthClientSynced()
	if err != nil {
		return err
	}

	// Bindings
	delegateContract, err := rp.GetContract(rocketpool.ContractName_RocketMinipoolDelegate)
	if err != nil {
		return fmt.Errorf("error getting minipool delegate contract: %w", err)
	}

	data.Address = *delegateContract.Address
	return nil
}
