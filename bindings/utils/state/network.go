package state

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/minipool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/rocket-pool/smartnode/bindings/utils/multicall"
	"golang.org/x/sync/errgroup"
)

const (
	networkEffectiveStakeBatchSize int = 250
)

type NetworkDetails struct {
	// Redstone
	RplPrice                          *big.Int               `json:"rpl_price"`
	MinCollateralFraction             *big.Int               `json:"min_collateral_fraction"`
	MaxCollateralFraction             *big.Int               `json:"max_collateral_fraction"`
	IntervalDuration                  time.Duration          `json:"interval_duration"`
	IntervalStart                     time.Time              `json:"interval_start"`
	NodeOperatorRewardsPercent        *big.Int               `json:"node_operator_rewards_percent"`
	TrustedNodeOperatorRewardsPercent *big.Int               `json:"trusted_node_operator_rewards_percent"`
	ProtocolDaoRewardsPercent         *big.Int               `json:"protocol_dao_rewards_percent"`
	PendingRPLRewards                 *big.Int               `json:"pending_rpl_rewards"`
	RewardIndex                       uint64                 `json:"reward_index"`
	ScrubPeriod                       time.Duration          `json:"scrub_period"`
	SmoothingPoolAddress              common.Address         `json:"smoothing_pool_address"`
	DepositPoolBalance                *big.Int               `json:"deposit_pool_balance"`
	DepositPoolExcess                 *big.Int               `json:"deposit_pool_excess"`
	QueueCapacity                     minipool.QueueCapacity `json:"queue_capacity"`
	QueueLength                       *big.Int               `json:"queue_length"`
	RPLInflationIntervalRate          *big.Int               `json:"rpl_inflation_interval_rate"`
	RPLTotalSupply                    *big.Int               `json:"rpl_total_supply"`
	PricesBlock                       uint64                 `json:"prices_block"`
	LatestReportablePricesBlock       uint64                 `json:"latest_reportable_prices_block"`
	ETHUtilizationRate                float64                `json:"eth_utilization_rate"`
	StakingETHBalance                 *big.Int               `json:"staking_eth_balance"`
	RETHExchangeRate                  float64                `json:"reth_exchange_rate"`
	TotalETHBalance                   *big.Int               `json:"total_eth_balance"`
	RETHBalance                       *big.Int               `json:"reth_balance"`
	TotalRETHSupply                   *big.Int               `json:"total_reth_supply"`
	TotalRPLStake                     *big.Int               `json:"total_rpl_stake"`
	SmoothingPoolBalance              *big.Int               `json:"smoothing_pool_balance"`
	PendingVoterShare                 *big.Int               `json:"pending_voter_share"`
	NodeFee                           float64                `json:"node_fee"`
	BalancesBlock                     uint64                 `json:"balances_block"`
	LatestReportableBalancesBlock     uint64                 `json:"latest_reportable_balances_block"`
	SubmitBalancesEnabled             bool                   `json:"submit_balances_enabled"`
	SubmitPricesEnabled               bool                   `json:"submit_prices_enabled"`
	MinipoolLaunchTimeout             *big.Int               `json:"minipool_launch_timeout"`

	// Atlas
	PromotionScrubPeriod      time.Duration `json:"promotion_scrub_period"`
	BondReductionWindowStart  time.Duration `json:"bond_reduction_window_start"`
	BondReductionWindowLength time.Duration `json:"bond_reduction_window_length"`
	DepositPoolUserBalance    *big.Int      `json:"deposit_pool_user_balance"`

	// Houston
	PricesSubmissionFrequency   uint64 `json:"prices_submission_frequency"`
	BalancesSubmissionFrequency uint64 `json:"balances_submission_frequency"`

	// Saturn
	MegapoolRevenueSplitSettings struct {
		NodeOperatorCommissionShare *big.Int `json:"node_operator_commission_share"`
		NodeOperatorCommissionAdder *big.Int `json:"node_operator_commission_adder"`
		VoterCommissionShare        *big.Int `json:"voter_commission_share"`
		PdaoCommissionShare         *big.Int `json:"pdao_commission_share"`
	}

	MegapoolRevenueSplitTimeWeightedAverages struct {
		NodeShare  *big.Int `json:"node_share"`
		VoterShare *big.Int `json:"voter_share"`
		PdaoShare  *big.Int `json:"pdao_share"`
	}
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
	var pricesSubmissionFrequency *big.Int
	var ethUtilizationRate *big.Int
	var rETHExchangeRate *big.Int
	var nodeFee *big.Int
	var balancesBlock *big.Int
	var balancesSubmissionFrequency *big.Int
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
	contracts.Multicaller.AddCall(contracts.RocketRewardsPool, &details.PendingVoterShare, "getPendingVoterShare")
	contracts.Multicaller.AddCall(contracts.RocketDAONodeTrustedSettingsMinipool, &scrubPeriodSeconds, "getScrubPeriod")
	contracts.Multicaller.AddCall(contracts.RocketDepositPool, &details.DepositPoolBalance, "getBalance")
	contracts.Multicaller.AddCall(contracts.RocketDepositPool, &details.DepositPoolExcess, "getExcessBalance")
	contracts.Multicaller.AddCall(contracts.RocketMinipoolQueue, &totalQueueCapacity, "getTotalCapacity")
	contracts.Multicaller.AddCall(contracts.RocketMinipoolQueue, &effectiveQueueCapacity, "getEffectiveCapacity")
	contracts.Multicaller.AddCall(contracts.RocketMinipoolQueue, &totalQueueLength, "getTotalLength")
	contracts.Multicaller.AddCall(contracts.RocketTokenRPL, &details.RPLInflationIntervalRate, "getInflationIntervalRate")
	contracts.Multicaller.AddCall(contracts.RocketTokenRPL, &details.RPLTotalSupply, "totalSupply")
	contracts.Multicaller.AddCall(contracts.RocketNetworkPrices, &pricesBlock, "getPricesBlock")
	contracts.Multicaller.AddCall(contracts.RocketNetworkBalances, &ethUtilizationRate, "getETHUtilizationRate")
	contracts.Multicaller.AddCall(contracts.RocketNetworkBalances, &details.StakingETHBalance, "getStakingETHBalance")
	contracts.Multicaller.AddCall(contracts.RocketTokenRETH, &rETHExchangeRate, "getExchangeRate")
	contracts.Multicaller.AddCall(contracts.RocketNetworkBalances, &details.TotalETHBalance, "getTotalETHBalance")
	contracts.Multicaller.AddCall(contracts.RocketTokenRETH, &details.TotalRETHSupply, "totalSupply")
	contracts.Multicaller.AddCall(contracts.RocketNodeStaking, &details.TotalRPLStake, "getTotalRPLStake")
	contracts.Multicaller.AddCall(contracts.RocketNetworkFees, &nodeFee, "getNodeFee")
	contracts.Multicaller.AddCall(contracts.RocketNetworkBalances, &balancesBlock, "getBalancesBlock")
	contracts.Multicaller.AddCall(contracts.RocketDAOProtocolSettingsNetwork, &details.SubmitBalancesEnabled, "getSubmitBalancesEnabled")
	contracts.Multicaller.AddCall(contracts.RocketDAOProtocolSettingsNetwork, &details.SubmitPricesEnabled, "getSubmitPricesEnabled")
	contracts.Multicaller.AddCall(contracts.RocketDAOProtocolSettingsMinipool, &minipoolLaunchTimeout, "getLaunchTimeout")

	// Atlas things
	contracts.Multicaller.AddCall(contracts.RocketDAONodeTrustedSettingsMinipool, &promotionScrubPeriodSeconds, "getPromotionScrubPeriod")
	contracts.Multicaller.AddCall(contracts.RocketDAONodeTrustedSettingsMinipool, &windowStartRaw, "getBondReductionWindowStart")
	contracts.Multicaller.AddCall(contracts.RocketDAONodeTrustedSettingsMinipool, &windowLengthRaw, "getBondReductionWindowLength")
	contracts.Multicaller.AddCall(contracts.RocketDepositPool, &details.DepositPoolUserBalance, "getUserBalance")

	// Houston
	contracts.Multicaller.AddCall(contracts.RocketDAOProtocolSettingsNetwork, &pricesSubmissionFrequency, "getSubmitPricesFrequency")
	contracts.Multicaller.AddCall(contracts.RocketDAOProtocolSettingsNetwork, &balancesSubmissionFrequency, "getSubmitBalancesFrequency")

	// Saturn
	contracts.Multicaller.AddCall(contracts.RocketDAOProtocolSettingsNetwork, &details.MegapoolRevenueSplitSettings.NodeOperatorCommissionShare, "getNodeShare")
	contracts.Multicaller.AddCall(contracts.RocketDAOProtocolSettingsNetwork, &details.MegapoolRevenueSplitSettings.NodeOperatorCommissionAdder, "getNodeShareSecurityCouncilAdder")
	contracts.Multicaller.AddCall(contracts.RocketDAOProtocolSettingsNetwork, &details.MegapoolRevenueSplitSettings.VoterCommissionShare, "getVoterShare")
	contracts.Multicaller.AddCall(contracts.RocketDAOProtocolSettingsNetwork, &details.MegapoolRevenueSplitSettings.PdaoCommissionShare, "getProtocolDAOShare")

	contracts.Multicaller.AddCall(contracts.RocketNetworkRevenues, &details.MegapoolRevenueSplitTimeWeightedAverages.NodeShare, "getCurrentNodeShare")
	contracts.Multicaller.AddCall(contracts.RocketNetworkRevenues, &details.MegapoolRevenueSplitTimeWeightedAverages.VoterShare, "getCurrentVoterShare")
	contracts.Multicaller.AddCall(contracts.RocketNetworkRevenues, &details.MegapoolRevenueSplitTimeWeightedAverages.PdaoShare, "getCurrentProtocolDAOShare")

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

	details.PricesSubmissionFrequency = pricesSubmissionFrequency.Uint64()
	details.BalancesSubmissionFrequency = balancesSubmissionFrequency.Uint64()
	details.ETHUtilizationRate = eth.WeiToEth(ethUtilizationRate)
	details.RETHExchangeRate = eth.WeiToEth(rETHExchangeRate)
	details.NodeFee = eth.WeiToEth(nodeFee)
	details.BalancesBlock = balancesBlock.Uint64()
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
		max := min(i+networkEffectiveStakeBatchSize, count)

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
