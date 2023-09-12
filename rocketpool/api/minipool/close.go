package minipool

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

type minipoolCloseManager struct {
}

func (m *minipoolCloseManager) CreateBindings(rp *rocketpool.RocketPool) error {
	return nil
}

func (m *minipoolCloseManager) GetState(node *node.Node, mc *batch.MultiCaller) {
	node.GetFeeDistributorInitialized(mc)
}

func (m *minipoolCloseManager) CheckState(node *node.Node, response *api.MinipoolCloseDetailsResponse) bool {
	response.IsFeeDistributorInitialized = node.Details.IsFeeDistributorInitialized
	return response.IsFeeDistributorInitialized
}

func (m *minipoolCloseManager) GetMinipoolDetails(mc *batch.MultiCaller, mp minipool.Minipool, index int) {
	mpCommon := mp.GetMinipoolCommon()
	mpCommon.GetNodeAddress(mc)
	mpCommon.GetNodeRefundBalance(mc)
	mpCommon.GetFinalised(mc)
	mpCommon.GetStatus(mc)
	mpCommon.GetUserDepositBalance(mc)
	mpCommon.GetPubkey(mc)
}

func (m *minipoolCloseManager) PrepareResponse(rp *rocketpool.RocketPool, bc beacon.Client, addresses []common.Address, mps []minipool.Minipool, response *api.MinipoolCloseDetailsResponse) error {
	// Get the current ETH balances of each minipool
	balances, err := rp.BalanceBatcher.GetEthBalances(addresses, nil)
	if err != nil {
		return fmt.Errorf("error getting minipool balances: %w", err)
	}

	// Get the closure details
	details := make([]api.MinipoolCloseDetails, len(addresses))
	for i, mp := range mps {
		mpDetails, err := getMinipoolCloseDetails(rp, mp, balances[i])
		if err != nil {
			return fmt.Errorf("error checking closure details for minipool %s: %w", mp.GetMinipoolCommon().Details.Address.Hex(), err)
		}
		details[i] = mpDetails
	}

	// Get the node shares
	err = rp.BatchQuery(len(addresses), minipoolCompleteShareBatchSize, func(mc *batch.MultiCaller, i int) error {
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
	statusMap, err := bc.GetValidatorStatuses(pubkeys, nil)
	if err != nil {
		return fmt.Errorf("error getting beacon status of minipools: %w", err)
	}

	// Review closeability based on validator status
	for i, mp := range details {
		pubkey := pubkeyMap[mp.Address]
		validator := statusMap[pubkey]
		if mp.Status != types.Dissolved {
			details[i].BeaconState = validator.Status
			if validator.Status != beacon.ValidatorState_WithdrawalDone {
				details[i].CanClose = false
			}
		}
	}

	response.Details = details
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

func closeMinipools(c *cli.Context, minipoolAddresses []common.Address) (*api.BatchTxResponse, error) {
	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.BatchTxResponse{}

	// Create minipools
	mps, err := minipool.CreateMinipoolsFromAddresses(rp, minipoolAddresses, false, nil)
	if err != nil {
		return nil, err
	}

	// Run the details getter
	err = rp.BatchQuery(len(minipoolAddresses), minipoolBatchSize, func(mc *batch.MultiCaller, i int) error {
		mps[i].GetMinipoolCommon().GetStatus(mc)
		mpv3, isMpv3 := minipool.GetMinipoolAsV3(mps[i])
		if isMpv3 {
			mpv3.GetUserDistributed(mc)
		}
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool details: %w", err)
	}

	// Get the TXs
	txInfos := make([]*core.TransactionInfo, len(minipoolAddresses))
	for i, mp := range mps {
		mpCommon := mp.GetMinipoolCommon()
		minipoolAddress := mpCommon.Details.Address
		mpv3, isMpv3 := minipool.GetMinipoolAsV3(mp)

		// If it's dissolved, just close it
		if mpCommon.Details.Status.Formatted() == types.Dissolved {
			// Get gas estimate
			txInfo, err := mp.GetMinipoolCommon().Close(opts)
			if err != nil {
				return nil, fmt.Errorf("error simulating close for minipool %s: %w", minipoolAddress.Hex(), err)
			}
			txInfos[i] = txInfo
		} else {
			// Check if it's an upgraded Atlas-era minipool
			if isMpv3 {
				if mpv3.Details.HasUserDistributed {
					// It's already been distributed so just finalize it
					txInfo, err := mpv3.Finalise(opts)
					if err != nil {
						return nil, fmt.Errorf("error simulating finalise for minipool %s: %w", minipoolAddress.Hex(), err)
					}
					txInfos[i] = txInfo
				} else {
					// Do a distribution, which will finalize it
					txInfo, err := mpv3.DistributeBalance(opts, false)
					if err != nil {
						return nil, fmt.Errorf("error simulation distribute balance for minipool %s: %w", minipoolAddress.Hex(), err)
					}
					txInfos[i] = txInfo
				}
			} else {
				return nil, fmt.Errorf("cannot create v3 binding for minipool %s, version %d", minipoolAddress.Hex(), mpCommon.Details.Version)
			}
		}
	}

	response.TxInfos = txInfos
	return &response, nil
}
