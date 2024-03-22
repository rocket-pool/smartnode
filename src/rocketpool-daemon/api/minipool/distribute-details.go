package minipool

import (
	"errors"
	"fmt"
	"math/big"
	"net/url"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type minipoolDistributeDetailsContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolDistributeDetailsContextFactory) Create(args url.Values) (*minipoolDistributeDetailsContext, error) {
	c := &minipoolDistributeDetailsContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *minipoolDistributeDetailsContextFactory) RegisterRoute(router *mux.Router) {
	RegisterMinipoolRoute[*minipoolDistributeDetailsContext, api.MinipoolDistributeDetailsData](
		router, "distribute/details", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolDistributeDetailsContext struct {
	handler *MinipoolHandler
	rp      *rocketpool.RocketPool
}

func (c *minipoolDistributeDetailsContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()

	// Requirements
	return errors.Join(
		sp.RequireNodeRegistered(),
	)
}

func (c *minipoolDistributeDetailsContext) GetState(node *node.Node, mc *batch.MultiCaller) {
}

func (c *minipoolDistributeDetailsContext) CheckState(node *node.Node, response *api.MinipoolDistributeDetailsData) bool {
	return true
}

func (c *minipoolDistributeDetailsContext) GetMinipoolDetails(mc *batch.MultiCaller, mp minipool.IMinipool, index int) {
	mpCommon := mp.Common()
	eth.AddQueryablesToMulticall(mc,
		mpCommon.NodeAddress,
		mpCommon.NodeRefundBalance,
		mpCommon.IsFinalised,
		mpCommon.Status,
		mpCommon.UserDepositBalance,
	)
}

func (c *minipoolDistributeDetailsContext) PrepareData(addresses []common.Address, mps []minipool.IMinipool, data *api.MinipoolDistributeDetailsData) error {
	// Get the current ETH balances of each minipool
	balances, err := c.rp.BalanceBatcher.GetEthBalances(addresses, nil)
	if err != nil {
		return fmt.Errorf("error getting minipool balances: %w", err)
	}

	// Get the distribute details
	details := make([]api.MinipoolDistributeDetails, len(addresses))
	for i, mp := range mps {
		mpDetails, err := getMinipoolDistributeDetails(c.rp, mp, balances[i])
		if err != nil {
			return fmt.Errorf("error checking closure details for minipool %s: %w", mp.Common().Address.Hex(), err)
		}
		details[i] = mpDetails
	}

	// Get the node shares
	err = c.rp.BatchQuery(len(addresses), minipoolCompleteShareBatchSize, func(mc *batch.MultiCaller, i int) error {
		mpDetails := details[i]
		status := mpDetails.Status
		if status == types.MinipoolStatus_Staking && mpDetails.CanDistribute {
			mps[i].Common().CalculateNodeShare(mc, &details[i].NodeShareOfDistributableBalance, details[i].DistributableBalance)
		}
		return nil
	}, nil)
	if err != nil {
		return fmt.Errorf("error getting node shares of minipool balances: %w", err)
	}

	// Update & return response
	data.Details = details
	return nil
}

func getMinipoolDistributeDetails(rp *rocketpool.RocketPool, mp minipool.IMinipool, balance *big.Int) (api.MinipoolDistributeDetails, error) {
	mpCommonDetails := mp.Common()

	// Create the details with the balance / share info and status details
	var details api.MinipoolDistributeDetails
	details.Address = mpCommonDetails.Address
	details.Version = mpCommonDetails.Version
	details.Balance = balance
	details.Refund = mpCommonDetails.NodeRefundBalance.Get()
	details.IsFinalized = mpCommonDetails.IsFinalised.Get()
	details.Status = mpCommonDetails.Status.Formatted()
	details.NodeShareOfDistributableBalance = big.NewInt(0)

	// Ignore minipools that are too old
	if details.Version < 3 {
		details.CanDistribute = false
		return details, nil
	}

	// Can't distribute a minipool that's already finalized
	if details.IsFinalized {
		details.CanDistribute = false
		return details, nil
	}

	// Ignore minipools with 0 balance
	if details.Balance.Cmp(zero()) == 0 {
		details.CanDistribute = false
		return details, nil
	}

	// Make sure it's in a distributable state
	switch details.Status {
	case types.MinipoolStatus_Staking:
		// Ignore minipools with a balance lower than the refund
		if details.Balance.Cmp(details.Refund) == -1 {
			details.CanDistribute = false
			return details, nil
		}

		// Ignore minipools with an effective balance higher than v3 rewards-vs-exit cap
		details.DistributableBalance = big.NewInt(0).Sub(details.Balance, details.Refund)
		eight := eth.EthToWei(8)
		if details.DistributableBalance.Cmp(eight) >= 0 {
			details.CanDistribute = false
			return details, nil
		}
	case types.MinipoolStatus_Dissolved:
		// Dissolved but non-finalized / non-closed minipools can just have the whole balance sent back to the NO
		details.NodeShareOfDistributableBalance = details.Balance
	default:
		details.CanDistribute = false
		return details, nil
	}

	details.CanDistribute = true
	return details, nil
}
