package minipool

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/types"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type minipoolExitDetailsContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolExitDetailsContextFactory) Create(vars map[string]string) (*minipoolExitDetailsContext, error) {
	c := &minipoolExitDetailsContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *minipoolExitDetailsContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterMinipoolRoute[*minipoolExitDetailsContext, api.MinipoolExitDetailsData](
		router, "exit/details", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolExitDetailsContext struct {
	handler *MinipoolHandler
}

func (c *minipoolExitDetailsContext) Initialize() error {
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

func (c *minipoolExitDetailsContext) GetState(node *node.Node, mc *batch.MultiCaller) {
}

func (c *minipoolExitDetailsContext) CheckState(node *node.Node, response *api.MinipoolExitDetailsData) bool {
	return true
}

func (c *minipoolExitDetailsContext) GetMinipoolDetails(mc *batch.MultiCaller, mp minipool.Minipool, index int) {
	mpCommon := mp.GetMinipoolCommon()
	mpCommon.GetNodeAddress(mc)
	mpCommon.GetStatus(mc)
}

func (c *minipoolExitDetailsContext) PrepareData(addresses []common.Address, mps []minipool.Minipool, response *api.MinipoolExitDetailsData) error {
	// Get the exit details
	details := make([]api.MinipoolExitDetails, len(addresses))
	for i, mp := range mps {
		mpCommonDetails := mp.GetMinipoolCommon().Details
		status := mpCommonDetails.Status.Formatted()
		mpDetails := api.MinipoolExitDetails{
			Address:       mpCommonDetails.Address,
			InvalidStatus: (status != types.Staking),
		}
		mpDetails.CanExit = !mpDetails.InvalidStatus
		details[i] = mpDetails
	}

	response.Details = details
	return nil
}
