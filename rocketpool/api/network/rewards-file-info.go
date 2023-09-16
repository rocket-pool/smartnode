package network

import (
	"errors"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/config"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// ===============
// === Factory ===
// ===============

type networkRewardsFileContextFactory struct {
	handler *NetworkHandler
}

func (f *networkRewardsFileContextFactory) Create(vars map[string]string) (*networkRewardsFileContext, error) {
	c := &networkRewardsFileContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("index", vars, cliutils.ValidateUint, &c.index),
	}
	return c, errors.Join(inputErrs...)
}

// ===============
// === Context ===
// ===============

type networkRewardsFileContext struct {
	handler *NetworkHandler
	rp      *rocketpool.RocketPool
	cfg     *config.RocketPoolConfig

	index       uint64
	rewardsPool *rewards.RewardsPool
}

func (c *networkRewardsFileContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.cfg = sp.GetConfig()

	// Requirements
	err := sp.RequireEthClientSynced()
	if err != nil {
		return err
	}

	// Bindings
	c.rewardsPool, err = rewards.NewRewardsPool(c.rp)
	if err != nil {
		return fmt.Errorf("error getting rewards pool binding: %w", err)
	}
	return nil
}

func (c *networkRewardsFileContext) GetState(mc *batch.MultiCaller) {
	c.rewardsPool.GetRewardIndex(mc)
}

func (c *networkRewardsFileContext) PrepareData(data *api.NetworkRewardsFileData, opts *bind.TransactOpts) error {
	data.CurrentIndex = c.rewardsPool.Details.RewardIndex.Formatted()

	// Get the path of the file to save
	filePath := c.cfg.Smartnode.GetRewardsTreePath(c.index, true)
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		data.TreeFileExists = false
	} else {
		data.TreeFileExists = true
	}
	return nil
}
