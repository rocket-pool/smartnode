package minipool

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
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

func getMinipoolCloseDetailsForNode(c *cli.Context) (*api.GetMinipoolCloseDetailsForNodeResponse, error) {
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
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, fmt.Errorf("error getting node account: %w", err)
	}

	// Response
	response := api.GetMinipoolCloseDetailsForNodeResponse{}

	// Create the bindings
	node, err := node.NewNode(rp, nodeAccount.Address)
	if err != nil {
		return nil, fmt.Errorf("error creating node %s binding: %w", nodeAccount.Address.Hex(), err)
	}

	// Get contract state
	err = rp.Query(func(mc *batch.MultiCaller) error {
		node.GetFeeDistributorInitialized(mc)
		node.GetMinipoolCount(mc)
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting contract state: %w", err)
	}

	// Check the fee distributor
	response.IsFeeDistributorInitialized = node.Details.IsFeeDistributorInitialized
	if !response.IsFeeDistributorInitialized {
		return &response, nil
	}

	// Get the minipool addresses for this node
	addresses, err := node.GetMinipoolAddresses(node.Details.MinipoolCount.Formatted(), nil)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool addresses: %w", err)
	}

	// Create each minipool binding
	mps, err := minipool.CreateMinipoolsFromAddresses(rp, addresses, false, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating minipool bindings: %w", err)
	}

	// Get the relevant details
	err = rp.BatchQuery(len(addresses), minipoolBatchSize, func(mc *batch.MultiCaller, i int) error {
		mpCommon := mps[i].GetMinipoolCommon()
		mpCommon.GetNodeRefundBalance(mc)
		mpCommon.GetFinalised(mc)
		mpCommon.GetStatus(mc)
		mpCommon.GetUserDepositBalance(mc)
		mpCommon.GetPubkey(mc)
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool details: %w", err)
	}

	// Get the current ETH balances of each minipool
	balances, err := rp.BalanceBatcher.GetEthBalances(addresses, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool balances: %w", err)
	}

	// Get the transaction opts
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Get the closure details
	details := make([]api.MinipoolCloseDetails, len(addresses))
	for i, mp := range mps {
		mpDetails, err := getMinipoolCloseDetails(rp, mp, nodeAccount.Address, balances[i], opts)
		if err != nil {
			return nil, fmt.Errorf("error checking closure details for minipool %s: %w", mp.GetMinipoolCommon().Details.Address.Hex(), err)
		}
		details[i] = mpDetails
	}

	// Get the node shares
	err = rp.BatchQuery(len(addresses), minipoolCompleteShareBatchSize, func(mc *batch.MultiCaller, i int) error {
		mpv3, success := minipool.GetMinipoolAsV3(mps[i])
		if success {
			mpv3.CalculateNodeShare(mc, &details[i].NodeShare, details[i].EffectiveBalance)
		}
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting node shares of minipool balances: %w", err)
	}

	// Get the beacon statuses for each closeable minipool
	pubkeys := []types.ValidatorPubkey{}
	pubkeyMap := map[common.Address]types.ValidatorPubkey{}
	for i, mp := range details {
		if mp.MinipoolStatus == types.Dissolved {
			// Ignore dissolved minipools
			continue
		}
		pubkey := mps[i].GetMinipoolCommon().Details.Pubkey
		pubkeyMap[mp.Address] = pubkey
		pubkeys = append(pubkeys, pubkey)
	}
	statusMap, err := bc.GetValidatorStatuses(pubkeys, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting beacon status of minipools: %w", err)
	}

	// Review closeability based on validator status
	for i, mp := range details {
		pubkey := pubkeyMap[mp.Address]
		validator := statusMap[pubkey]
		if mp.MinipoolStatus != types.Dissolved {
			details[i].BeaconState = validator.Status
			if validator.Status != beacon.ValidatorState_WithdrawalDone {
				details[i].CanClose = false
			}
		}
	}

	response.Details = details
	return &response, nil
}

func closeMinipool(c *cli.Context, minipoolAddress common.Address, checkValidity bool) (*api.CloseMinipoolResponse, error) {
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
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, fmt.Errorf("error getting node account: %w", err)
	}

	// Response
	response := api.CloseMinipoolResponse{}

	// Create minipool
	mp, err := minipool.CreateMinipoolFromAddress(rp, minipoolAddress, false, nil)
	if err != nil {
		return nil, err
	}
	mpCommon := mp.GetMinipoolCommon()

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Check validity if requested
	if checkValidity {
		// Get the relevant details
		err = rp.Query(func(mc *batch.MultiCaller) error {
			mpCommon.GetNodeRefundBalance(mc)
			mpCommon.GetFinalised(mc)
			mpCommon.GetStatus(mc)
			mpCommon.GetUserDepositBalance(mc)
			mpCommon.GetPubkey(mc)
			return nil
		}, nil)
		if err != nil {
			return nil, fmt.Errorf("error getting minipool details: %w", err)
		}

		// Get the balance
		balance, err := rp.Client.BalanceAt(context.Background(), minipoolAddress, nil)
		if err != nil {
			return nil, fmt.Errorf("error getting minipool balance: %w", err)
		}

		// Check the EL details
		mpDetails, err := getMinipoolCloseDetails(rp, mp, nodeAccount.Address, balance, opts)
		if err != nil {
			return nil, fmt.Errorf("error checking closure details for minipool: %w", mp.GetMinipoolCommon().Details.Address.Hex(), err)
		}
		if !mpDetails.CanClose {
			return nil, fmt.Errorf("minipool cannot be closed in its current state")
		}

		// Get the CL details
		pubkey := mpCommon.Details.Pubkey
		statusMap, err := bc.GetValidatorStatuses([]types.ValidatorPubkey{pubkey}, nil)
		if err != nil {
			return nil, fmt.Errorf("error getting beacon status of minipool: %w", err)
		}

		// Review closeability based on validator status
		validator := statusMap[pubkey]
		if mpDetails.MinipoolStatus != types.Dissolved {
			mpDetails.BeaconState = validator.Status
			if validator.Status != beacon.ValidatorState_WithdrawalDone {
				mpDetails.CanClose = false
			}
		}
		if !mpDetails.CanClose {
			return nil, fmt.Errorf("minipool cannot be closed in its current state")
		}
	}

	// Close
	txInfo, err := mp.GetMinipoolCommon().Close(opts)
	if err != nil {
		return nil, err
	}
	response.TxInfo = txInfo

	return &response, nil
}

