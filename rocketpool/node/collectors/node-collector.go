package collectors

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rprewards "github.com/rocket-pool/smartnode/shared/services/rewards"
	"github.com/rocket-pool/smartnode/shared/utils/eth2"
	"golang.org/x/sync/errgroup"
)

// Represents the collector for the user's node
type NodeCollector struct {
	// The total amount of RPL staked on the node
	totalStakedRpl *prometheus.Desc

	// The effective amount of RPL staked on the node (honoring the 150% collateral cap)
	effectiveStakedRpl *prometheus.Desc

	// The amount of staked RPL that will be eligible for rewards (including Beacon Chain data and accounding for pending bond reductions)
	rewardableStakedRpl *prometheus.Desc

	// The cumulative RPL rewards earned by the node
	cumulativeRplRewards *prometheus.Desc

	// The expected RPL rewards for the node at the next rewards checkpoint
	expectedRplRewards *prometheus.Desc

	// The estimated APR of RPL for the node from the next rewards checkpoint
	rplApr *prometheus.Desc

	// The token balances of your node wallet
	balances *prometheus.Desc

	// The number of active minipools owned by the node
	activeMinipoolCount *prometheus.Desc

	// The amount of ETH this node deposited into minipools
	depositedEth *prometheus.Desc

	// The node's total share of its minipool's beacon chain balances
	beaconShare *prometheus.Desc

	// The total balances of all this node's validators on the beacon chain
	beaconBalance *prometheus.Desc

	// The sync progress of the clients
	clientSyncProgress *prometheus.Desc

	// The total EL balance of all minipools belonging to this node
	minipoolBalance *prometheus.Desc

	// The node's share of the total minipool EL balance
	minipoolShare *prometheus.Desc

	// The amount of ETH waiting to be refunded for all minipools
	refundBalance *prometheus.Desc

	// The RPL rewards from the last period that have not been claimed yet
	unclaimedRewards *prometheus.Desc

	// The claimed ETH rewards from the smoothing pool
	claimedEthRewards *prometheus.Desc

	// The unclaimed ETH rewards from the smoothing pool
	unclaimedEthRewards *prometheus.Desc

	// The collateral ratio with respect to the amount of borrowed ETH
	borrowedCollateralRatio *prometheus.Desc

	// The collateral ratio with respect to the amount of bonded ETH
	bondedCollateralRatio *prometheus.Desc

	// The Rocket Pool contract manager
	rp *rocketpool.RocketPool

	// The beacon client
	bc *services.BeaconClientManager

	// The execution client
	ec *services.ExecutionClientManager

	// The node's address
	nodeAddress common.Address

	// The event log interval for the current eth1 client
	eventLogInterval *big.Int

	// The next block to start from when looking at cumulative RPL rewards
	nextRewardsStartBlock *big.Int

	// The cumulative amount of RPL earned
	cumulativeRewards float64

	// The claimed ETH rewards from SP
	cumulativeClaimedEthRewards float64

	// Map of reward intervals that have already been processed
	handledIntervals map[uint64]bool

	// The Rocket Pool config
	cfg *config.RocketPoolConfig

	// The thread-safe locker for the network state
	stateLocker *StateLocker

	// Prefix for logging
	logPrefix string
}

