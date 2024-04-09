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

type minipoolExitDetailsContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolExitDetailsContextFactory) Create(args url.Values) (*minipoolExitDetailsContext, error) {
	c := &minipoolExitDetailsContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *minipoolExitDetailsContextFactory) RegisterRoute(router *mux.Router) {
	RegisterMinipoolRoute[*minipoolExitDetailsContext, api.MinipoolExitDetailsData](
		router, "exit/details", f, f.handler.ctx, f.handler.logger, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolExitDetailsContext struct {
	handler *MinipoolHandler
}

func (c *minipoolExitDetailsContext) Initialize() (types.ResponseStatus, error) {
	return types.ResponseStatus_Success, nil
}

func (c *minipoolExitDetailsContext) GetState(node *node.Node, mc *batch.MultiCaller) {
}

func (c *minipoolExitDetailsContext) CheckState(node *node.Node, response *api.MinipoolExitDetailsData) bool {
	return true
}

func (c *minipoolExitDetailsContext) GetMinipoolDetails(mc *batch.MultiCaller, mp minipool.IMinipool, index int) {
	mpCommon := mp.Common()
	eth.AddQueryablesToMulticall(mc,
		mpCommon.NodeAddress,
		mpCommon.Status,
	)
}

func (c *minipoolExitDetailsContext) PrepareData(addresses []common.Address, mps []minipool.IMinipool, response *api.MinipoolExitDetailsData) (types.ResponseStatus, error) {
	// Get the exit details
	details := make([]api.MinipoolExitDetails, len(addresses))
	for i, mp := range mps {
		mpCommonDetails := mp.Common()
		status := mpCommonDetails.Status.Formatted()
		mpDetails := api.MinipoolExitDetails{
			Address:       mpCommonDetails.Address,
			InvalidStatus: (status != rptypes.MinipoolStatus_Staking),
		}
		mpDetails.CanExit = !mpDetails.InvalidStatus
		details[i] = mpDetails
	}

	response.Details = details
	return types.ResponseStatus_Success, nil
}
