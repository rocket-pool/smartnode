package state

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/rocketpool-go/utils/multicall"
	"golang.org/x/sync/errgroup"
)

const (
	networkEffectiveStakeBatchSize int = 2000
)

type NetworkDetails struct {
	// Redstone
	RplPrice                          *big.Int
	MinCollateralFraction             *big.Int
	MaxCollateralFraction             *big.Int
	IntervalDuration                  time.Duration
	IntervalStart                     time.Time
	NodeOperatorRewardsPercent        *big.Int
	TrustedNodeOperatorRewardsPercent *big.Int
	ProtocolDaoRewardsPercent         *big.Int
	PendingRPLRewards                 *big.Int
	RewardIndex                       uint64
	ScrubPeriod                       time.Duration
	SmoothingPoolAddress              common.Address
	DepositPoolBalance                *big.Int
	DepositPoolExcess                 *big.Int
	QueueCapacity                     minipool.QueueCapacity
	RPLInflationIntervalRate          *big.Int
	RPLTotalSupply                    *big.Int
	PricesBlock                       uint64
	LatestReportablePricesBlock       uint64
	ETHUtilizationRate                float64
	StakingETHBalance                 *big.Int
	RETHExchangeRate                  float64
	TotalETHBalance                   *big.Int
	RETHBalance                       *big.Int
	TotalRETHSupply                   *big.Int
	TotalRPLStake                     *big.Int
	SmoothingPoolBalance              *big.Int
	NodeFee                           float64
	BalancesBlock                     *big.Int
	LatestReportableBalancesBlock     *big.Int
	SubmitBalancesEnabled             bool
	SubmitPricesEnabled               bool
	MinipoolLaunchTimeout             *big.Int

	// Atlas
	PromotionScrubPeriod      time.Duration
	BondReductionWindowStart  time.Duration
	BondReductionWindowLength time.Duration
	DepositPoolUserBalance    *big.Int
}

