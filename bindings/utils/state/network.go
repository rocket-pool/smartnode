package state

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/bindings/minipool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/utils/multicall"
	"github.com/rocket-pool/smartnode/shared/units"
)

const (
	networkEffectiveStakeBatchSize int = 250
)

type MegapoolRevenueSplitSettings struct {
	NodeOperatorCommissionShare units.Wei `json:"node_operator_commission_share"`
	NodeOperatorCommissionAdder units.Wei `json:"node_operator_commission_adder"`
	VoterCommissionShare        units.Wei `json:"voter_commission_share"`
	PdaoCommissionShare         units.Wei `json:"pdao_commission_share"`
}
type MegapoolRevenueSplitTimeWeightedAverages struct {
	NodeShare  units.Wei `json:"node_share"`
	VoterShare units.Wei `json:"voter_share"`
	PdaoShare  units.Wei `json:"pdao_share"`
}

type NetworkDetails struct {
	// Redstone
	RplPrice                          units.Wei              `json:"rpl_price"`
	MinCollateralFraction             units.Wei              `json:"min_collateral_fraction"`
	MaxCollateralFraction             units.Wei              `json:"max_collateral_fraction"`
	MinimumLegacyRplStakeFraction     units.Wei              `json:"minimum_legacy_rpl_stake_fraction"`
	IntervalDuration                  time.Duration          `json:"interval_duration"`
	IntervalStart                     time.Time              `json:"interval_start"`
	NodeOperatorRewardsPercent        units.Wei              `json:"node_operator_rewards_percent"`
	TrustedNodeOperatorRewardsPercent units.Wei              `json:"trusted_node_operator_rewards_percent"`
	ProtocolDaoRewardsPercent         units.Wei              `json:"protocol_dao_rewards_percent"`
	PendingRPLRewards                 units.Wei              `json:"pending_rpl_rewards"`
	RewardIndex                       uint64                 `json:"reward_index"`
	ScrubPeriod                       time.Duration          `json:"scrub_period"`
	SmoothingPoolAddress              common.Address         `json:"smoothing_pool_address"`
	DepositPoolBalance                units.Wei              `json:"deposit_pool_balance"`
	DepositPoolExcess                 units.Wei              `json:"deposit_pool_excess"`
	QueueCapacity                     minipool.QueueCapacity `json:"queue_capacity"`
	QueueLength                       *big.Int               `json:"queue_length"`
	RPLInflationIntervalRate          units.Wei              `json:"rpl_inflation_interval_rate"`
	RPLTotalSupply                    units.Wei              `json:"rpl_total_supply"`
	PricesBlock                       uint64                 `json:"prices_block"`
	LatestReportablePricesBlock       uint64                 `json:"latest_reportable_prices_block"`
	ETHUtilizationRate                units.Eth              `json:"eth_utilization_rate"`
	StakingETHBalance                 units.Wei              `json:"staking_eth_balance"`
	RETHExchangeRate                  units.Eth              `json:"reth_exchange_rate"`
	TotalETHBalance                   units.Wei              `json:"total_eth_balance"`
	RETHBalance                       units.Wei              `json:"reth_balance"`
	TotalRETHSupply                   units.Wei              `json:"total_reth_supply"`
	TotalRPLStake                     units.Wei              `json:"total_rpl_stake"`
	TotalNetworkMegapoolStakedRpl     units.Wei              `json:"total_network_megapool_staked_rpl"`
	TotalLegacyStakedRpl              units.Wei              `json:"total_legacy_staked_rpl"`
	SmoothingPoolBalance              units.Wei              `json:"smoothing_pool_balance"`
	PendingVoterShare                 units.Wei              `json:"pending_voter_share"`
	NodeFee                           units.Eth              `json:"node_fee"`
	BalancesBlock                     uint64                 `json:"balances_block"`
	LatestReportableBalancesBlock     uint64                 `json:"latest_reportable_balances_block"`
	SubmitBalancesEnabled             bool                   `json:"submit_balances_enabled"`
	SubmitPricesEnabled               bool                   `json:"submit_prices_enabled"`
	MinipoolLaunchTimeout             *big.Int               `json:"minipool_launch_timeout"`

	// Atlas
	PromotionScrubPeriod      time.Duration `json:"promotion_scrub_period"`
	BondReductionWindowStart  time.Duration `json:"bond_reduction_window_start"`
	BondReductionWindowLength time.Duration `json:"bond_reduction_window_length"`
	DepositPoolUserBalance    units.Wei     `json:"deposit_pool_user_balance"`

	// Houston
	PricesSubmissionFrequency   uint64 `json:"prices_submission_frequency"`
	BalancesSubmissionFrequency uint64 `json:"balances_submission_frequency"`

	// Saturn
	MegapoolRevenueSplitSettings             MegapoolRevenueSplitSettings
	MegapoolRevenueSplitTimeWeightedAverages MegapoolRevenueSplitTimeWeightedAverages
	PendingVoterShareEth                     units.Wei `json:"pending_voter_share_eth"`
	ReducedBond                              units.Wei `json:"reduced_bond"`
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
	var totalQueueCapacity units.Wei
	var effectiveQueueCapacity units.Wei
	var totalQueueLength *big.Int
	var pricesBlock *big.Int
	var pricesSubmissionFrequency *big.Int
	var ethUtilizationRate units.Wei
	var rETHExchangeRate units.Wei
	var nodeFee units.Wei
	var balancesBlock *big.Int
	var balancesSubmissionFrequency *big.Int
	var minipoolLaunchTimeout *big.Int
	var promotionScrubPeriodSeconds *big.Int
	var windowStartRaw *big.Int
	var windowLengthRaw *big.Int

	allErrors := make([]error, 0)
	addCall := func(contract *rocketpool.Contract, out any, method string, args ...any) {
		allErrors = append(allErrors, contracts.Multicaller.AddCall(contract, out, method, args...))
	}

	// Multicall getters
	addCall(contracts.RocketNetworkPrices, &details.RplPrice, "getRPLPrice")
	addCall(contracts.RocketRewardsPool, &rewardIndex, "getRewardIndex")
	addCall(contracts.RocketRewardsPool, &intervalStart, "getClaimIntervalTimeStart")
	addCall(contracts.RocketRewardsPool, &intervalDuration, "getClaimIntervalTime")
	addCall(contracts.RocketRewardsPool, &details.NodeOperatorRewardsPercent, "getClaimingContractPerc", "rocketClaimNode")
	addCall(contracts.RocketRewardsPool, &details.TrustedNodeOperatorRewardsPercent, "getClaimingContractPerc", "rocketClaimTrustedNode")
	addCall(contracts.RocketRewardsPool, &details.ProtocolDaoRewardsPercent, "getClaimingContractPerc", "rocketClaimDAO")
	addCall(contracts.RocketRewardsPool, &details.PendingRPLRewards, "getPendingRPLRewards")
	addCall(contracts.RocketRewardsPool, &details.PendingVoterShare, "getPendingVoterShare")
	addCall(contracts.RocketDAONodeTrustedSettingsMinipool, &scrubPeriodSeconds, "getScrubPeriod")
	addCall(contracts.RocketDepositPool, &details.DepositPoolBalance, "getBalance")
	addCall(contracts.RocketDepositPool, &details.DepositPoolExcess, "getExcessBalance")
	addCall(contracts.RocketMinipoolQueue, &totalQueueCapacity, "getTotalCapacity")
	addCall(contracts.RocketMinipoolQueue, &effectiveQueueCapacity, "getEffectiveCapacity")
	addCall(contracts.RocketMinipoolQueue, &totalQueueLength, "getTotalLength")
	addCall(contracts.RocketTokenRPL, &details.RPLInflationIntervalRate, "getInflationIntervalRate")
	addCall(contracts.RocketTokenRPL, &details.RPLTotalSupply, "totalSupply")
	addCall(contracts.RocketNetworkPrices, &pricesBlock, "getPricesBlock")
	addCall(contracts.RocketNetworkBalances, &ethUtilizationRate, "getETHUtilizationRate")
	addCall(contracts.RocketNetworkBalances, &details.StakingETHBalance, "getStakingETHBalance")
	addCall(contracts.RocketTokenRETH, &rETHExchangeRate, "getExchangeRate")
	addCall(contracts.RocketNetworkBalances, &details.TotalETHBalance, "getTotalETHBalance")
	addCall(contracts.RocketTokenRETH, &details.TotalRETHSupply, "totalSupply")
	addCall(contracts.RocketNetworkFees, &nodeFee, "getNodeFee")
	addCall(contracts.RocketNetworkBalances, &balancesBlock, "getBalancesBlock")
	addCall(contracts.RocketDAOProtocolSettingsNetwork, &details.SubmitBalancesEnabled, "getSubmitBalancesEnabled")
	addCall(contracts.RocketDAOProtocolSettingsNetwork, &details.SubmitPricesEnabled, "getSubmitPricesEnabled")
	addCall(contracts.RocketDAOProtocolSettingsMinipool, &minipoolLaunchTimeout, "getLaunchTimeout")

	// Atlas things
	addCall(contracts.RocketDAONodeTrustedSettingsMinipool, &promotionScrubPeriodSeconds, "getPromotionScrubPeriod")
	addCall(contracts.RocketDAONodeTrustedSettingsMinipool, &windowStartRaw, "getBondReductionWindowStart")
	addCall(contracts.RocketDAONodeTrustedSettingsMinipool, &windowLengthRaw, "getBondReductionWindowLength")
	addCall(contracts.RocketDepositPool, &details.DepositPoolUserBalance, "getUserBalance")

	// Houston
	addCall(contracts.RocketDAOProtocolSettingsNetwork, &pricesSubmissionFrequency, "getSubmitPricesFrequency")
	addCall(contracts.RocketDAOProtocolSettingsNetwork, &balancesSubmissionFrequency, "getSubmitBalancesFrequency")

	// Saturn
	addCall(contracts.RocketDAOProtocolSettingsNetwork, &details.MegapoolRevenueSplitSettings.NodeOperatorCommissionShare, "getNodeShare")
	addCall(contracts.RocketDAOProtocolSettingsNetwork, &details.MegapoolRevenueSplitSettings.NodeOperatorCommissionAdder, "getNodeShareSecurityCouncilAdder")
	addCall(contracts.RocketDAOProtocolSettingsNetwork, &details.MegapoolRevenueSplitSettings.VoterCommissionShare, "getVoterShare")
	addCall(contracts.RocketDAOProtocolSettingsNetwork, &details.MegapoolRevenueSplitSettings.PdaoCommissionShare, "getProtocolDAOShare")
	addCall(contracts.RocketDAOProtocolSettingsNode, &details.ReducedBond, "getReducedBond")
	addCall(contracts.RocketDAOProtocolSettingsNode, &details.MinimumLegacyRplStakeFraction, "getMinimumLegacyRPLStake")
	addCall(contracts.RocketNetworkRevenues, &details.MegapoolRevenueSplitTimeWeightedAverages.NodeShare, "getCurrentNodeShare")
	addCall(contracts.RocketNetworkRevenues, &details.MegapoolRevenueSplitTimeWeightedAverages.VoterShare, "getCurrentVoterShare")
	addCall(contracts.RocketNetworkRevenues, &details.MegapoolRevenueSplitTimeWeightedAverages.PdaoShare, "getCurrentProtocolDAOShare")
	addCall(contracts.RocketRewardsPool, &details.PendingVoterShareEth, "getPendingVoterShare")
	addCall(contracts.RocketNodeStaking, &details.TotalNetworkMegapoolStakedRpl, "getTotalMegapoolStakedRPL")
	addCall(contracts.RocketNodeStaking, &details.TotalRPLStake, "getTotalStakedRPL")
	addCall(contracts.RocketNodeStaking, &details.TotalLegacyStakedRpl, "getTotalLegacyStakedRPL")

	for _, err := range allErrors {
		if err != nil {
			return nil, fmt.Errorf("error getting network details: %w", err)
		}
	}

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
	details.ETHUtilizationRate = ethUtilizationRate.ToEth()
	details.RETHExchangeRate = rETHExchangeRate.ToEth()
	details.NodeFee = nodeFee.ToEth()
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
func GetTotalEffectiveRplStake(rp *rocketpool.RocketPool, contracts *NetworkContracts) (units.Wei, error) {
	opts := &bind.CallOpts{
		BlockNumber: contracts.ElBlockNumber,
	}

	// Get the list of node addresses
	addresses, err := getNodeAddressesFast(rp, contracts, opts)
	if err != nil {
		return units.Wei{}, fmt.Errorf("error getting node addresses: %w", err)
	}
	count := len(addresses)
	minimumStakes := make([]units.Wei, count)
	legacyStakes := make([]units.Wei, count)
	megapoolStakes := make([]units.Wei, count)
	// Sync
	var wg errgroup.Group
	wg.SetLimit(threadLimit)

	// Run the getters in batches
	for i := 0; i < count; i += networkEffectiveStakeBatchSize {
		i := i
		m := min(i+networkEffectiveStakeBatchSize, count)

		wg.Go(func() error {
			var err error
			mc, err := multicall.NewMultiCaller(rp.Client, contracts.Multicaller.ContractAddress)
			if err != nil {
				return err
			}
			for j := i; j < m; j++ {
				address := addresses[j]
				err = mc.AddCall(contracts.RocketNodeStaking, &minimumStakes[j], "getNodeMinimumLegacyRPLStake", address)
				if err != nil {
					return fmt.Errorf("error adding node minimum legacy RPL stake call for address %s: %w", address.Hex(), err)
				}
				err = mc.AddCall(contracts.RocketNodeStaking, &legacyStakes[j], "getNodeLegacyStakedRPL", address)
				if err != nil {
					return fmt.Errorf("error adding node legacy RPL stake call for address %s: %w", address.Hex(), err)
				}
				err = mc.AddCall(contracts.RocketNodeStaking, &megapoolStakes[j], "getNodeMegapoolStakedRPL", address)
				if err != nil {
					return fmt.Errorf("error adding node megapool RPL stake call for address %s: %w", address.Hex(), err)
				}
			}
			_, err = mc.FlexibleCall(true, opts)
			if err != nil {
				return fmt.Errorf("error executing multicall: %w", err)
			}
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return units.Wei{}, fmt.Errorf("error getting effective stakes for all nodes: %w", err)
	}

	totalEffectiveStake := units.Wei{}
	for i, legacyStake := range legacyStakes {
		minimumStake := minimumStakes[i]
		if legacyStake.Cmp(minimumStake) >= 0 {
			totalEffectiveStake = totalEffectiveStake.Add(legacyStake)
		}
		totalEffectiveStake = totalEffectiveStake.Add(megapoolStakes[i])
	}

	return totalEffectiveStake, nil
}
