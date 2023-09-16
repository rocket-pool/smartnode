package network

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/rocketpool/common/rewards"
	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/config"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// ===============
// === Factory ===
// ===============

type networkDownloadRewardsContextFactory struct {
	handler *NetworkHandler
}

func (f *networkDownloadRewardsContextFactory) Create(vars map[string]string) (*networkDownloadRewardsContext, error) {
	c := &networkDownloadRewardsContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("interval", vars, cliutils.ValidateUint, &c.interval),
	}
	return c, errors.Join(inputErrs...)
}

// ===============
// === Context ===
// ===============

type networkDownloadRewardsContext struct {
	handler     *NetworkHandler
	rp          *rocketpool.RocketPool
	cfg         *config.RocketPoolConfig
	nodeAddress common.Address

	interval uint64
}

func (c *networkDownloadRewardsContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.cfg = sp.GetConfig()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	return sp.RequireNodeRegistered()
}

func (c *networkDownloadRewardsContext) GetState(mc *batch.MultiCaller) {
}

func (c *networkDownloadRewardsContext) PrepareData(data *api.SuccessData, opts *bind.TransactOpts) error {
	// Get the event info for the interval
	intervalInfo, err := rewards.GetIntervalInfo(c.rp, c.cfg, c.nodeAddress, c.interval, nil)
	if err != nil {
		return fmt.Errorf("error getting interval %d info: %w", c.interval, err)
	}

	// Download the rewards file
	err = rewards.DownloadRewardsFile(c.cfg, c.interval, intervalInfo.CID, true)
	if err != nil {
		return fmt.Errorf("error downloading interval %d rewards file: %w", c.interval, err)
	}
	data.Success = true
	return nil
}
