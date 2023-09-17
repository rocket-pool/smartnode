package minipool

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings"
	"github.com/rocket-pool/rocketpool-go/types"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type minipoolStakeDetailsContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolStakeDetailsContextFactory) Create(vars map[string]string) (*minipoolStakeDetailsContext, error) {
	c := &minipoolStakeDetailsContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *minipoolStakeDetailsContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterMinipoolRoute[*minipoolStakeDetailsContext, api.MinipoolStakeDetailsData](
		router, "stake/details", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolStakeDetailsContext struct {
	handler *MinipoolHandler
	rp      *rocketpool.RocketPool

	oSettings *settings.OracleDaoSettings
}

func (c *minipoolStakeDetailsContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()

	// Requirements
	err := errors.Join(
		sp.RequireNodeRegistered(),
	)
	if err != nil {
		return err
	}

	// Bindings
	c.oSettings, err = settings.NewOracleDaoSettings(c.rp)
	if err != nil {
		return fmt.Errorf("error creating oDAO settings binding: %w", err)
	}
	return nil
}

func (c *minipoolStakeDetailsContext) GetState(node *node.Node, mc *batch.MultiCaller) {
	c.oSettings.GetScrubPeriod(mc)
}

func (c *minipoolStakeDetailsContext) CheckState(node *node.Node, response *api.MinipoolStakeDetailsData) bool {
	return true
}

func (c *minipoolStakeDetailsContext) GetMinipoolDetails(mc *batch.MultiCaller, mp minipool.Minipool, index int) {
	mpCommon := mp.GetMinipoolCommon()
	mpCommon.GetStatus(mc)
	mpCommon.GetStatusTime(mc)
}

func (c *minipoolStakeDetailsContext) PrepareData(addresses []common.Address, mps []minipool.Minipool, data *api.MinipoolStakeDetailsData) error {
	scrubPeriod := c.oSettings.Details.Minipools.ScrubPeriod.Formatted()

	// Get the time of the latest block
	latestEth1Block, err := c.rp.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("error getting the latest block header: %w", err)
	}
	latestBlockTime := time.Unix(int64(latestEth1Block.Time), 0)

	// Get the stake details
	details := make([]api.MinipoolStakeDetails, len(addresses))
	for i, mp := range mps {
		mpCommonDetails := mp.GetMinipoolCommon().Details
		mpDetails := api.MinipoolStakeDetails{
			Address: mpCommonDetails.Address,
		}

		mpDetails.State = mpCommonDetails.Status.Formatted()
		if mpDetails.State != types.Prelaunch {
			mpDetails.InvalidState = true
		} else {
			creationTime := mpCommonDetails.StatusTime.Formatted()
			mpDetails.RemainingTime = creationTime.Add(scrubPeriod).Sub(latestBlockTime)
			if mpDetails.RemainingTime > 0 {
				mpDetails.StillInScrubPeriod = true
			}
		}

		mpDetails.CanStake = !(mpDetails.InvalidState || mpDetails.StillInScrubPeriod)
		details[i] = mpDetails
	}

	// Update & return response
	data.Details = details
	return nil
}
