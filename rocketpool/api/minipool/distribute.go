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
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getDistributeBalanceDetailsForNode(c *cli.Context) (*api.GetMinipoolDistributeDetailsForNodeResponse, error) {
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
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, fmt.Errorf("error getting node account: %w", err)
	}

	// Response
	response := api.GetMinipoolDistributeDetailsForNodeResponse{}

	// Create the bindings
	node, err := node.NewNode(rp, nodeAccount.Address)
	if err != nil {
		return nil, fmt.Errorf("error creating node %s binding: %w", nodeAccount.Address.Hex(), err)
	}

	// Get contract state
	err = rp.Query(func(mc *batch.MultiCaller) error {
		node.GetMinipoolCount(mc)
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting contract state: %w", err)
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
		mpCommon.GetNodeAddress(mc)
		mpCommon.GetNodeRefundBalance(mc)
		mpCommon.GetFinalised(mc)
		mpCommon.GetStatus(mc)
		mpCommon.GetUserDepositBalance(mc)
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

	// Get the distribute details
	details := make([]api.MinipoolDistributeDetails, len(addresses))
	for i, mp := range mps {
		mpDetails, err := getMinipoolDistributeDetails(rp, mp, nodeAccount.Address, balances[i])
		if err != nil {
			return nil, fmt.Errorf("error checking closure details for minipool %s: %w", mp.GetMinipoolCommon().Details.Address.Hex(), err)
		}
		details[i] = mpDetails
	}

	// Get the node shares
	err = rp.BatchQuery(len(addresses), minipoolCompleteShareBatchSize, func(mc *batch.MultiCaller, i int) error {
		mpDetails := details[i]
		status := mpDetails.Status
		if status == types.Staking && mpDetails.CanDistribute {
			mps[i].GetMinipoolCommon().CalculateNodeShare(mc, &details[i].NodeShareOfDistributableBalance, details[i].DistributableBalance)
		}
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting node shares of minipool balances: %w", err)
	}

	// Update & return response
	response.Details = details
	return &response, nil
}

func getMinipoolDistributeDetails(rp *rocketpool.RocketPool, mp minipool.Minipool, nodeAddress common.Address, balance *big.Int) (api.MinipoolDistributeDetails, error) {
	mpCommonDetails := mp.GetMinipoolCommon().Details

	// Validate minipool owner
	if mpCommonDetails.NodeAddress != nodeAddress {
		return api.MinipoolDistributeDetails{}, fmt.Errorf("minipool %s does not belong to the node", mpCommonDetails.Address.Hex())
	}

	// Create the details with the balance / share info and status details
	var details api.MinipoolDistributeDetails
	details.Address = mpCommonDetails.Address
	details.Version = mpCommonDetails.Version
	details.Balance = balance
	details.Refund = mpCommonDetails.NodeRefundBalance
	details.IsFinalized = mpCommonDetails.IsFinalised
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
	case types.Staking:
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
	case types.Dissolved:
		// Dissolved but non-finalized / non-closed minipools can just have the whole balance sent back to the NO
		details.NodeShareOfDistributableBalance = details.Balance
	default:
		details.CanDistribute = false
		return details, nil
	}

	details.CanDistribute = true
	return details, nil
}

func distributeBalances(c *cli.Context, minipoolAddresses []common.Address) (*api.BatchTxResponse, error) {
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

	// Get the TXs
	txInfos := make([]*core.TransactionInfo, len(minipoolAddresses))
	for i, mp := range mps {
		mpCommon := mp.GetMinipoolCommon()
		minipoolAddress := mpCommon.Details.Address
		mpv3, success := minipool.GetMinipoolAsV3(mp)
		if !success {
			return nil, fmt.Errorf("minipool %s cannot be converted to v3 (current version: %d)", minipoolAddress.Hex(), mp.GetMinipoolCommon().Details.Version)
		}

		txInfo, err := mpv3.DistributeBalance(opts, true)
		if err != nil {
			return nil, fmt.Errorf("error simulating delegate upgrade transaction for minipool %s: %w", minipoolAddress.Hex(), err)
		}
		txInfos[i] = txInfo
	}

	response.TxInfos = txInfos
	return &response, nil
}
