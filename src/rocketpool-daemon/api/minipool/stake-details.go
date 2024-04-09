package minipool

import (
	"fmt"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/dao/oracle"
	"github.com/rocket-pool/rocketpool-go/v2/minipool"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/v2/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type minipoolStakeDetailsContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolStakeDetailsContextFactory) Create(args url.Values) (*minipoolStakeDetailsContext, error) {
	c := &minipoolStakeDetailsContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *minipoolStakeDetailsContextFactory) RegisterRoute(router *mux.Router) {
	RegisterMinipoolRoute[*minipoolStakeDetailsContext, api.MinipoolStakeDetailsData](
		router, "stake/details", f, f.handler.ctx, f.handler.logger, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolStakeDetailsContext struct {
	handler *MinipoolHandler
	rp      *rocketpool.RocketPool

	oSettings *oracle.OracleDaoSettings
}

func (c *minipoolStakeDetailsContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()

	// Bindings
	oMgr, err := oracle.NewOracleDaoManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating oDAO manager binding: %w", err)
	}
	c.oSettings = oMgr.Settings
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating oDAO settings binding: %w", err)
	}
	return types.ResponseStatus_Success, nil
}

func (c *minipoolStakeDetailsContext) GetState(node *node.Node, mc *batch.MultiCaller) {
	c.oSettings.Minipool.ScrubPeriod.AddToQuery(mc)
}

func (c *minipoolStakeDetailsContext) CheckState(node *node.Node, response *api.MinipoolStakeDetailsData) bool {
	return true
}

func (c *minipoolStakeDetailsContext) GetMinipoolDetails(mc *batch.MultiCaller, mp minipool.IMinipool, index int) {
	mpCommon := mp.Common()
	eth.AddQueryablesToMulticall(mc,
		mpCommon.Status,
		mpCommon.StatusTime,
	)
}

func (c *minipoolStakeDetailsContext) PrepareData(addresses []common.Address, mps []minipool.IMinipool, data *api.MinipoolStakeDetailsData) (types.ResponseStatus, error) {
	scrubPeriod := c.oSettings.Minipool.ScrubPeriod.Formatted()

	// Get the time of the latest block
	ctx := c.handler.ctx
	latestEth1Block, err := c.rp.Client.HeaderByNumber(ctx, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting the latest block header: %w", err)
	}
	latestBlockTime := time.Unix(int64(latestEth1Block.Time), 0)

	// Get the stake details
	details := make([]api.MinipoolStakeDetails, len(addresses))
	for i, mp := range mps {
		mpCommonDetails := mp.Common()
		mpDetails := api.MinipoolStakeDetails{
			Address: mpCommonDetails.Address,
		}

		mpDetails.State = mpCommonDetails.Status.Formatted()
		if mpDetails.State != rptypes.MinipoolStatus_Prelaunch {
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
	return types.ResponseStatus_Success, nil
}