func getMinipoolCloseDetails(rp *rocketpool.RocketPool, mp minipool.Minipool, nodeAddress common.Address, balance *big.Int, opts *bind.TransactOpts) (api.MinipoolCloseDetails, error) {
	mpCommonDetails := mp.GetMinipoolCommon().Details

	// Validate minipool owner
	if mpCommonDetails.NodeAddress != nodeAddress {
		return api.MinipoolCloseDetails{}, fmt.Errorf("minipool %s does not belong to the node", mpCommonDetails.Address.Hex())
	}

	// Create the details with the balance / share info and status details
	var details api.MinipoolCloseDetails
	details.Address = mpCommonDetails.Address
	details.MinipoolVersion = mpCommonDetails.Version
	details.Balance = balance
	details.Refund = mpCommonDetails.NodeRefundBalance
	details.IsFinalized = mpCommonDetails.IsFinalised
	details.MinipoolStatus = mpCommonDetails.Status.Formatted()
	details.UserDepositBalance = mpCommonDetails.UserDepositBalance
	details.NodeShare = big.NewInt(0)

	// Ignore minipools that are too old
	if details.MinipoolVersion < 3 {
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
	switch details.MinipoolStatus {
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

	// If it's dissolved, just close it
	if details.MinipoolStatus == types.Dissolved {
		// Get gas estimate
		txInfo, err := mp.GetMinipoolCommon().Close(opts)
		if err != nil {
			return api.MinipoolCloseDetails{}, fmt.Errorf("error simulating close for MP %s: %w", mpCommonDetails.Address.Hex(), err)
		}
		details.TxInfo = txInfo
	} else {
		// Check if it's an upgraded Atlas-era minipool
		mpv3, success := minipool.GetMinipoolAsV3(mp)
		if success {
			details.Distributed = mpv3.Details.HasUserDistributed
			if details.Distributed {
				// It's already been distributed so just finalize it
				txInfo, err := mpv3.Finalise(opts)
				if err != nil {
					return api.MinipoolCloseDetails{}, fmt.Errorf("error simulating finalise for MP %s: %w", mpCommonDetails.Address.Hex(), err)
				}
				details.TxInfo = txInfo
			} else {
				// Do a distribution, which will finalize it
				txInfo, err := mpv3.DistributeBalance(opts, false)
				if err != nil {
					return api.MinipoolCloseDetails{}, fmt.Errorf("error simulation distribute balance for MP %s: %w", mpCommonDetails.Address.Hex(), err)
				}
				details.TxInfo = txInfo
			}
		} else {
			return api.MinipoolCloseDetails{}, fmt.Errorf("cannot create v3 binding for minipool %s, version %d", mpCommonDetails.Address.Hex(), mpCommonDetails.Version)
		}
	}

	return details, nil
}
