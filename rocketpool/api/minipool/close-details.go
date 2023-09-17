package minipool

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/rocketpool/common/beacon"
	sharedtypes "github.com/rocket-pool/smartnode/shared/types"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type minipoolCloseDetailsContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolCloseDetailsContextFactory) Create(vars map[string]string) (*minipoolCloseDetailsContext, error) {
	c := &minipoolCloseDetailsContext{
		handler: f.handler,
	}
	return c, nil
}

// ===============
// === Context ===
// ===============

type minipoolCloseDetailsContext struct {
	handler *MinipoolHandler
	rp      *rocketpool.RocketPool
	bc      beacon.Client
}

func (c *minipoolCloseDetailsContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.bc = sp.GetBeaconClient()
	return nil
}

func (c *minipoolCloseDetailsContext) GetState(node *node.Node, mc *batch.MultiCaller) {
	node.GetFeeDistributorInitialized(mc)
}

func (c *minipoolCloseDetailsContext) CheckState(node *node.Node, response *api.MinipoolCloseDetailsData) bool {
	response.IsFeeDistributorInitialized = node.Details.IsFeeDistributorInitialized
	return response.IsFeeDistributorInitialized
}

func (c *minipoolCloseDetailsContext) GetMinipoolDetails(mc *batch.MultiCaller, mp minipool.Minipool, index int) {
	mpCommon := mp.GetMinipoolCommon()
	mpCommon.GetNodeAddress(mc)
	mpCommon.GetNodeRefundBalance(mc)
	mpCommon.GetFinalised(mc)
	mpCommon.GetStatus(mc)
	mpCommon.GetUserDepositBalance(mc)
	mpCommon.GetPubkey(mc)
}

func (c *minipoolCloseDetailsContext) PrepareResponse(addresses []common.Address, mps []minipool.Minipool, data *api.MinipoolCloseDetailsData) error {
	// Get the current ETH balances of each minipool
	balances, err := c.rp.BalanceBatcher.GetEthBalances(addresses, nil)
	if err != nil {
		return fmt.Errorf("error getting minipool balances: %w", err)
	}

	// Get the closure details
	details := make([]api.MinipoolCloseDetails, len(addresses))
	for i, mp := range mps {
		mpDetails, err := getMinipoolCloseDetails(c.rp, mp, balances[i])
		if err != nil {
			return fmt.Errorf("error checking closure details for minipool %s: %w", mp.GetMinipoolCommon().Details.Address.Hex(), err)
		}
		details[i] = mpDetails
	}

	// Get the node shares
	err = c.rp.BatchQuery(len(addresses), minipoolCompleteShareBatchSize, func(mc *batch.MultiCaller, i int) error {
		mpv3, success := minipool.GetMinipoolAsV3(mps[i])
		if success {
			details[i].Distributed = mpv3.Details.HasUserDistributed
			mpv3.CalculateNodeShare(mc, &details[i].NodeShareOfEffectiveBalance, details[i].EffectiveBalance)
		}
		return nil
	}, nil)
	if err != nil {
		return fmt.Errorf("error getting node shares of minipool balances: %w", err)
	}

	// Get the beacon statuses for each closeable minipool
	pubkeys := []types.ValidatorPubkey{}
	pubkeyMap := map[common.Address]types.ValidatorPubkey{}
	for i, mp := range details {
		if mp.Status == types.Dissolved {
			// Ignore dissolved minipools
			continue
		}
		pubkey := mps[i].GetMinipoolCommon().Details.Pubkey
		pubkeyMap[mp.Address] = pubkey
		pubkeys = append(pubkeys, pubkey)
	}
	statusMap, err := c.bc.GetValidatorStatuses(pubkeys, nil)
	if err != nil {
		return fmt.Errorf("error getting beacon status of minipools: %w", err)
	}

	// Review closeability based on validator status
	for i, mp := range details {
		pubkey := pubkeyMap[mp.Address]
		validator := statusMap[pubkey]
		if mp.Status != types.Dissolved {
			details[i].BeaconState = validator.Status
			if validator.Status != sharedtypes.ValidatorState_WithdrawalDone {
				details[i].CanClose = false
			}
		}
	}

	data.Details = details
	return nil
}

func getMinipoolCloseDetails(rp *rocketpool.RocketPool, mp minipool.Minipool, balance *big.Int) (api.MinipoolCloseDetails, error) {
	mpCommonDetails := mp.GetMinipoolCommon().Details

	// Create the details with the balance / share info and status details
	var details api.MinipoolCloseDetails
	details.Address = mpCommonDetails.Address
	details.Version = mpCommonDetails.Version
	details.Balance = balance
	details.Refund = mpCommonDetails.NodeRefundBalance
	details.IsFinalized = mpCommonDetails.IsFinalised
	details.Status = mpCommonDetails.Status.Formatted()
	details.UserDepositBalance = mpCommonDetails.UserDepositBalance
	details.NodeShareOfEffectiveBalance = big.NewInt(0)

	// Ignore minipools that are too old
	if details.Version < 3 {
		details.CanClose = false
		return details, nil
	}

	// Can't close a minipool that's already finalized
	if details.IsFinalized {
		details.CanClose = false
		return details, nil
	}

	// Make sure it's in a closeable state
	details.EffectiveBalance = big.NewInt(0).Sub(details.Balance, details.Refund)
	switch details.Status {
	case types.Dissolved:
		details.CanClose = true

	case types.Staking, types.Withdrawable:
		// Ignore minipools with a balance lower than the refund
		if details.Balance.Cmp(details.Refund) == -1 {
			details.CanClose = false
			return details, nil
		}

		// Ignore minipools with an effective balance lower than v3 rewards-vs-exit cap
		eight := eth.EthToWei(8)
		if details.EffectiveBalance.Cmp(eight) == -1 {
			details.CanClose = false
			return details, nil
		}

		details.CanClose = true

	case types.Initialized, types.Prelaunch:
		details.CanClose = false
		return details, nil
	}

	return details, nil
}
