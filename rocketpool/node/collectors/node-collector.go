package collectors

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/tokens"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
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

	// The RPL collateral level for the node
	rplCollateral *prometheus.Desc

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

	// The RPL rewards from the last period that have not been claimed yet
	unclaimedRewards *prometheus.Desc

	// The claimed ETH rewards from the smoothing pool
	claimedEthRewards *prometheus.Desc

	// The unclaimed ETH rewards from the smoothing pool
	unclaimedEthRewards *prometheus.Desc

	// The Rocket Pool contract manager
	rp *rocketpool.RocketPool

	// The beacon client
	bc beacon.Client

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
}

// Create a new NodeCollector instance
func NewNodeCollector(rp *rocketpool.RocketPool, bc beacon.Client, nodeAddress common.Address, cfg *config.RocketPoolConfig) *NodeCollector {

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
		rplCollateral: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "rpl_collateral"),
			"The RPL collateral level for the node",
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
		rp:               rp,
		bc:               bc,
		nodeAddress:      nodeAddress,
		eventLogInterval: big.NewInt(int64(eventLogInterval)),
		handledIntervals: map[uint64]bool{},
		cfg:              cfg,
	}
}

// Write metric descriptions to the Prometheus channel
func (collector *NodeCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- collector.totalStakedRpl
	channel <- collector.effectiveStakedRpl
	channel <- collector.cumulativeRplRewards
	channel <- collector.expectedRplRewards
	channel <- collector.rplApr
	channel <- collector.balances
	channel <- collector.activeMinipoolCount
	channel <- collector.depositedEth
	channel <- collector.beaconShare
	channel <- collector.unclaimedRewards
	channel <- collector.claimedEthRewards
	channel <- collector.unclaimedEthRewards
}

