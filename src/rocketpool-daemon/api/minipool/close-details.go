package minipool

import (
	"fmt"
	"math/big"
	"net/url"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/minipool"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/v2/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type minipoolCloseDetailsContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolCloseDetailsContextFactory) Create(args url.Values) (*minipoolCloseDetailsContext, error) {
	c := &minipoolCloseDetailsContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *minipoolCloseDetailsContextFactory) RegisterRoute(router *mux.Router) {
	RegisterMinipoolRoute[*minipoolCloseDetailsContext, api.MinipoolCloseDetailsData](
		router, "close/details", f, f.handler.ctx, f.handler.logger, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolCloseDetailsContext struct {
	handler *MinipoolHandler
	rp      *rocketpool.RocketPool
	bc      beacon.IBeaconClient
}

func (c *minipoolCloseDetailsContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.bc = sp.GetBeaconClient()
	return types.ResponseStatus_Success, nil
}

func (c *minipoolCloseDetailsContext) GetState(node *node.Node, mc *batch.MultiCaller) {
	node.IsFeeDistributorInitialized.AddToQuery(mc)
}

func (c *minipoolCloseDetailsContext) CheckState(node *node.Node, response *api.MinipoolCloseDetailsData) bool {
	response.IsFeeDistributorInitialized = node.IsFeeDistributorInitialized.Get()
	return response.IsFeeDistributorInitialized
}

func (c *minipoolCloseDetailsContext) GetMinipoolDetails(mc *batch.MultiCaller, mp minipool.IMinipool, index int) {
	mpCommon := mp.Common()
	eth.AddQueryablesToMulticall(mc,
		mpCommon.NodeAddress,
		mpCommon.NodeRefundBalance,
		mpCommon.IsFinalised,
		mpCommon.Status,
		mpCommon.UserDepositBalance,
		mpCommon.Pubkey,
	)
	mpv3, success := minipool.GetMinipoolAsV3(mp)
	if success {
		mpv3.HasUserDistributed.AddToQuery(mc)
	}
}

func (c *minipoolCloseDetailsContext) PrepareData(addresses []common.Address, mps []minipool.IMinipool, data *api.MinipoolCloseDetailsData) (types.ResponseStatus, error) {
	ctx := c.handler.ctx
	// Get the current ETH balances of each minipool
	balances, err := c.rp.BalanceBatcher.GetEthBalances(addresses, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting minipool balances: %w", err)
	}

	// Get the closure details
	details := make([]api.MinipoolCloseDetails, len(addresses))
	for i, mp := range mps {
		details[i] = getMinipoolCloseDetails(c.rp, mp, balances[i])
	}

	// Get the node shares
	err = c.rp.BatchQuery(len(addresses), minipoolCompleteShareBatchSize, func(mc *batch.MultiCaller, i int) error {
		mpv3, success := minipool.GetMinipoolAsV3(mps[i])
		if success {
			details[i].Distributed = mpv3.HasUserDistributed.Get()
			mpv3.CalculateNodeShare(mc, &details[i].NodeShareOfEffectiveBalance, details[i].EffectiveBalance)
		}
		return nil
	}, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting node shares of minipool balances: %w", err)
	}

	// Get the beacon statuses for each closeable minipool
	pubkeys := []beacon.ValidatorPubkey{}
	pubkeyMap := map[common.Address]beacon.ValidatorPubkey{}
	for i, mp := range details {
		if mp.Status == rptypes.MinipoolStatus_Dissolved {
			// Ignore dissolved minipools
			continue
		}
		pubkey := mps[i].Common().Pubkey.Get()
		pubkeyMap[mp.Address] = pubkey
		pubkeys = append(pubkeys, pubkey)
	}
	statusMap, err := c.bc.GetValidatorStatuses(ctx, pubkeys, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting beacon status of minipools: %w", err)
	}

	// Review closeability based on validator status
	for i, mp := range details {
		pubkey := pubkeyMap[mp.Address]
		validator := statusMap[pubkey]
		if mp.Status != rptypes.MinipoolStatus_Dissolved {
			details[i].BeaconState = validator.Status
			if validator.Status != beacon.ValidatorState_WithdrawalDone {
				details[i].CanClose = false
			}
		}
	}

	data.Details = details
	return types.ResponseStatus_Success, nil
}

func getMinipoolCloseDetails(rp *rocketpool.RocketPool, mp minipool.IMinipool, balance *big.Int) api.MinipoolCloseDetails {
	mpCommonDetails := mp.Common()

	// Create the details with the balance / share info and status details
	var details api.MinipoolCloseDetails
	details.Address = mpCommonDetails.Address
	details.Version = mpCommonDetails.Version
	details.Balance = balance
	details.Refund = mpCommonDetails.NodeRefundBalance.Get()
	details.IsFinalized = mpCommonDetails.IsFinalised.Get()
	details.Status = mpCommonDetails.Status.Formatted()
	details.UserDepositBalance = mpCommonDetails.UserDepositBalance.Get()
	details.NodeShareOfEffectiveBalance = big.NewInt(0)

	// Ignore minipools that are too old
	if details.Version < 3 {
		details.CanClose = false
		return details
	}

	// Can't close a minipool that's already finalized
	if details.IsFinalized {
		details.CanClose = false
		return details
	}

	// Make sure it's in a closeable state
	details.EffectiveBalance = big.NewInt(0).Sub(details.Balance, details.Refund)
	switch details.Status {
	case rptypes.MinipoolStatus_Dissolved:
		details.CanClose = true

	case rptypes.MinipoolStatus_Staking, rptypes.MinipoolStatus_Withdrawable:
		// Ignore minipools with a balance lower than the refund
		if details.Balance.Cmp(details.Refund) == -1 {
			details.CanClose = false
			return details
		}

		// Ignore minipools with an effective balance lower than v3 rewards-vs-exit cap
		eight := eth.EthToWei(8)
		if details.EffectiveBalance.Cmp(eight) == -1 {
			details.CanClose = false
			return details
		}

		details.CanClose = true

	case rptypes.MinipoolStatus_Initialized, rptypes.MinipoolStatus_Prelaunch:
		details.CanClose = false
		return details
	}

	return details
}
