package network

import (
	"fmt"

	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/smartnode/rocketpool/common/beacon"
	rputils "github.com/rocket-pool/smartnode/rocketpool/utils/rp"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type networkDepositInfoContextFactory struct {
	handler *NetworkHandler
}

func (f *networkDepositInfoContextFactory) Create(vars map[string]string) (*networkDepositInfoContext, error) {
	c := &networkDepositInfoContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *networkDepositInfoContextFactory) Run(c *networkDepositInfoContext) (*api.ApiResponse[api.NetworkDepositContractInfoData], error) {
	return runNetworkCall[api.NetworkDepositContractInfoData](c)
}

// ===============
// === Context ===
// ===============

type networkDepositInfoContext struct {
	handler *NetworkHandler
	bc      beacon.Client
	*commonContext
}

func (c *networkDepositInfoContext) CreateBindings(ctx *commonContext) error {
	c.commonContext = ctx
	c.bc = c.handler.serviceProvider.GetBeaconClient()
	return nil
}

func (c *networkDepositInfoContext) GetState(mc *batch.MultiCaller) {
}

func (c *networkDepositInfoContext) PrepareData(data *api.NetworkDepositContractInfoData) error {
	// Get the deposit contract info
	info, err := rputils.GetDepositContractInfo(c.rp, c.cfg, c.bc)
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
