package network

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/rocketpool-go/v2/rewards"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
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
		router, "rewards-file-info", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type networkRewardsFileContext struct {
	handler *NetworkHandler
	rp      *rocketpool.RocketPool
	cfg     *config.SmartNodeConfig

	index       uint64
	rewardsPool *rewards.RewardsPool
}

func (c *networkRewardsFileContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.cfg = sp.GetConfig()

	// Requirements
	status, err := sp.RequireRocketPoolContracts(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	c.rewardsPool, err = rewards.NewRewardsPool(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting rewards pool binding: %w", err)
	}
	return types.ResponseStatus_Success, nil
}

func (c *networkRewardsFileContext) GetState(mc *batch.MultiCaller) {
	c.rewardsPool.RewardIndex.AddToQuery(mc)
}

func (c *networkRewardsFileContext) PrepareData(data *api.NetworkRewardsFileData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	data.CurrentIndex = c.rewardsPool.RewardIndex.Formatted()

	// Get the path of the file to save
	filePath := c.cfg.GetRewardsTreePath(c.index)
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		data.TreeFileExists = false
	} else {
		data.TreeFileExists = true
	}
	return types.ResponseStatus_Success, nil
}
