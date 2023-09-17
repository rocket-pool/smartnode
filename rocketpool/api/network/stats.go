package network

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/deposit"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/tokens"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type networkStatsContextFactory struct {
	handler *NetworkHandler
}

func (f *networkStatsContextFactory) Create(vars map[string]string) (*networkStatsContext, error) {
	c := &networkStatsContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *networkStatsContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*networkStatsContext, api.NetworkStatsData](
		router, "stats", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type networkStatsContext struct {
	handler *NetworkHandler
	rp      *rocketpool.RocketPool

	depositPool     *deposit.DepositPool
	mpQueue         *minipool.MinipoolQueue
	networkBalances *network.NetworkBalances
	networkPrices   *network.NetworkPrices
	networkFees     *network.NetworkFees
	nodeMgr         *node.NodeManager
	nodeStaking     *node.NodeStaking
	mpMgr           *minipool.MinipoolManager
	reth            *tokens.TokenReth
	smoothingPool   *core.Contract
}

func (c *networkStatsContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()

	// Requirements
	err := sp.RequireEthClientSynced()
	if err != nil {
		return err
	}

	// Bindings
	c.depositPool, err = deposit.NewDepositPool(c.rp)
	if err != nil {
		return fmt.Errorf("error getting deposit pool binding: %w", err)
	}
	c.mpQueue, err = minipool.NewMinipoolQueue(c.rp)
	if err != nil {
		return fmt.Errorf("error getting minipool queue binding: %w", err)
	}
	c.networkBalances, err = network.NewNetworkBalances(c.rp)
	if err != nil {
		return fmt.Errorf("error getting network balances binding: %w", err)
	}
	c.networkPrices, err = network.NewNetworkPrices(c.rp)
	if err != nil {
		return fmt.Errorf("error getting network prices binding: %w", err)
	}
	c.networkFees, err = network.NewNetworkFees(c.rp)
	if err != nil {
		return fmt.Errorf("error getting network fees binding: %w", err)
	}
	c.nodeMgr, err = node.NewNodeManager(c.rp)
	if err != nil {
		return fmt.Errorf("error getting node manager binding: %w", err)
	}
	c.nodeStaking, err = node.NewNodeStaking(c.rp)
	if err != nil {
		return fmt.Errorf("error getting node staking binding: %w", err)
	}
	c.mpMgr, err = minipool.NewMinipoolManager(c.rp)
	if err != nil {
		return fmt.Errorf("error getting minipool manager binding: %w", err)
	}
	c.reth, err = tokens.NewTokenReth(c.rp)
	if err != nil {
		return fmt.Errorf("error getting rETH token binding: %w", err)
	}
	c.smoothingPool, err = c.rp.GetContract(rocketpool.ContractName_RocketSmoothingPool)
	if err != nil {
		return fmt.Errorf("error getting rETH token binding: %w", err)
	}
	return nil
}

func (c *networkStatsContext) GetState(mc *batch.MultiCaller) {
	c.depositPool.GetBalance(mc)
	c.mpQueue.GetTotalCapacity(mc)
	c.networkBalances.GetEthUtilizationRate(mc)
	c.networkFees.GetNodeFee(mc)
	c.nodeMgr.GetNodeCount(mc)
	c.mpMgr.GetMinipoolCount(mc)
	c.mpMgr.GetFinalisedMinipoolCount(mc)
	c.networkPrices.GetRplPrice(mc)
	c.nodeStaking.GetTotalRPLStake(mc)
	c.reth.GetExchangeRate(mc)
}

func (c *networkStatsContext) PrepareData(data *api.NetworkStatsData, opts *bind.TransactOpts) error {
	// Handle the details
	data.DepositPoolBalance = c.depositPool.Details.Balance
	data.MinipoolCapacity = c.mpQueue.Details.TotalCapacity
	data.StakerUtilization = c.networkBalances.Details.EthUtilizationRate.RawValue
	data.NodeFee = c.networkFees.Details.NodeFee.RawValue
	data.NodeCount = c.nodeMgr.Details.NodeCount.Formatted()
	data.RplPrice = c.networkPrices.Details.RplPrice.RawValue
	data.TotalRplStaked = c.nodeStaking.Details.TotalRplStake
	data.RethPrice = c.reth.Details.ExchangeRate.RawValue

	// Get the total effective RPL stake
	effectiveRplStaked, err := c.nodeMgr.GetTotalEffectiveRplStake(c.rp, c.nodeMgr.Details.NodeCount.Formatted(), nil)
	if err != nil {
		return fmt.Errorf("error getting total effective RPL stake: %w", err)
	}
	data.EffectiveRplStaked = effectiveRplStaked

	// Get the minipool counts by status
	data.FinalizedMinipoolCount = c.mpMgr.Details.FinalisedMinipoolCount.Formatted()
	minipoolCounts, err := c.mpMgr.GetMinipoolCountPerStatus(c.mpMgr.Details.MinipoolCount.Formatted(), nil)
	if err != nil {
		return fmt.Errorf("error getting minipool counts per status: %w", err)
	}
	data.InitializedMinipoolCount = minipoolCounts.Initialized.Uint64()
	data.PrelaunchMinipoolCount = minipoolCounts.Prelaunch.Uint64()
	data.StakingMinipoolCount = minipoolCounts.Staking.Uint64()
	data.WithdrawableMinipoolCount = minipoolCounts.Withdrawable.Uint64()
	data.DissolvedMinipoolCount = minipoolCounts.Dissolved.Uint64()

	// Get the number of nodes opted into the smoothing pool
	spCount, err := c.nodeMgr.GetSmoothingPoolRegisteredNodeCount(c.nodeMgr.Details.NodeCount.Formatted(), nil)
	if err != nil {
		return fmt.Errorf("error getting smoothing pool opt-in count: %w", err)
	}
	data.SmoothingPoolNodes = spCount

	// Get the smoothing pool balance
	data.SmoothingPoolAddress = *c.smoothingPool.Address
	smoothingPoolBalance, err := c.rp.Client.BalanceAt(context.Background(), *c.smoothingPool.Address, nil)
	if err != nil {
		return fmt.Errorf("error getting smoothing pool balance: %w", err)
	}
	data.SmoothingPoolBalance = smoothingPoolBalance

	// Get the TVL
	activeMinipools := data.InitializedMinipoolCount +
		data.PrelaunchMinipoolCount +
		data.StakingMinipoolCount +
		data.WithdrawableMinipoolCount +
		data.DissolvedMinipoolCount
	tvl := big.NewInt(int64(activeMinipools))
	tvl.Mul(tvl, big.NewInt(32))
	tvl.Add(tvl, data.DepositPoolBalance)
	tvl.Add(tvl, data.MinipoolCapacity)
	rplWorth := big.NewInt(0).Mul(data.TotalRplStaked, data.RplPrice)
	tvl.Add(tvl, rplWorth)
	data.TotalValueLocked = tvl

	return nil
}