// TODO: Finish this, involves porting e.g. GetClaimIntervalTime() over
func _getNetworkDetailsFast(rp *rocketpool.RocketPool, multicallerAddress common.Address, balanceBatcherAddress common.Address, contracts *NetworkContracts, isAtlasDeployed bool, opts *bind.CallOpts) (*NetworkDetails, error) {
	// Create the multicaller
	mc, err := multicall.NewMultiCaller(rp.Client, multicallerAddress)
	if err != nil {
		return nil, err
	}

	// Create the balance batcher
	balanceBatcher, err := multicall.NewBalanceBatcher(rp.Client, balanceBatcherAddress)
	if err != nil {
		return nil, err
	}

	details := &NetworkDetails{}

	// Local vars for things that need to be converted
	var rewardIndex *big.Int
	var scrubPeriodSeconds *big.Int
	var totalQueueCapacity *big.Int
	var effectiveQueueCapacity *big.Int
	var pricesBlock *big.Int
	var latestReportablePricesBlock *big.Int
	var ethUtilizationRate *big.Int
	var rETHExchangeRate *big.Int
	var nodeFee *big.Int
	var balancesBlock *big.Int
	var latestReportableBalancesBlock *big.Int
	var minipoolLaunchTimeout *big.Int
	var promotionScrubPeriodSeconds *big.Int
	var windowStartRaw *big.Int
	var windowLengthRaw *big.Int

	// Multicall getters
	mc.AddCall(contracts.RocketNetworkPrices, &details.RplPrice, "getRPLPrice")
	mc.AddCall(contracts.RocketDAOProtocolSettingsNode, &details.MinCollateralFraction, "getMinimumPerMinipoolStake")
	mc.AddCall(contracts.RocketDAOProtocolSettingsNode, &details.MaxCollateralFraction, "getMaximumPerMinipoolStake")
	mc.AddCall(contracts.RocketRewardsPool, &rewardIndex, "getRewardIndex")

	mc.AddCall(contracts.RocketRewardsPool, &details.IntervalStart, "getClaimIntervalTimeStart")
	mc.AddCall(contracts.RocketDAONodeTrustedSettingsMinipool, &scrubPeriodSeconds, "getScrubPeriod")
	mc.AddCall(contracts.RocketDepositPool, &details.DepositPoolBalance, "getBalance")
	mc.AddCall(contracts.RocketDepositPool, &details.DepositPoolExcess, "getExcessBalance")
	mc.AddCall(contracts.RocketMinipoolQueue, &totalQueueCapacity, "getTotalCapacity")
	mc.AddCall(contracts.RocketMinipoolQueue, &effectiveQueueCapacity, "getEffectiveCapacity")
	mc.AddCall(contracts.RocketTokenRPL, &details.RPLInflationIntervalRate, "getInflationIntervalRate")
	mc.AddCall(contracts.RocketTokenRPL, &details.RPLTotalSupply, "totalSupply")
	mc.AddCall(contracts.RocketNetworkPrices, &pricesBlock, "getPricesBlock")
	mc.AddCall(contracts.RocketNetworkPrices, &latestReportablePricesBlock, "getLatestReportableBlock")
	mc.AddCall(contracts.RocketNetworkBalances, &ethUtilizationRate, "getETHUtilizationRate")
	mc.AddCall(contracts.RocketNetworkBalances, &details.StakingETHBalance, "getStakingETHBalance")
	mc.AddCall(contracts.RocketTokenRETH, &rETHExchangeRate, "getExchangeRate")
	mc.AddCall(contracts.RocketNetworkBalances, &details.TotalETHBalance, "getTotalETHBalance")
	mc.AddCall(contracts.RocketTokenRETH, &details.TotalRETHSupply, "totalSupply")
	mc.AddCall(contracts.RocketNodeStaking, &details.TotalRPLStake, "getTotalRPLStake")
	mc.AddCall(contracts.RocketNetworkFees, &nodeFee, "getNodeFee")
	mc.AddCall(contracts.RocketNetworkBalances, &balancesBlock, "getBalancesBlock")
	mc.AddCall(contracts.RocketNetworkBalances, &latestReportableBalancesBlock, "getLatestReportableBlock")
	mc.AddCall(contracts.RocketDAOProtocolSettingsNetwork, &details.SubmitBalancesEnabled, "getSubmitBalancesEnabled")
	mc.AddCall(contracts.RocketDAOProtocolSettingsNetwork, &details.SubmitPricesEnabled, "getSubmitPricesEnabled")
	mc.AddCall(contracts.RocketDAOProtocolSettingsMinipool, &minipoolLaunchTimeout, "getLaunchTimeout")

	if isAtlasDeployed {
		mc.AddCall(contracts.RocketDAONodeTrustedSettingsMinipool, &promotionScrubPeriodSeconds, "getPromotionScrubPeriod")
		mc.AddCall(contracts.RocketDAONodeTrustedSettingsMinipool, &windowStartRaw, "getBondReductionWindowStart")
		mc.AddCall(contracts.RocketDAONodeTrustedSettingsMinipool, &windowLengthRaw, "getBondReductionWindowLength")
		mc.AddCall(contracts.RocketDepositPool, &details.DepositPoolUserBalance, "getUserBalance")
	}

	_, err = mc.FlexibleCall(true, opts)
	if err != nil {
		return nil, fmt.Errorf("error executing multicall: %w", err)
	}

	// Conversion for raw parameters
	details.RewardIndex = rewardIndex.Uint64()
	details.ScrubPeriod = time.Duration(scrubPeriodSeconds.Uint64()) * time.Second
	details.SmoothingPoolAddress = *contracts.RocketSmoothingPool.Address
	details.QueueCapacity = minipool.QueueCapacity{
		Total:     totalQueueCapacity,
		Effective: effectiveQueueCapacity,
	}
	details.PricesBlock = pricesBlock.Uint64()
	details.LatestReportablePricesBlock = latestReportablePricesBlock.Uint64()
	details.ETHUtilizationRate = eth.WeiToEth(ethUtilizationRate)
	details.RETHExchangeRate = eth.WeiToEth(rETHExchangeRate)
	details.NodeFee = eth.WeiToEth(nodeFee)
	details.BalancesBlock = balancesBlock
	details.LatestReportableBalancesBlock = latestReportableBalancesBlock
	details.MinipoolLaunchTimeout = minipoolLaunchTimeout
	details.PromotionScrubPeriod = time.Duration(promotionScrubPeriodSeconds.Uint64()) * time.Second
	details.BondReductionWindowStart = time.Duration(windowStartRaw.Uint64()) * time.Second
	details.BondReductionWindowLength = time.Duration(windowLengthRaw.Uint64()) * time.Second

	// Get various balances
	addresses := []common.Address{
		*contracts.RocketSmoothingPool.Address,
		*contracts.RocketTokenRETH.Address,
	}
	balances, err := balanceBatcher.GetEthBalances(addresses, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting contract balances: %w", err)
	}
	details.SmoothingPoolBalance = balances[0]
	details.RETHBalance = balances[1]

	// PORT THIS
	/*
		wg.Go(func() error {
			var err error
			state.NetworkDetails.IntervalDuration, err = GetClaimIntervalTime(cfg, state.NetworkDetails.RewardIndex, rp, opts)
			if err != nil {
				return fmt.Errorf("error getting interval duration: %w", err)
			}
			return nil
		})

		wg.Go(func() error {
			var err error
			state.NetworkDetails.NodeOperatorRewardsPercent, err = GetNodeOperatorRewardsPercent(cfg, state.NetworkDetails.RewardIndex, rp, opts)
			if err != nil {
				return fmt.Errorf("error getting node operator rewards percent")
			}
			return nil
		})

		wg.Go(func() error {
			var err error
			state.NetworkDetails.TrustedNodeOperatorRewardsPercent, err = GetTrustedNodeOperatorRewardsPercent(cfg, state.NetworkDetails.RewardIndex, rp, opts)
			if err != nil {
				return fmt.Errorf("error getting trusted node operator rewards percent")
			}
			return nil
		})

		wg.Go(func() error {
			var err error
			state.NetworkDetails.ProtocolDaoRewardsPercent, err = GetProtocolDaoRewardsPercent(cfg, state.NetworkDetails.RewardIndex, rp, opts)
			if err != nil {
				return fmt.Errorf("error getting protocol DAO rewards percent")
			}
			return nil
		})

		wg.Go(func() error {
			var err error
			state.NetworkDetails.PendingRPLRewards, err = GetPendingRPLRewards(cfg, state.NetworkDetails.RewardIndex, rp, opts)
			if err != nil {
				return fmt.Errorf("error getting pending RPL rewards")
			}
			return nil
		})
	*/

	return details, nil
}

