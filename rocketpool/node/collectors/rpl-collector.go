package collectors

import (
	"fmt"
	"math/big"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	v110_node "github.com/rocket-pool/rocketpool-go/legacy/v1.1.0/node"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"golang.org/x/sync/errgroup"
)

// Represents the collector for the RPL metrics
type RplCollector struct {
	// The RPL price (in terms of ETH)
	rplPrice *prometheus.Desc

	// The total amount of RPL staked on the network
	totalValueStaked *prometheus.Desc

	// The total effective amount of RPL staked on the network
	totalEffectiveStaked *prometheus.Desc

	// The date and time of the next RPL rewards checkpoint
	checkpointTime *prometheus.Desc

	// The Rocket Pool config
	cfg *config.RocketPoolConfig

	// The Rocket Pool contract manager
	rp *rocketpool.RocketPool

	// The manager for the network state in Atlas
	m *state.NetworkStateManager

	// Prefix for logging
	logPrefix string
}

// Create a new RplCollector instance
func NewRplCollector(rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, m *state.NetworkStateManager) *RplCollector {
	subsystem := "rpl"
	return &RplCollector{
		rplPrice: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "rpl_price"),
			"The RPL price (in terms of ETH)",
			nil, nil,
		),
		totalValueStaked: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "total_value_staked"),
			"The total amount of RPL staked on the network",
			nil, nil,
		),
		totalEffectiveStaked: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "total_effective_staked"),
			"The total effective amount of RPL staked on the network",
			nil, nil,
		),
		checkpointTime: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "checkpoint_time"),
			"The date and time of the next RPL rewards checkpoint",
			nil, nil,
		),
		rp:        rp,
		cfg:       cfg,
		m:         m,
		logPrefix: "RPL Collector",
	}
}

// Write metric descriptions to the Prometheus channel
func (collector *RplCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- collector.rplPrice
	channel <- collector.totalValueStaked
	channel <- collector.totalEffectiveStaked
	channel <- collector.checkpointTime
}

// Collect the latest metric values and pass them to Prometheus
func (collector *RplCollector) Collect(channel chan<- prometheus.Metric) {
	latestState := collector.m.GetLatestState()
	if latestState == nil {
		collector.collectImpl_Legacy(channel)
	} else {
		collector.collectImpl_Atlas(latestState, channel)
	}
}

// Collect the latest metric values and pass them to Prometheus
func (collector *RplCollector) collectImpl_Legacy(channel chan<- prometheus.Metric) {

	// Sync
	var wg errgroup.Group
	rplPriceFloat := float64(-1)
	totalValueStakedFloat := float64(-1)
	totalEffectiveStakedFloat := float64(-1)
	var lastCheckpoint time.Time
	var rewardsInterval time.Duration

	// Get the RPL price (in terms of ETH)
	wg.Go(func() error {
		rplPrice, err := network.GetRPLPrice(collector.rp, nil)
		if err != nil {
			return fmt.Errorf("Error getting RPL price: %w", err)
		}

		rplPriceFloat = eth.WeiToEth(rplPrice)
		return nil
	})

	// Get the total amount of RPL staked on the network
	wg.Go(func() error {
		totalValueStaked, err := node.GetTotalRPLStake(collector.rp, nil)
		if err != nil {
			return fmt.Errorf("Error getting total amount of RPL staked on the network: %w", err)
		}

		totalValueStakedFloat = eth.WeiToEth(totalValueStaked)
		return nil
	})

	// Get the total effective amount of RPL staked on the network
	wg.Go(func() error {
		legacyNodeStakingAddress := collector.cfg.Smartnode.GetV110NodeStakingAddress()
		totalEffectiveStaked, err := v110_node.GetTotalEffectiveRPLStake(collector.rp, nil, &legacyNodeStakingAddress)
		if err != nil {
			return fmt.Errorf("Error getting total effective amount of RPL staked on the network: %w", err)
		}

		totalEffectiveStakedFloat = eth.WeiToEth(totalEffectiveStaked)
		return nil
	})

	// Get the start of the rewards checkpoint
	wg.Go(func() error {
		_lastCheckpoint, err := rewards.GetClaimIntervalTimeStart(collector.rp, nil)
		if err != nil {
			return fmt.Errorf("Error getting the previous rewards claim time: %w", err)
		}

		lastCheckpoint = _lastCheckpoint
		return err
	})

	// Get the rewards checkpoint interval
	wg.Go(func() error {
		_rewardsInterval, err := rewards.GetClaimIntervalTime(collector.rp, nil)
		if err != nil {
			return fmt.Errorf("Error getting the rewards interval: %w", err)
		}

		rewardsInterval = _rewardsInterval
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		collector.logError(err)
		return
	}

	nextRewardsTime := float64(lastCheckpoint.Add(rewardsInterval).Unix()) * 1000

	channel <- prometheus.MustNewConstMetric(
		collector.rplPrice, prometheus.GaugeValue, rplPriceFloat)
	channel <- prometheus.MustNewConstMetric(
		collector.totalValueStaked, prometheus.GaugeValue, totalValueStakedFloat)
	channel <- prometheus.MustNewConstMetric(
		collector.totalEffectiveStaked, prometheus.GaugeValue, totalEffectiveStakedFloat)
	channel <- prometheus.MustNewConstMetric(
		collector.checkpointTime, prometheus.GaugeValue, nextRewardsTime)
}

// Collect the latest metric values and pass them to Prometheus
func (collector *RplCollector) collectImpl_Atlas(state *state.NetworkState, channel chan<- prometheus.Metric) {

	rplPriceFloat := eth.WeiToEth(state.NetworkDetails.RplPrice)
	totalValueStakedFloat := eth.WeiToEth(state.NetworkDetails.TotalRPLStake)
	var totalEffectiveStake *big.Int
	lastCheckpoint := state.NetworkDetails.IntervalStart
	rewardsInterval := state.NetworkDetails.IntervalDuration

	_totalEffectiveStake := big.NewInt(0)
	for _, node := range state.NodeDetails {
		_totalEffectiveStake.Add(_totalEffectiveStake, node.EffectiveRPLStake)
	}
	totalEffectiveStake = _totalEffectiveStake

	nextRewardsTime := float64(lastCheckpoint.Add(rewardsInterval).Unix()) * 1000

	channel <- prometheus.MustNewConstMetric(
		collector.rplPrice, prometheus.GaugeValue, rplPriceFloat)
	channel <- prometheus.MustNewConstMetric(
		collector.totalValueStaked, prometheus.GaugeValue, totalValueStakedFloat)
	channel <- prometheus.MustNewConstMetric(
		collector.totalEffectiveStaked, prometheus.GaugeValue, eth.WeiToEth(totalEffectiveStake))
	channel <- prometheus.MustNewConstMetric(
		collector.checkpointTime, prometheus.GaugeValue, nextRewardsTime)
}

// Log error messages
func (collector *RplCollector) logError(err error) {
	fmt.Printf("[%s] %s\n", collector.logPrefix, err.Error())
}
