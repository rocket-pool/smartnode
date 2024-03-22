package minipool

import (
	"errors"
	"net/url"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type minipoolDissolveDetailsContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolDissolveDetailsContextFactory) Create(args url.Values) (*minipoolDissolveDetailsContext, error) {
	c := &minipoolDissolveDetailsContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *minipoolDissolveDetailsContextFactory) RegisterRoute(router *mux.Router) {
	RegisterMinipoolRoute[*minipoolDissolveDetailsContext, api.MinipoolDissolveDetailsData](
		router, "dissolve/details", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolDissolveDetailsContext struct {
	handler *MinipoolHandler
}

func (c *minipoolDissolveDetailsContext) Initialize() error {
	sp := c.handler.serviceProvider

	// Requirements
	return errors.Join(
		sp.RequireNodeRegistered(),
	)
}

func (c *minipoolDissolveDetailsContext) GetState(node *node.Node, mc *batch.MultiCaller) {
}

func (c *minipoolDissolveDetailsContext) CheckState(node *node.Node, response *api.MinipoolDissolveDetailsData) bool {
	return true
}

func (c *minipoolDissolveDetailsContext) GetMinipoolDetails(mc *batch.MultiCaller, mp minipool.IMinipool, index int) {
	mpCommon := mp.Common()
	eth.AddQueryablesToMulticall(mc,
		mpCommon.NodeAddress,
		mpCommon.Status,
	)
}

func (c *minipoolDissolveDetailsContext) PrepareData(addresses []common.Address, mps []minipool.IMinipool, data *api.MinipoolDissolveDetailsData) error {
	details := make([]api.MinipoolDissolveDetails, len(mps))
	for i, mp := range mps {
		mpCommonDetails := mp.Common()
		status := mpCommonDetails.Status.Formatted()
		mpDetails := api.MinipoolDissolveDetails{
			Address:       mpCommonDetails.Address,
			InvalidStatus: !(status == types.MinipoolStatus_Initialized || status == types.MinipoolStatus_Prelaunch),
		}
		mpDetails.CanDissolve = !mpDetails.InvalidStatus
		details[i] = mpDetails
	}

	data.Details = details
	return nil
}