// Gets the details for a node using the efficient multicall contract
func GetTotalEffectiveRplStake(rp *rocketpool.RocketPool, multicallerAddress common.Address, contracts *NetworkContracts, opts *bind.CallOpts) (*big.Int, error) {
	// Get the list of node addresses
	addresses, err := getNodeAddressesFast(rp, contracts, multicallerAddress, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting node addresses: %w", err)
	}
	count := len(addresses)
	minimumStakes := make([]*big.Int, count)
	effectiveStakes := make([]*big.Int, count)

	// Sync
	var wg errgroup.Group
	wg.SetLimit(threadLimit)

	// Run the getters in batches
	for i := 0; i < count; i += networkEffectiveStakeBatchSize {
		i := i
		max := i + networkEffectiveStakeBatchSize
		if max > count {
			max = count
		}

		wg.Go(func() error {
			var err error
			mc, err := multicall.NewMultiCaller(rp.Client, multicallerAddress)
			if err != nil {
				return err
			}
			for j := i; j < max; j++ {
				address := addresses[j]
				mc.AddCall(contracts.RocketNodeStaking, &minimumStakes[j], "getNodeMinimumRPLStake", address)
				mc.AddCall(contracts.RocketNodeStaking, &effectiveStakes[j], "getNodeEffectiveRPLStake", address)
			}
			_, err = mc.FlexibleCall(true, opts)
			if err != nil {
				return fmt.Errorf("error executing multicall: %w", err)
			}
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, fmt.Errorf("error getting effective stakes for all nodes: %w", err)
	}

	totalEffectiveStake := big.NewInt(0)
	for i, effectiveStake := range effectiveStakes {
		minimumStake := minimumStakes[i]
		// Fix the effective stake
		if effectiveStake.Cmp(minimumStake) >= 0 {
			totalEffectiveStake.Add(totalEffectiveStake, effectiveStake)
		}
	}

	return totalEffectiveStake, nil
}
