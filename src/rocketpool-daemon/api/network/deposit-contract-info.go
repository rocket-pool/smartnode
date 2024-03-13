package network

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/node-manager-core/api/server"
	rputils "github.com/rocket-pool/smartnode/rocketpool-daemon/common/utils"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type networkDepositInfoContextFactory struct {
	handler *NetworkHandler
}

func (f *networkDepositInfoContextFactory) Create(args url.Values) (*networkDepositInfoContext, error) {
	c := &networkDepositInfoContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *networkDepositInfoContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*networkDepositInfoContext, api.NetworkDepositContractInfoData](
		router, "deposit-contract-info", f, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type networkDepositInfoContext struct {
	handler *NetworkHandler
}

func (c *networkDepositInfoContext) PrepareData(data *api.NetworkDepositContractInfoData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	rp := sp.GetRocketPool()
	cfg := sp.GetConfig()
	bc := sp.GetBeaconClient()

	// Requirements
	err := sp.RequireEthClientSynced(c.handler.context)
	if err != nil {
		return err
	}

	// Get the deposit contract info
	info, err := rputils.GetDepositContractInfo(c.handler.context, rp, cfg, bc)
	if err != nil {
		return fmt.Errorf("error getting deposit contract info: %w", err)
	}
	data.SufficientSync = true
	data.RPNetwork = info.RPNetwork
	data.RPDepositContract = info.RPDepositContract
	data.BeaconNetwork = info.BeaconNetwork
	data.BeaconDepositContract = info.BeaconDepositContract
	return nil
}
