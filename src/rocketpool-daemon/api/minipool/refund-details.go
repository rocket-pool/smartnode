package minipool

import (
	"math/big"
	"net/url"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/minipool"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type minipoolRefundDetailsContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolRefundDetailsContextFactory) Create(args url.Values) (*minipoolRefundDetailsContext, error) {
	c := &minipoolRefundDetailsContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *minipoolRefundDetailsContextFactory) RegisterRoute(router *mux.Router) {
	RegisterMinipoolRoute[*minipoolRefundDetailsContext, api.MinipoolRefundDetailsData](
		router, "refund/details", f, f.handler.ctx, f.handler.logger, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolRefundDetailsContext struct {
	handler *MinipoolHandler
}

func (c *minipoolRefundDetailsContext) Initialize() (types.ResponseStatus, error) {
	return types.ResponseStatus_Success, nil
}

func (c *minipoolRefundDetailsContext) GetState(node *node.Node, mc *batch.MultiCaller) {
}

func (c *minipoolRefundDetailsContext) CheckState(node *node.Node, response *api.MinipoolRefundDetailsData) bool {
	return true
}

func (c *minipoolRefundDetailsContext) GetMinipoolDetails(mc *batch.MultiCaller, mp minipool.IMinipool, index int) {
	mpCommon := mp.Common()
	eth.AddQueryablesToMulticall(mc,
		mpCommon.NodeAddress,
		mpCommon.NodeRefundBalance,
	)
}

func (c *minipoolRefundDetailsContext) PrepareData(addresses []common.Address, mps []minipool.IMinipool, data *api.MinipoolRefundDetailsData) (types.ResponseStatus, error) {
	// Get the refund details
	details := make([]api.MinipoolRefundDetails, len(addresses))
	for i, mp := range mps {
		mpCommonDetails := mp.Common()
		mpDetails := api.MinipoolRefundDetails{
			Address:                   mpCommonDetails.Address,
			InsufficientRefundBalance: (mpCommonDetails.NodeRefundBalance.Get().Cmp(big.NewInt(0)) == 0),
		}
		mpDetails.CanRefund = !mpDetails.InsufficientRefundBalance
		details[i] = mpDetails
	}

	data.Details = details
	return types.ResponseStatus_Success, nil
}