// Collect the latest metric values and pass them to Prometheus
func (collector *NodeCollector) Collect(channel chan<- prometheus.Metric) {

	// Sync
	var wg errgroup.Group
	stakedRpl := float64(0)
	effectiveStakedRpl := float64(0)
	var rewardsInterval time.Duration
	var inflationInterval *big.Int
	var totalRplSupply *big.Int
	var totalEffectiveStake *big.Int
	var nodeOperatorRewardsPercent float64
	ethBalance := float64(0)
	oldRplBalance := float64(0)
	newRplBalance := float64(0)
	rethBalance := float64(0)
	var activeMinipoolCount float64
	var rplPrice float64
	collateralRatio := float64(0)
	var addresses []common.Address
	var beaconHead beacon.BeaconHead
	unclaimedEthRewards := float64(0)
	unclaimedRplRewards := float64(0)

	// Get the total staked RPL
	wg.Go(func() error {
		stakedRplWei, err := node.GetNodeRPLStake(collector.rp, collector.nodeAddress, nil)
		if err != nil {
			return fmt.Errorf("Error getting total staked RPL: %w", err)
		}

		stakedRpl = eth.WeiToEth(stakedRplWei)
		return nil
	})

	// Get the effective staked RPL
	wg.Go(func() error {
		effectiveStakedRplWei, err := node.GetNodeEffectiveRPLStake(collector.rp, collector.nodeAddress, nil)
		if err != nil {
			return fmt.Errorf("Error getting effective staked RPL: %w", err)
		}

		effectiveStakedRpl = eth.WeiToEth(effectiveStakedRplWei)
		return nil
	})

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

		// Get the info for each claimed interval
		for _, claimedInterval := range claimed {
			_, exists := collector.handledIntervals[claimedInterval]
			if !exists {
				intervalInfo, err := rprewards.GetIntervalInfo(collector.rp, collector.cfg, collector.nodeAddress, claimedInterval)
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
			intervalInfo, err := rprewards.GetIntervalInfo(collector.rp, collector.cfg, collector.nodeAddress, unclaimedInterval)
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

	// Get the rewards checkpoint interval
	wg.Go(func() error {
		_rewardsInterval, err := rewards.GetClaimIntervalTime(collector.rp, nil)
		if err != nil {
			return fmt.Errorf("Error getting rewards checkpoint interval: %w", err)
		}
		rewardsInterval = _rewardsInterval
		return nil
	})

	// Get the RPL inflation interval
	wg.Go(func() error {
		_inflationInterval, err := tokens.GetRPLInflationIntervalRate(collector.rp, nil)
		if err != nil {
			return fmt.Errorf("Error getting RPL inflation interval: %w", err)
		}
		inflationInterval = _inflationInterval
		return nil
	})

	// Get the total RPL supply
	wg.Go(func() error {
		_totalRplSupply, err := tokens.GetRPLTotalSupply(collector.rp, nil)
		if err != nil {
			return fmt.Errorf("Error getting total RPL supply: %w", err)
		}
		totalRplSupply = _totalRplSupply
		return nil
	})

	// Get the total network effective stake
	wg.Go(func() error {
		_totalEffectiveStake, err := node.GetTotalEffectiveRPLStake(collector.rp, nil)
		if err != nil {
			return fmt.Errorf("Error getting total network effective stake: %w", err)
		}
		totalEffectiveStake = _totalEffectiveStake
		return nil
	})

	// Get the node operator rewards percent
	wg.Go(func() error {
		_nodeOperatorRewardsPercent, err := rewards.GetNodeOperatorRewardsPercent(collector.rp, nil)
		if err != nil {
			return fmt.Errorf("Error getting node operator rewards percent: %w", err)
		}
		nodeOperatorRewardsPercent = eth.WeiToEth(_nodeOperatorRewardsPercent)
		return nil
	})

	// Get the node balances
	wg.Go(func() error {
		balances, err := tokens.GetBalances(collector.rp, collector.nodeAddress, nil)
		if err != nil {
			return fmt.Errorf("Error getting node balances: %w", err)
		}
		ethBalance = eth.WeiToEth(balances.ETH)
		oldRplBalance = eth.WeiToEth(balances.FixedSupplyRPL)
		newRplBalance = eth.WeiToEth(balances.RPL)
		rethBalance = eth.WeiToEth(balances.RETH)
		return nil
	})

	// Get the number of active minipools on the node
	wg.Go(func() error {
		_activeMinipoolCount, err := minipool.GetNodeActiveMinipoolCount(collector.rp, collector.nodeAddress, nil)
		if err != nil {
			return fmt.Errorf("Error getting node active minipool count: %w", err)
		}
		activeMinipoolCount = float64(_activeMinipoolCount)
		return nil
	})

	// Get the RPL price
	wg.Go(func() error {
		rplPriceWei, err := network.GetRPLPrice(collector.rp, nil)
		if err != nil {
			return fmt.Errorf("Error getting RPL price: %w", err)
		}
		rplPrice = eth.WeiToEth(rplPriceWei)
		return nil
	})

	// Get the list of minipool addresses for this node
	wg.Go(func() error {
		_addresses, err := minipool.GetNodeMinipoolAddresses(collector.rp, collector.nodeAddress, nil)
		if err != nil {
			return fmt.Errorf("Error getting node minipool addresses: %w", err)
		}
		addresses = _addresses
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
		log.Printf("%s\n", err.Error())
		return
	}

	// Calculate the estimated rewards
	rewardsIntervalDays := rewardsInterval.Seconds() / (60 * 60 * 24)
	inflationPerDay := eth.WeiToEth(inflationInterval)
	totalRplAtNextCheckpoint := (math.Pow(inflationPerDay, float64(rewardsIntervalDays)) - 1) * eth.WeiToEth(totalRplSupply)
	if totalRplAtNextCheckpoint < 0 {
		totalRplAtNextCheckpoint = 0
	}
	estimatedRewards := float64(0)
	if totalEffectiveStake.Cmp(big.NewInt(0)) == 1 {
		estimatedRewards = effectiveStakedRpl / eth.WeiToEth(totalEffectiveStake) * totalRplAtNextCheckpoint * nodeOperatorRewardsPercent
	}

	// Calculate the RPL APR
	rplApr := estimatedRewards / stakedRpl / rewardsInterval.Hours() * (24 * 365) * 100

	// Calculate the collateral ratio
	if activeMinipoolCount > 0 {
		collateralRatio = rplPrice * stakedRpl / (activeMinipoolCount * 16.0)
	}

	// Calculate the total deposits and corresponding beacon chain balance share
	minipoolDetails, err := eth2.GetBeaconBalances(collector.rp, collector.bc, addresses, beaconHead, nil)
	if err != nil {
		log.Printf("%s\n", err.Error())
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

	// Update all the metrics
	channel <- prometheus.MustNewConstMetric(
		collector.totalStakedRpl, prometheus.GaugeValue, stakedRpl)
	channel <- prometheus.MustNewConstMetric(
		collector.effectiveStakedRpl, prometheus.GaugeValue, effectiveStakedRpl)
	channel <- prometheus.MustNewConstMetric(
		collector.rplCollateral, prometheus.GaugeValue, collateralRatio)
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
		collector.unclaimedRewards, prometheus.GaugeValue, unclaimedRplRewards)
	channel <- prometheus.MustNewConstMetric(
		collector.unclaimedEthRewards, prometheus.GaugeValue, unclaimedEthRewards)
	channel <- prometheus.MustNewConstMetric(
		collector.claimedEthRewards, prometheus.GaugeValue, collector.cumulativeClaimedEthRewards)
}
