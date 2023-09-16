package network

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
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
	rp      *rocketpool.RocketPool

	delegateContract *core.Contract
}

func (c *networkDelegateContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()

	// Requirements
	err := sp.RequireEthClientSynced()
	if err != nil {
		return err
	}

	// Bindings
	c.delegateContract, err = c.rp.GetContract(rocketpool.ContractName_RocketMinipoolDelegate)
	if err != nil {
		return fmt.Errorf("error getting minipool delegate contract: %w", err)
	}
	return nil
}

func (c *networkDelegateContext) GetState(mc *batch.MultiCaller) {
}

func (c *networkDelegateContext) PrepareData(data *api.NetworkLatestDelegateData, opts *bind.TransactOpts) error {
	data.Address = *c.delegateContract.Address
	return nil
}
