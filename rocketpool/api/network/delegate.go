package network

import (
	"fmt"

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

func (f *networkDelegateContextFactory) Run(c *networkDelegateContext) (*api.ApiResponse[api.NetworkLatestDelegateData], error) {
	return runNetworkCall[api.NetworkLatestDelegateData](c)
}

// ===============
// === Context ===
// ===============

type networkDelegateContext struct {
	handler          *NetworkHandler
	delegateContract *core.Contract
	*commonContext
}

func (c *networkDelegateContext) CreateBindings(ctx *commonContext) error {
	var err error
	c.commonContext = ctx

	c.delegateContract, err = c.rp.GetContract(rocketpool.ContractName_RocketMinipoolDelegate)
	if err != nil {
		return fmt.Errorf("error getting minipool delegate contract: %w", err)
	}
	return nil
}

func (c *networkDelegateContext) GetState(mc *batch.MultiCaller) {
}

func (c *networkDelegateContext) PrepareData(data *api.NetworkLatestDelegateData) error {
	data.Address = *c.delegateContract.Address
	return nil
}
