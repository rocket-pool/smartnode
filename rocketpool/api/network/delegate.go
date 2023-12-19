package network

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type networkDelegateContextFactory struct {
	handler *NetworkHandler
}

func (f *networkDelegateContextFactory) Create(args url.Values) (*networkDelegateContext, error) {
	c := &networkDelegateContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *networkDelegateContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*networkDelegateContext, api.NetworkLatestDelegateData](
		router, "latest-delegate", f, f.handler.serviceProvider,
	)
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
