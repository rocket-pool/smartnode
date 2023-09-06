package minipool

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getDistributeBalanceDetailsForNode(c *cli.Context) (*api.GetMinipoolDistributeDetailsForNodeResponse, error) {
	return createMinipoolQuery(c,
		nil,
		nil,
		nil,
		func(mc *batch.MultiCaller, mp minipool.Minipool) {
			mpCommon := mp.GetMinipoolCommon()
			mpCommon.GetNodeAddress(mc)
			mpCommon.GetNodeRefundBalance(mc)
			mpCommon.GetFinalised(mc)
			mpCommon.GetStatus(mc)
			mpCommon.GetUserDepositBalance(mc)
		},
		func(rp *rocketpool.RocketPool, nodeAddress common.Address, addresses []common.Address, mps []minipool.Minipool, response *api.GetMinipoolDistributeDetailsForNodeResponse) error {
			// Get the current ETH balances of each minipool
			balances, err := rp.BalanceBatcher.GetEthBalances(addresses, nil)
			if err != nil {
				return fmt.Errorf("error getting minipool balances: %w", err)
			}

			// Get the distribute details
			details := make([]api.MinipoolDistributeDetails, len(addresses))
			for i, mp := range mps {
				mpDetails, err := getMinipoolDistributeDetails(rp, mp, nodeAddress, balances[i])
				if err != nil {
					return fmt.Errorf("error checking closure details for minipool %s: %w", mp.GetMinipoolCommon().Details.Address.Hex(), err)
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
				return fmt.Errorf("error getting node shares of minipool balances: %w", err)
			}

			// Update & return response
			response.Details = details
			return nil
		},
	)
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
	return createBatchTxResponseForV3(c, minipoolAddresses, func(mpv3 *minipool.MinipoolV3, opts *bind.TransactOpts) (*core.TransactionInfo, error) {
		return mpv3.DistributeBalance(opts, true)
	}, "distribute-balance")
}
