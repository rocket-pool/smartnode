package collectors

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/config"
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

	// The thread-safe locker for the network state
	stateLocker *StateLocker

	// Prefix for logging
	logPrefix string
}

// Create a new RplCollector instance
func NewRplCollector(rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, stateLocker *StateLocker) *RplCollector {
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
		rp:          rp,
		cfg:         cfg,
		stateLocker: stateLocker,
		logPrefix:   "RPL Collector",
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
	// Get the latest state
	state := collector.stateLocker.GetState()
	if state == nil {
		return
	}

	rplPriceFloat := eth.WeiToEth(state.NetworkDetails.RplPrice)
	totalValueStakedFloat := eth.WeiToEth(state.NetworkDetails.TotalRPLStake)
	totalEffectiveStake := collector.stateLocker.GetTotalEffectiveRPLStake()
	lastCheckpoint := state.NetworkDetails.IntervalStart
	rewardsInterval := state.NetworkDetails.IntervalDuration
	nextRewardsTime := float64(lastCheckpoint.Add(rewardsInterval).Unix()) * 1000
	if totalEffectiveStake == nil {
		return
	}

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