// Create a new NodeCollector instance
func NewNodeCollector(rp *rocketpool.RocketPool, bc *services.BeaconClientManager, ec *services.ExecutionClientManager, nodeAddress common.Address, cfg *config.RocketPoolConfig, stateLocker *StateLocker) *NodeCollector {

	// Get the event log interval
	eventLogInterval, err := cfg.GetEventLogInterval()
	if err != nil {
		log.Printf("Error getting event log interval: %s\n", err.Error())
		return nil
	}

	subsystem := "node"
	return &NodeCollector{
		totalStakedRpl: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "total_staked_rpl"),
			"The total amount of RPL staked on the node",
			nil, nil,
		),
		effectiveStakedRpl: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "effective_staked_rpl"),
			"The effective amount of RPL staked on the node (honoring the 150% collateral cap)",
			nil, nil,
		),
		rewardableStakedRpl: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "rewardable_staked_rpl"),
			"The amount of staked RPL that will be eligible for rewards (including Beacon Chain data and accounding for pending bond reductions)",
			nil, nil,
		),
		cumulativeRplRewards: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "cumulative_rpl_rewards"),
			"The cumulative RPL rewards earned by the node",
			nil, nil,
		),
		expectedRplRewards: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "expected_rpl_rewards"),
			"The expected RPL rewards for the node at the next rewards checkpoint",
			nil, nil,
		),
		rplApr: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "rpl_apr"),
			"The estimated APR of RPL for the node from the next rewards checkpoint",
			nil, nil,
		),
		balances: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "balance"),
			"How much ETH is in this node wallet",
			[]string{"Token"}, nil,
		),
		activeMinipoolCount: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "active_minipool_count"),
			"The number of active minipools owned by the node",
			nil, nil,
		),
		depositedEth: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "deposited_eth"),
			"The amount of ETH this node deposited into minipools",
			nil, nil,
		),
		beaconShare: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "beacon_share"),
			"The node's total share of its minipool's beacon chain balances",
			nil, nil,
		),
		beaconBalance: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "beacon_balance"),
			"The total balances of all this node's validators on the beacon chain",
			nil, nil,
		),
		clientSyncProgress: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "sync_progress"),
			"The sync progress of the beacon and execution clients",
			[]string{"client"}, nil,
		),
		minipoolBalance: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "minipool_balance"),
			"The total EL balance of all minipools belonging to this node",
			nil, nil,
		),
		minipoolShare: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "minipool_share"),
			"The node's share of the total minipool EL balance",
			nil, nil,
		),
		refundBalance: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "refund_balance"),
			"The amount of ETH waiting to be refunded for all minipools",
			nil, nil,
		),
		unclaimedRewards: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "unclaimed_rewards"),
			"The RPL rewards from the last period that have not been claimed yet",
			nil, nil,
		),
		claimedEthRewards: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "claimed_eth_rewards"),
			"The claimed ETH rewards from the smoothing pool",
			nil, nil,
		),
		unclaimedEthRewards: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "unclaimed_eth_rewards"),
			"The unclaimed ETH rewards from the smoothing pool",
			nil, nil,
		),
		borrowedCollateralRatio: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "borrowed_collateral_ratio"),
			"The collateral ratio with respect to the amount of borrowed ETH",
			nil, nil,
		),
		bondedCollateralRatio: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "bonded_collateral_ratio"),
			"The collateral ratio with respect to the amount of bonded ETH",
			nil, nil,
		),
		rp:               rp,
		bc:               bc,
		ec:               ec,
		nodeAddress:      nodeAddress,
		eventLogInterval: big.NewInt(int64(eventLogInterval)),
		handledIntervals: map[uint64]bool{},
		cfg:              cfg,
		stateLocker:      stateLocker,
		logPrefix:        "Node Collector",
	}
}

// Write metric descriptions to the Prometheus channel
func (collector *NodeCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- collector.totalStakedRpl
	channel <- collector.effectiveStakedRpl
	channel <- collector.rewardableStakedRpl
	channel <- collector.cumulativeRplRewards
	channel <- collector.expectedRplRewards
	channel <- collector.rplApr
	channel <- collector.balances
	channel <- collector.activeMinipoolCount
	channel <- collector.depositedEth
	channel <- collector.beaconBalance
	channel <- collector.beaconShare
	channel <- collector.clientSyncProgress
	channel <- collector.minipoolBalance
	channel <- collector.minipoolShare
	channel <- collector.refundBalance
	channel <- collector.unclaimedRewards
	channel <- collector.claimedEthRewards
	channel <- collector.unclaimedEthRewards
	channel <- collector.borrowedCollateralRatio
	channel <- collector.bondedCollateralRatio
}

