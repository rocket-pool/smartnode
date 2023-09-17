package minipool

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type minipoolRefundDetailsContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolRefundDetailsContextFactory) Create(vars map[string]string) (*minipoolRefundDetailsContext, error) {
	c := &minipoolRefundDetailsContext{
		handler: f.handler,
	}
	return c, nil
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

func (c *minipoolRefundDetailsContext) GetMinipoolDetails(mc *batch.MultiCaller, mp minipool.Minipool, index int) {
	mpCommon := mp.GetMinipoolCommon()
	mpCommon.GetNodeAddress(mc)
	mpCommon.GetNodeRefundBalance(mc)
}

func (c *minipoolRefundDetailsContext) PrepareData(addresses []common.Address, mps []minipool.Minipool, data *api.MinipoolRefundDetailsData) error {
	// Get the refund details
	details := make([]api.MinipoolRefundDetails, len(addresses))
	for i, mp := range mps {
		mpCommonDetails := mp.GetMinipoolCommon().Details
		mpDetails := api.MinipoolRefundDetails{
			Address:                   mpCommonDetails.Address,
			InsufficientRefundBalance: (mpCommonDetails.NodeRefundBalance.Cmp(big.NewInt(0)) == 0),
		}
		mpDetails.CanRefund = !mpDetails.InsufficientRefundBalance
		details[i] = mpDetails
	}

	data.Details = details
	return nil
}
