package network

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/rocketpool/common/beacon"
	rputils "github.com/rocket-pool/smartnode/rocketpool/utils/rp"
	"github.com/rocket-pool/smartnode/shared/config"
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

// ===============
// === Context ===
// ===============

type networkDepositInfoContext struct {
	handler *NetworkHandler
	rp      *rocketpool.RocketPool
	cfg     *config.RocketPoolConfig
	bc      beacon.Client
}

func (c *networkDepositInfoContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.cfg = sp.GetConfig()
	c.bc = c.handler.serviceProvider.GetBeaconClient()

	// Requirements
	return sp.RequireEthClientSynced()
}

func (c *networkDepositInfoContext) GetState(mc *batch.MultiCaller) {
}

func (c *networkDepositInfoContext) PrepareData(data *api.NetworkDepositContractInfoData, opts *bind.TransactOpts) error {
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
