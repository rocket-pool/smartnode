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
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type minipoolPromoteDetailsContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolPromoteDetailsContextFactory) Create(args url.Values) (*minipoolPromoteDetailsContext, error) {
	c := &minipoolPromoteDetailsContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *minipoolPromoteDetailsContextFactory) RegisterRoute(router *mux.Router) {
	RegisterMinipoolRoute[*minipoolPromoteDetailsContext, api.MinipoolPromoteDetailsData](
		router, "promote/details", f, f.handler.ctx, f.handler.logger, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolPromoteDetailsContext struct {
	handler *MinipoolHandler
	rp      *rocketpool.RocketPool

	oSettings *oracle.OracleDaoSettings
}

func (c *minipoolPromoteDetailsContext) Initialize() (types.ResponseStatus, error) {
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

func (c *minipoolPromoteDetailsContext) GetState(node *node.Node, mc *batch.MultiCaller) {
	c.oSettings.Minipool.PromotionScrubPeriod.AddToQuery(mc)
}

func (c *minipoolPromoteDetailsContext) CheckState(node *node.Node, response *api.MinipoolPromoteDetailsData) bool {
	return true
}

func (c *minipoolPromoteDetailsContext) GetMinipoolDetails(mc *batch.MultiCaller, mp minipool.IMinipool, index int) {
	mpv3, success := minipool.GetMinipoolAsV3(mp)
	if success {
		eth.AddQueryablesToMulticall(mc,
			mpv3.NodeAddress,
			mpv3.StatusTime,
			mpv3.IsVacant,
		)
	}
}

func (c *minipoolPromoteDetailsContext) PrepareData(addresses []common.Address, mps []minipool.IMinipool, data *api.MinipoolPromoteDetailsData) (types.ResponseStatus, error) {
	// Get the time of the latest block
	ctx := c.handler.ctx
	latestEth1Block, err := c.rp.Client.HeaderByNumber(ctx, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting the latest block time: %w", err)
	}
	latestBlockTime := time.Unix(int64(latestEth1Block.Time), 0)

	// Get the promotion details
	details := make([]api.MinipoolPromoteDetails, len(addresses))
	for i, mp := range mps {
		mpCommon := mp.Common()
		mpDetails := api.MinipoolPromoteDetails{
			Address:    mpCommon.Address,
			CanPromote: false,
		}

		// Check its eligibility
		mpv3, success := minipool.GetMinipoolAsV3(mps[i])
		if success && mpv3.IsVacant.Get() {
			creationTime := mpCommon.StatusTime.Formatted()
			remainingTime := creationTime.Add(c.oSettings.Minipool.ScrubPeriod.Formatted()).Sub(latestBlockTime)
			if remainingTime < 0 {
				mpDetails.CanPromote = true
			}
		}

		details[i] = mpDetails
	}

	data.Details = details
	return types.ResponseStatus_Success, nil
}
