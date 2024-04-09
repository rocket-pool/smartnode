package collectors

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/v2/shared/keys"
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

	// The Smartnode service provider
	sp *services.ServiceProvider

	// The logger
	logger *slog.Logger

	// The thread-safe locker for the network state
	stateLocker *StateLocker
}

// Create a new RplCollector instance
func NewRplCollector(logger *log.Logger, sp *services.ServiceProvider, stateLocker *StateLocker) *RplCollector {
	subsystem := "rpl"
	sublogger := logger.With(slog.String(keys.RoutineKey, "RPL Collector"))
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
		sp:          sp,
		logger:      sublogger,
		stateLocker: stateLocker,
	}
}

// Write metric descriptions to the Prometheus channel
func (c *RplCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- c.rplPrice
	channel <- c.totalValueStaked
	channel <- c.totalEffectiveStaked
	channel <- c.checkpointTime
}

// Collect the latest metric values and pass them to Prometheus
func (c *RplCollector) Collect(channel chan<- prometheus.Metric) {
	// Get the latest state
	state := c.stateLocker.GetState()
	if state == nil {
		return
	}

	rplPriceFloat := eth.WeiToEth(state.NetworkDetails.RplPrice)
	totalValueStakedFloat := eth.WeiToEth(state.NetworkDetails.TotalRPLStake)
	totalEffectiveStake := c.stateLocker.GetTotalEffectiveRPLStake()
	lastCheckpoint := state.NetworkDetails.IntervalStart
	rewardsInterval := state.NetworkDetails.IntervalDuration
	nextRewardsTime := float64(lastCheckpoint.Add(rewardsInterval).Unix()) * 1000
	if totalEffectiveStake == nil {
		return
	}

	channel <- prometheus.MustNewConstMetric(
		c.rplPrice, prometheus.GaugeValue, rplPriceFloat)
	channel <- prometheus.MustNewConstMetric(
		c.totalValueStaked, prometheus.GaugeValue, totalValueStakedFloat)
	channel <- prometheus.MustNewConstMetric(
		c.totalEffectiveStaked, prometheus.GaugeValue, eth.WeiToEth(totalEffectiveStake))
	channel <- prometheus.MustNewConstMetric(
		c.checkpointTime, prometheus.GaugeValue, nextRewardsTime)
}
