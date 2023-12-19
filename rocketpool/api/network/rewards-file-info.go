package network

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/config"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
)

// ===============
// === Factory ===
// ===============

type networkRewardsFileContextFactory struct {
	handler *NetworkHandler
}

func (f *networkRewardsFileContextFactory) Create(args url.Values) (*networkRewardsFileContext, error) {
	c := &networkRewardsFileContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("index", args, input.ValidateUint, &c.index),
	}
	return c, errors.Join(inputErrs...)
}

func (f *networkRewardsFileContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*networkRewardsFileContext, api.NetworkRewardsFileData](
		router, "rewards-file-info", f, f.handler.serviceProvider,
	)
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
	c.rewardsPool.RewardIndex.AddToQuery(mc)
}

func (c *networkRewardsFileContext) PrepareData(data *api.NetworkRewardsFileData, opts *bind.TransactOpts) error {
	data.CurrentIndex = c.rewardsPool.RewardIndex.Formatted()

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
