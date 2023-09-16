package network

import (
	"context"
	"fmt"
	"math/big"

	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/deposit"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/tokens"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getStats(c *cli.Context) (*api.NetworkStatsData, error) {
	// Get services
	if err := services.RequireEthClientSynced(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.NetworkStatsData{}

	// Create bindings
	depositPool, err := deposit.NewDepositPool(rp)
	if err != nil {
		return nil, fmt.Errorf("error getting deposit pool binding: %w", err)
	}
	mpQueue, err := minipool.NewMinipoolQueue(rp)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool queue binding: %w", err)
	}
	networkBalances, err := network.NewNetworkBalances(rp)
	if err != nil {
		return nil, fmt.Errorf("error getting network balances binding: %w", err)
	}
	networkPrices, err := network.NewNetworkPrices(rp)
	if err != nil {
		return nil, fmt.Errorf("error getting network prices binding: %w", err)
	}
	networkFees, err := network.NewNetworkFees(rp)
	if err != nil {
		return nil, fmt.Errorf("error getting network fees binding: %w", err)
	}
	nodeMgr, err := node.NewNodeManager(rp)
	if err != nil {
		return nil, fmt.Errorf("error getting node manager binding: %w", err)
	}
	nodeStaking, err := node.NewNodeStaking(rp)
	if err != nil {
		return nil, fmt.Errorf("error getting node staking binding: %w", err)
	}
	mpMgr, err := minipool.NewMinipoolManager(rp)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool manager binding: %w", err)
	}
	reth, err := tokens.NewTokenReth(rp)
	if err != nil {
		return nil, fmt.Errorf("error getting rETH token binding: %w", err)
	}
	sp, err := rp.GetContract(rocketpool.ContractName_RocketSmoothingPool)
	if err != nil {
		return nil, fmt.Errorf("error getting rETH token binding: %w", err)
	}

	// Get contract state
	err = rp.Query(func(mc *batch.MultiCaller) error {
		depositPool.GetBalance(mc)
		mpQueue.GetTotalCapacity(mc)
		networkBalances.GetEthUtilizationRate(mc)
		networkFees.GetNodeFee(mc)
		nodeMgr.GetNodeCount(mc)
		mpMgr.GetMinipoolCount(mc)
		mpMgr.GetFinalisedMinipoolCount(mc)
		networkPrices.GetRplPrice(mc)
		nodeStaking.GetTotalRPLStake(mc)
		reth.GetExchangeRate(mc)
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting contract state: %w", err)
	}

	// Handle the details
	response.DepositPoolBalance = depositPool.Details.Balance
	response.MinipoolCapacity = mpQueue.Details.TotalCapacity
	response.StakerUtilization = networkBalances.Details.EthUtilizationRate.RawValue
	response.NodeFee = networkFees.Details.NodeFee.RawValue
	response.NodeCount = nodeMgr.Details.NodeCount.Formatted()
	response.RplPrice = networkPrices.Details.RplPrice.RawValue
	response.TotalRplStaked = nodeStaking.Details.TotalRplStake
	response.RethPrice = reth.Details.ExchangeRate.RawValue

	// Get the total effective RPL stake
	effectiveRplStaked, err := nodeMgr.GetTotalEffectiveRplStake(rp, nodeMgr.Details.NodeCount.Formatted(), nil)
	if err != nil {
		return nil, fmt.Errorf("error getting total effective RPL stake: %w", err)
	}
	response.EffectiveRplStaked = effectiveRplStaked

	// Get the minipool counts by status
	response.FinalizedMinipoolCount = mpMgr.Details.FinalisedMinipoolCount.Formatted()
	minipoolCounts, err := mpMgr.GetMinipoolCountPerStatus(mpMgr.Details.MinipoolCount.Formatted(), nil)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool counts per status: %w", err)
	}
	response.InitializedMinipoolCount = minipoolCounts.Initialized.Uint64()
	response.PrelaunchMinipoolCount = minipoolCounts.Prelaunch.Uint64()
	response.StakingMinipoolCount = minipoolCounts.Staking.Uint64()
	response.WithdrawableMinipoolCount = minipoolCounts.Withdrawable.Uint64()
	response.DissolvedMinipoolCount = minipoolCounts.Dissolved.Uint64()

	// Get the number of nodes opted into the smoothing pool
	spCount, err := nodeMgr.GetSmoothingPoolRegisteredNodeCount(nodeMgr.Details.NodeCount.Formatted(), nil)
	if err != nil {
		return nil, fmt.Errorf("error getting smoothing pool opt-in count: %w", err)
	}
	response.SmoothingPoolNodes = spCount

	// Get the smoothing pool balance
	response.SmoothingPoolAddress = *sp.Address
	smoothingPoolBalance, err := rp.Client.BalanceAt(context.Background(), *sp.Address, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting smoothing pool balance: %w", err)
	}
	response.SmoothingPoolBalance = smoothingPoolBalance

	// Get the TVL
	activeMinipools := response.InitializedMinipoolCount +
		response.PrelaunchMinipoolCount +
		response.StakingMinipoolCount +
		response.WithdrawableMinipoolCount +
		response.DissolvedMinipoolCount
	tvl := big.NewInt(int64(activeMinipools))
	tvl.Mul(tvl, big.NewInt(32))
	tvl.Add(tvl, response.DepositPoolBalance)
	tvl.Add(tvl, response.MinipoolCapacity)
	rplWorth := big.NewInt(0).Mul(response.TotalRplStaked, response.RplPrice)
	tvl.Add(tvl, rplWorth)
	response.TotalValueLocked = tvl

	// Return response
	return &response, nil

}