// Collect the latest metric values and pass them to Prometheus
func (collector *NodeCollector) Collect(channel chan<- prometheus.Metric) {
	// Get the latest state
	state := collector.stateLocker.GetState()
	if state == nil {
		return
	}

	nd := state.NodeDetailsByAddress[collector.nodeAddress]
	minipools := state.MinipoolDetailsByNode[collector.nodeAddress]

	// Sync
	var wg errgroup.Group
	stakedRpl := eth.WeiToEth(nd.RplStake)
	effectiveStakedRpl := eth.WeiToEth(nd.EffectiveRPLStake)
	rewardsInterval := state.NetworkDetails.IntervalDuration
	inflationInterval := state.NetworkDetails.RPLInflationIntervalRate
	totalRplSupply := state.NetworkDetails.RPLTotalSupply
	totalEffectiveStake := collector.stateLocker.GetTotalEffectiveRPLStake()
	nodeOperatorRewardsPercent := eth.WeiToEth(state.NetworkDetails.NodeOperatorRewardsPercent)
	previousIntervalTotalNodeWeight := big.NewInt(0)
	ethBalance := eth.WeiToEth(nd.BalanceETH)
	oldRplBalance := eth.WeiToEth(nd.BalanceOldRPL)
	newRplBalance := eth.WeiToEth(nd.BalanceRPL)
	rethBalance := eth.WeiToEth(nd.BalanceRETH)
	eligibleBorrowedEth := state.GetEligibleBorrowedEth(nd)
	var activeMinipoolCount float64
	rplPriceRaw := state.NetworkDetails.RplPrice
	rplPrice := eth.WeiToEth(rplPriceRaw)
	var beaconHead beacon.BeaconHead
	unclaimedEthRewards := float64(0)
	unclaimedRplRewards := float64(0)
	if totalEffectiveStake == nil {
		return
	}

	// Get the cumulative claimed and unclaimed RPL rewards
	wg.Go(func() error {
		//legacyClaimNodeAddress := collector.cfg.Smartnode.GetLegacyClaimNodeAddress()
		//legacyRewardsPoolAddress := collector.cfg.Smartnode.GetLegacyRewardsPoolAddress()

		// Legacy rewards
		unclaimedRplWei := big.NewInt(0)
		unclaimedEthWei := big.NewInt(0)
		newRewards := big.NewInt(0)
		newClaimedEthRewards := big.NewInt(0)

		// TODO: PERFORMANCE IMPROVEMENTS
		/*newRewards, err := legacyrewards.CalculateLifetimeNodeRewards(collector.rp, collector.nodeAddress, collector.eventLogInterval, collector.nextRewardsStartBlock, &legacyRewardsPoolAddress, &legacyClaimNodeAddress)
		if err != nil {
			return fmt.Errorf("Error getting cumulative RPL rewards: %w", err)
		}*/

		// Get the claimed and unclaimed intervals
		unclaimed, claimed, err := rprewards.GetClaimStatus(collector.rp, collector.nodeAddress)
		if err != nil {
			return err
		}

		// Get the totalNodeWeight for the last completed interval
		previousRewardIndex := state.NetworkDetails.RewardIndex
		if previousRewardIndex > 0 {
			previousRewardIndex = previousRewardIndex - 1
		}

		previousInterval, err := rprewards.GetIntervalInfo(collector.rp, collector.cfg, collector.nodeAddress, previousRewardIndex, nil)
		if err != nil {
			return err
		}

		if !previousInterval.TreeFileExists {
			return fmt.Errorf("Error retrieving previous interval's total node weight: rewards file %s doesn't exist for interval %d", previousInterval.TreeFilePath, previousRewardIndex)
		}
		// Convert to a float, accuracy loss is meaningless compared to the heuristic's natural inaccuracy.
		previousIntervalTotalNodeWeight = &previousInterval.TotalNodeWeight.Int

		// Get the info for each claimed interval
		for _, claimedInterval := range claimed {
			_, exists := collector.handledIntervals[claimedInterval]
			if !exists {
				intervalInfo, err := rprewards.GetIntervalInfo(collector.rp, collector.cfg, collector.nodeAddress, claimedInterval, nil)
				if err != nil {
					return err
				}
				if !intervalInfo.TreeFileExists {
					return fmt.Errorf("Error calculating lifetime node rewards: rewards file %s doesn't exist but interval %d was claimed", intervalInfo.TreeFilePath, claimedInterval)
				}

				newRewards.Add(newRewards, &intervalInfo.CollateralRplAmount.Int)
				newClaimedEthRewards.Add(newClaimedEthRewards, &intervalInfo.SmoothingPoolEthAmount.Int)
				collector.handledIntervals[claimedInterval] = true
			}
		}
		// Get the unclaimed rewards
		for _, unclaimedInterval := range unclaimed {
			intervalInfo, err := rprewards.GetIntervalInfo(collector.rp, collector.cfg, collector.nodeAddress, unclaimedInterval, nil)
			if err != nil {
				return err
			}
			if !intervalInfo.TreeFileExists {
				return fmt.Errorf("Error calculating lifetime node rewards: rewards file %s doesn't exist and interval %d is unclaimed", intervalInfo.TreeFilePath, unclaimedInterval)
			}
			if intervalInfo.NodeExists {
				unclaimedRplWei.Add(unclaimedRplWei, &intervalInfo.CollateralRplAmount.Int)
				unclaimedEthWei.Add(unclaimedEthWei, &intervalInfo.SmoothingPoolEthAmount.Int)
			}
		}

		// Get the block for the next rewards checkpoint
		header, err := collector.rp.Client.HeaderByNumber(context.Background(), nil)
		if err != nil {
			return fmt.Errorf("Error getting latest block header: %w", err)
		}

		collector.cumulativeRewards += eth.WeiToEth(newRewards)
		collector.cumulativeClaimedEthRewards += eth.WeiToEth(newClaimedEthRewards)
		unclaimedRplRewards = eth.WeiToEth(unclaimedRplWei)
		unclaimedEthRewards = eth.WeiToEth(unclaimedEthWei)
		collector.nextRewardsStartBlock = big.NewInt(0).Add(header.Number, big.NewInt(1))

		return nil
	})

	// get the beacon client sync status:
	wg.Go(func() error {
		progress := float64(0)

		syncStatus, err := collector.bc.GetSyncStatus()

		if err != nil {
			// NOTE: returning here causes the metric to not be emitted. the endpoint stays responsive, but also slightly more accurate (progress=nothing instead of 0)
			fmt.Printf("error getting beacon chain sync status: %w", err)
			return nil
		} else {
			progress = syncStatus.Progress
			if !syncStatus.Syncing {
				progress = 1.0
			}
		}
		// note this metric is emitted asynchronously, while others in this file tend to be emitted at the end of the outer function (mostly due to dependencies between metrics). See https://github.com/rocket-pool/smartnode/issues/186
		channel <- prometheus.MustNewConstMetric(
			collector.clientSyncProgress, prometheus.GaugeValue, progress, "beacon")
		return nil
	})

	// get the execution client sync status:
	wg.Go(func() error {
		syncStatus := collector.ec.CheckStatus(collector.cfg)
		// note this metric is emitted asynchronously, while others in this file tend to be emitted at the end of the outer function (mostly due to dependencies between metrics). See https://github.com/rocket-pool/smartnode/issues/186
		channel <- prometheus.MustNewConstMetric(
			collector.clientSyncProgress, prometheus.GaugeValue, syncStatus.PrimaryClientStatus.SyncProgress, "execution")
		return nil
	})

	// Get the number of active minipools on the node
	wg.Go(func() error {
		minipoolCount := len(minipools)
		for _, mpd := range minipools {
			if mpd.Finalised {
				minipoolCount--
			}
		}
		activeMinipoolCount = float64(minipoolCount)
		return nil
	})

	// Get the beacon head
	wg.Go(func() error {
		_beaconHead, err := collector.bc.GetBeaconHead()
		if err != nil {
			return fmt.Errorf("Error getting beacon chain head: %w", err)
		}
		beaconHead = _beaconHead
		return nil
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		collector.logError(err)
		return
	}

	// Calculate the node weight
	minCollateral := big.NewInt(0).Mul(eligibleBorrowedEth, state.NetworkDetails.MinCollateralFraction)
	minCollateral.Div(minCollateral, state.NetworkDetails.RplPrice)

	nodeWeight := big.NewInt(0)
	// The node must satisfy collateral requirements and have eligible ETH from which to earn rewards.
	if nd.RplStake.Cmp(minCollateral) != -1 && eligibleBorrowedEth.Sign() > 0 {
		nodeWeight = state.GetNodeWeight(eligibleBorrowedEth, nd.RplStake)
	}

	// Calculate the rewardable RPL
	reductionWindowStart := state.NetworkDetails.BondReductionWindowStart
	reductionWindowLength := state.NetworkDetails.BondReductionWindowLength
	reductionWindowEnd := reductionWindowStart + reductionWindowLength

	genesisTime := time.Unix(int64(state.BeaconConfig.GenesisTime), 0)
	secondsSinceGenesis := time.Duration(state.BeaconSlotNumber*state.BeaconConfig.SecondsPerSlot) * time.Second
	blockTime := genesisTime.Add(secondsSinceGenesis)

	zero := big.NewInt(0)
	pendingBorrowedEth := big.NewInt(0)
	pendingBondedEth := big.NewInt(0)
	rewardableBorrowedEth := big.NewInt(0)
	rewardableBondedEth := big.NewInt(0)
	for _, mpd := range minipools {
		if mpd.Finalised {
			// Ignore finalized minipools in the ratio math
			continue
		}
		bonded := big.NewInt(0)

		reduceBondTime := time.Unix(mpd.ReduceBondTime.Int64(), 0)
		timeSinceReductionStart := blockTime.Sub(reduceBondTime)
		if mpd.ReduceBondTime.Cmp(zero) == 0 ||
			mpd.ReduceBondCancelled ||
			timeSinceReductionStart > reductionWindowEnd {
			// No pending bond reduction
			bonded = mpd.NodeDepositBalance
		} else {
			// Pending bond reducton
			bonded.Set(mpd.ReduceBondValue)
		}
		borrowed := big.NewInt(0).Sub(eth.EthToWei(32), bonded)
		pendingBorrowedEth.Add(pendingBorrowedEth, borrowed)
		pendingBondedEth.Add(pendingBondedEth, bonded)

		validator, exists := state.ValidatorDetails[mpd.Pubkey]
		if !exists {
			// Validator doesn't exist on Beacon yet
			continue
		}
		if validator.ActivationEpoch > beaconHead.Epoch {
			// Validator hasn't activated yet
			continue
		}
		if validator.ExitEpoch <= beaconHead.Epoch {
			// Validator exited
			continue
		}

		rewardableBorrowedEth.Add(rewardableBorrowedEth, borrowed)
		rewardableBondedEth.Add(rewardableBondedEth, bonded)
	}

	// Calculate the "rewardable" minimum based on the Beacon Chain, including pending bond reductions
	rewardableMinimumStake := big.NewInt(0).Mul(rewardableBorrowedEth, state.NetworkDetails.MinCollateralFraction)
	rewardableMinimumStake.Div(rewardableMinimumStake, rplPriceRaw)

	// Calculate the "rewardable" maximum based on the Beacon Chain, including the pending bond reductions
	rewardableMaximumStake := big.NewInt(0).Mul(rewardableBondedEth, state.NetworkDetails.MaxCollateralFraction)
	rewardableMaximumStake.Div(rewardableMaximumStake, rplPriceRaw)

	// Calculate the actual "rewardable" amount
	rewardableRplStake := big.NewInt(0).Set(nd.RplStake)
	if rewardableRplStake.Cmp(rewardableMinimumStake) < 0 {
		rewardableRplStake.SetUint64(0)
	} else if rewardableRplStake.Cmp(rewardableMaximumStake) > 0 {
		rewardableRplStake.Set(rewardableMaximumStake)
	}
	rewardableStakeFloat := eth.WeiToEth(rewardableRplStake)

	// Calculate the estimated rewards
	rewardsIntervalDays := rewardsInterval.Seconds() / (60 * 60 * 24)
	inflationPerDay := eth.WeiToEth(inflationInterval)
	totalRplAtNextCheckpoint := (math.Pow(inflationPerDay, float64(rewardsIntervalDays)) - 1) * eth.WeiToEth(totalRplSupply)
	if totalRplAtNextCheckpoint < 0 {
		totalRplAtNextCheckpoint = 0
	}

	/*
	 * Calculates a RPIP-30 RPL reward estimate. Assumes that RPIP-30 has been fully phased in
	 *
	 * Formula:
	 * 		current_node_weight / (current_node_weight + previous_interval_total_node_weight) * estimated_collateral_rewards
	 *
	 * Note that if the node has no effective stake, has no eligibleBorrowedETH, or if this is the very first rewards
	 * period we don't attempt an estimate and simply use 0.
	 */
	estimatedRewards := float64(0)
	if totalEffectiveStake.Cmp(big.NewInt(0)) == 1 && nodeWeight.Cmp(big.NewInt(0)) == 1 && state.NetworkDetails.RewardIndex > 0 {

		nodeWeightSum := big.NewInt(0).Add(nodeWeight, previousIntervalTotalNodeWeight)

		// nodeWeightRatio = current_node_weight / (current_node_weight + previous_interval_total_node_weight)
		nodeWeightRatio, _ := big.NewFloat(0).Quo(
			big.NewFloat(0).SetInt(nodeWeight),
			big.NewFloat(0).SetInt(nodeWeightSum)).Float64()

		// estimatedRewards = nodeWeightRatio * estimated_collateral_rewards
		estimatedRewards = nodeWeightRatio * totalRplAtNextCheckpoint * nodeOperatorRewardsPercent
	}

	// Calculate the RPL APR
	rplApr := float64(0)
	if stakedRpl > 0 {
		rplApr = estimatedRewards / stakedRpl / rewardsInterval.Hours() * (24 * 365) * 100
	}

	// Calculate the total deposits and corresponding beacon chain balance share
	opts := &bind.CallOpts{
		BlockNumber: big.NewInt(0).SetUint64(state.ElBlockNumber),
	}
	minipoolDetails, err := eth2.GetBeaconBalancesFromState(collector.rp, minipools, state, beaconHead, opts)
	if err != nil {
		collector.logError(err)
		return
	}
	totalDepositBalance := float64(0)
	totalNodeShare := float64(0)
	totalBeaconBalance := float64(0)
	for _, minipool := range minipoolDetails {
		totalDepositBalance += eth.WeiToEth(minipool.NodeDeposit)
		totalNodeShare += eth.WeiToEth(minipool.NodeBalance)
		totalBeaconBalance += eth.WeiToEth(minipool.TotalBalance)
	}

	totalMinipoolBalance := float64(0)
	totalMinipoolShare := float64(0)
	totalRefundBalance := float64(0)
	for _, minipool := range minipools {
		totalMinipoolBalance += eth.WeiToEth(minipool.DistributableBalance)
		totalMinipoolShare += eth.WeiToEth(minipool.NodeShareOfBalance)
		totalRefundBalance += eth.WeiToEth(minipool.NodeRefundBalance)
	}

	// RPL collateral
	pendingBondedEthFloat := eth.WeiToEth(pendingBondedEth)
	var bondedCollateralRatio float64
	if pendingBondedEthFloat == 0 {
		bondedCollateralRatio = 0
	} else {
		bondedCollateralRatio = rplPrice * stakedRpl / pendingBondedEthFloat
	}

	pendingBorrowedEthFloat := eth.WeiToEth(pendingBorrowedEth)
	var borrowedCollateralRatio float64
	if pendingBorrowedEthFloat == 0 {
		borrowedCollateralRatio = 0
	} else {
		borrowedCollateralRatio = rplPrice * stakedRpl / pendingBorrowedEthFloat
	}

	// Update all the metrics
	channel <- prometheus.MustNewConstMetric(
		collector.totalStakedRpl, prometheus.GaugeValue, stakedRpl)
	channel <- prometheus.MustNewConstMetric(
		collector.effectiveStakedRpl, prometheus.GaugeValue, effectiveStakedRpl)
	channel <- prometheus.MustNewConstMetric(
		collector.rewardableStakedRpl, prometheus.GaugeValue, rewardableStakeFloat)
	channel <- prometheus.MustNewConstMetric(
		collector.cumulativeRplRewards, prometheus.GaugeValue, collector.cumulativeRewards)
	channel <- prometheus.MustNewConstMetric(
		collector.expectedRplRewards, prometheus.GaugeValue, estimatedRewards)
	channel <- prometheus.MustNewConstMetric(
		collector.rplApr, prometheus.GaugeValue, rplApr)
	channel <- prometheus.MustNewConstMetric(
		collector.balances, prometheus.GaugeValue, ethBalance, "ETH")
	channel <- prometheus.MustNewConstMetric(
		collector.balances, prometheus.GaugeValue, oldRplBalance, "Legacy RPL")
	channel <- prometheus.MustNewConstMetric(
		collector.balances, prometheus.GaugeValue, newRplBalance, "New RPL")
	channel <- prometheus.MustNewConstMetric(
		collector.balances, prometheus.GaugeValue, rethBalance, "rETH")
	channel <- prometheus.MustNewConstMetric(
		collector.activeMinipoolCount, prometheus.GaugeValue, activeMinipoolCount)
	channel <- prometheus.MustNewConstMetric(
		collector.depositedEth, prometheus.GaugeValue, totalDepositBalance)
	channel <- prometheus.MustNewConstMetric(
		collector.beaconShare, prometheus.GaugeValue, totalNodeShare)
	channel <- prometheus.MustNewConstMetric(
		collector.beaconBalance, prometheus.GaugeValue, totalBeaconBalance)
	channel <- prometheus.MustNewConstMetric(
		collector.minipoolBalance, prometheus.GaugeValue, totalMinipoolBalance)
	channel <- prometheus.MustNewConstMetric(
		collector.minipoolShare, prometheus.GaugeValue, totalMinipoolShare)
	channel <- prometheus.MustNewConstMetric(
		collector.refundBalance, prometheus.GaugeValue, totalRefundBalance)
	channel <- prometheus.MustNewConstMetric(
		collector.unclaimedRewards, prometheus.GaugeValue, unclaimedRplRewards)
	channel <- prometheus.MustNewConstMetric(
		collector.unclaimedEthRewards, prometheus.GaugeValue, unclaimedEthRewards)
	channel <- prometheus.MustNewConstMetric(
		collector.claimedEthRewards, prometheus.GaugeValue, collector.cumulativeClaimedEthRewards)
	channel <- prometheus.MustNewConstMetric(
		collector.borrowedCollateralRatio, prometheus.GaugeValue, borrowedCollateralRatio)
	channel <- prometheus.MustNewConstMetric(
		collector.bondedCollateralRatio, prometheus.GaugeValue, bondedCollateralRatio)
}

// Log error messages
func (collector *NodeCollector) logError(err error) {
	fmt.Printf("[%s] %s\n", collector.logPrefix, err.Error())
}
