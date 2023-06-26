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
	networkEffectiveStakeBatchSize int = 250
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
	QueueLength                       *big.Int
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

// Create a snapshot of all of the network's details
func NewNetworkDetails(rp *rocketpool.RocketPool, contracts *NetworkContracts) (*NetworkDetails, error) {
	opts := &bind.CallOpts{
		BlockNumber: contracts.ElBlockNumber,
	}

	details := &NetworkDetails{}

	// Local vars for things that need to be converted
	var rewardIndex *big.Int
	var intervalStart *big.Int
	var intervalDuration *big.Int
	var scrubPeriodSeconds *big.Int
	var totalQueueCapacity *big.Int
	var effectiveQueueCapacity *big.Int
	var totalQueueLength *big.Int
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
	contracts.Multicaller.AddCall(contracts.RocketNetworkPrices, &details.RplPrice, "getRPLPrice")
	contracts.Multicaller.AddCall(contracts.RocketDAOProtocolSettingsNode, &details.MinCollateralFraction, "getMinimumPerMinipoolStake")
	contracts.Multicaller.AddCall(contracts.RocketDAOProtocolSettingsNode, &details.MaxCollateralFraction, "getMaximumPerMinipoolStake")
	contracts.Multicaller.AddCall(contracts.RocketRewardsPool, &rewardIndex, "getRewardIndex")
	contracts.Multicaller.AddCall(contracts.RocketRewardsPool, &intervalStart, "getClaimIntervalTimeStart")
	contracts.Multicaller.AddCall(contracts.RocketRewardsPool, &intervalDuration, "getClaimIntervalTime")
	contracts.Multicaller.AddCall(contracts.RocketRewardsPool, &details.NodeOperatorRewardsPercent, "getClaimingContractPerc", "rocketClaimNode")
	contracts.Multicaller.AddCall(contracts.RocketRewardsPool, &details.TrustedNodeOperatorRewardsPercent, "getClaimingContractPerc", "rocketClaimTrustedNode")
	contracts.Multicaller.AddCall(contracts.RocketRewardsPool, &details.ProtocolDaoRewardsPercent, "getClaimingContractPerc", "rocketClaimDAO")
	contracts.Multicaller.AddCall(contracts.RocketRewardsPool, &details.PendingRPLRewards, "getPendingRPLRewards")
	contracts.Multicaller.AddCall(contracts.RocketDAONodeTrustedSettingsMinipool, &scrubPeriodSeconds, "getScrubPeriod")
	contracts.Multicaller.AddCall(contracts.RocketDepositPool, &details.DepositPoolBalance, "getBalance")
	contracts.Multicaller.AddCall(contracts.RocketDepositPool, &details.DepositPoolExcess, "getExcessBalance")
	contracts.Multicaller.AddCall(contracts.RocketMinipoolQueue, &totalQueueCapacity, "getTotalCapacity")
	contracts.Multicaller.AddCall(contracts.RocketMinipoolQueue, &effectiveQueueCapacity, "getEffectiveCapacity")
	contracts.Multicaller.AddCall(contracts.RocketMinipoolQueue, &totalQueueLength, "getTotalLength")
	contracts.Multicaller.AddCall(contracts.RocketTokenRPL, &details.RPLInflationIntervalRate, "getInflationIntervalRate")
	contracts.Multicaller.AddCall(contracts.RocketTokenRPL, &details.RPLTotalSupply, "totalSupply")
	contracts.Multicaller.AddCall(contracts.RocketNetworkPrices, &pricesBlock, "getPricesBlock")
	contracts.Multicaller.AddCall(contracts.RocketNetworkPrices, &latestReportablePricesBlock, "getLatestReportableBlock")
	contracts.Multicaller.AddCall(contracts.RocketNetworkBalances, &ethUtilizationRate, "getETHUtilizationRate")
	contracts.Multicaller.AddCall(contracts.RocketNetworkBalances, &details.StakingETHBalance, "getStakingETHBalance")
	contracts.Multicaller.AddCall(contracts.RocketTokenRETH, &rETHExchangeRate, "getExchangeRate")
	contracts.Multicaller.AddCall(contracts.RocketNetworkBalances, &details.TotalETHBalance, "getTotalETHBalance")
	contracts.Multicaller.AddCall(contracts.RocketTokenRETH, &details.TotalRETHSupply, "totalSupply")
	contracts.Multicaller.AddCall(contracts.RocketNodeStaking, &details.TotalRPLStake, "getTotalRPLStake")
	contracts.Multicaller.AddCall(contracts.RocketNetworkFees, &nodeFee, "getNodeFee")
	contracts.Multicaller.AddCall(contracts.RocketNetworkBalances, &balancesBlock, "getBalancesBlock")
	contracts.Multicaller.AddCall(contracts.RocketNetworkBalances, &latestReportableBalancesBlock, "getLatestReportableBlock")
	contracts.Multicaller.AddCall(contracts.RocketDAOProtocolSettingsNetwork, &details.SubmitBalancesEnabled, "getSubmitBalancesEnabled")
	contracts.Multicaller.AddCall(contracts.RocketDAOProtocolSettingsNetwork, &details.SubmitPricesEnabled, "getSubmitPricesEnabled")
	contracts.Multicaller.AddCall(contracts.RocketDAOProtocolSettingsMinipool, &minipoolLaunchTimeout, "getLaunchTimeout")

	// Atlas things
	contracts.Multicaller.AddCall(contracts.RocketDAONodeTrustedSettingsMinipool, &promotionScrubPeriodSeconds, "getPromotionScrubPeriod")
	contracts.Multicaller.AddCall(contracts.RocketDAONodeTrustedSettingsMinipool, &windowStartRaw, "getBondReductionWindowStart")
	contracts.Multicaller.AddCall(contracts.RocketDAONodeTrustedSettingsMinipool, &windowLengthRaw, "getBondReductionWindowLength")
	contracts.Multicaller.AddCall(contracts.RocketDepositPool, &details.DepositPoolUserBalance, "getUserBalance")

	_, err := contracts.Multicaller.FlexibleCall(true, opts)
	if err != nil {
		return nil, fmt.Errorf("error executing multicall: %w", err)
	}

	// Conversion for raw parameters
	details.RewardIndex = rewardIndex.Uint64()
	details.IntervalStart = convertToTime(intervalStart)
	details.IntervalDuration = convertToDuration(intervalDuration)
	details.ScrubPeriod = convertToDuration(scrubPeriodSeconds)
	details.SmoothingPoolAddress = *contracts.RocketSmoothingPool.Address
	details.QueueCapacity = minipool.QueueCapacity{
		Total:     totalQueueCapacity,
		Effective: effectiveQueueCapacity,
	}
	details.QueueLength = totalQueueLength
	details.PricesBlock = pricesBlock.Uint64()
	details.LatestReportablePricesBlock = latestReportablePricesBlock.Uint64()
	details.ETHUtilizationRate = eth.WeiToEth(ethUtilizationRate)
	details.RETHExchangeRate = eth.WeiToEth(rETHExchangeRate)
	details.NodeFee = eth.WeiToEth(nodeFee)
	details.BalancesBlock = balancesBlock
	details.LatestReportableBalancesBlock = latestReportableBalancesBlock
	details.MinipoolLaunchTimeout = minipoolLaunchTimeout
	details.PromotionScrubPeriod = convertToDuration(promotionScrubPeriodSeconds)
	details.BondReductionWindowStart = convertToDuration(windowStartRaw)
	details.BondReductionWindowLength = convertToDuration(windowLengthRaw)

	// Get various balances
	addresses := []common.Address{
		*contracts.RocketSmoothingPool.Address,
		*contracts.RocketTokenRETH.Address,
	}
	balances, err := contracts.BalanceBatcher.GetEthBalances(addresses, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting contract balances: %w", err)
	}
	details.SmoothingPoolBalance = balances[0]
	details.RETHBalance = balances[1]

	return details, nil
}

// Gets the details for a node using the efficient multicall contract
func GetTotalEffectiveRplStake(rp *rocketpool.RocketPool, contracts *NetworkContracts) (*big.Int, error) {
	opts := &bind.CallOpts{
		BlockNumber: contracts.ElBlockNumber,
	}

	// Get the list of node addresses
	addresses, err := getNodeAddressesFast(rp, contracts, opts)
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
			mc, err := multicall.NewMultiCaller(rp.Client, contracts.Multicaller.ContractAddress)
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
