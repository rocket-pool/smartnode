package minipool

import (
	"net/url"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/minipool"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	rptypes "github.com/rocket-pool/rocketpool-go/v2/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
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
		router, "dissolve/details", f, f.handler.ctx, f.handler.logger, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolDissolveDetailsContext struct {
	handler *MinipoolHandler
}

func (c *minipoolDissolveDetailsContext) Initialize() (types.ResponseStatus, error) {
	return types.ResponseStatus_Success, nil
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

func (c *minipoolDissolveDetailsContext) PrepareData(addresses []common.Address, mps []minipool.IMinipool, data *api.MinipoolDissolveDetailsData) (types.ResponseStatus, error) {
	details := make([]api.MinipoolDissolveDetails, len(mps))
	for i, mp := range mps {
		mpCommonDetails := mp.Common()
		status := mpCommonDetails.Status.Formatted()
		mpDetails := api.MinipoolDissolveDetails{
			Address:       mpCommonDetails.Address,
			InvalidStatus: !(status == rptypes.MinipoolStatus_Initialized || status == rptypes.MinipoolStatus_Prelaunch),
		}
		mpDetails.CanDissolve = !mpDetails.InvalidStatus
		details[i] = mpDetails
	}

	data.Details = details
	return types.ResponseStatus_Success, nil
}
