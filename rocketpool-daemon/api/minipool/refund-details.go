package minipool

import (
	"errors"
	"math/big"
	"net/url"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"

	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
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
	server.RegisterMinipoolRoute[*minipoolRefundDetailsContext, api.MinipoolRefundDetailsData](
		router, "refund/details", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolRefundDetailsContext struct {
	handler *MinipoolHandler
}

func (c *minipoolRefundDetailsContext) Initialize() error {
	sp := c.handler.serviceProvider

	// Requirements
	err := errors.Join(
		sp.RequireNodeRegistered(),
	)
	if err != nil {
		return err
	}
	return nil
}

func (c *minipoolRefundDetailsContext) GetState(node *node.Node, mc *batch.MultiCaller) {
}

func (c *minipoolRefundDetailsContext) CheckState(node *node.Node, response *api.MinipoolRefundDetailsData) bool {
	return true
}

func (c *minipoolRefundDetailsContext) GetMinipoolDetails(mc *batch.MultiCaller, mp minipool.IMinipool, index int) {
	mpCommon := mp.Common()
	core.AddQueryablesToMulticall(mc,
		mpCommon.NodeAddress,
		mpCommon.NodeRefundBalance,
	)
}

func (c *minipoolRefundDetailsContext) PrepareData(addresses []common.Address, mps []minipool.IMinipool, data *api.MinipoolRefundDetailsData) error {
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
	return nil
}
