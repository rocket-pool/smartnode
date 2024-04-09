package collectors

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/v2/shared/keys"
)

const namespace = "rocketpool"

// Represents the collector for the Demand metrics
type DemandCollector struct {
	// The amount of ETH currently in the Deposit Pool
	depositPoolBalance *prometheus.Desc

	// The excess ETH balance of the Deposit Pool
	depositPoolExcess *prometheus.Desc

	// The total ETH capacity of the Minipool queue
	totalMinipoolCapacity *prometheus.Desc

	// The effective ETH capacity of the Minipool queue
	effectiveMinipoolCapacity *prometheus.Desc

	// The number of minipools currently in the queue
	queueLength *prometheus.Desc

	// The Smartnode service provider
	sp *services.ServiceProvider

	// The logger
	logger *slog.Logger

	// The thread-safe locker for the network state
	stateLocker *StateLocker
}

// Create a new DemandCollector instance
func NewDemandCollector(logger *log.Logger, sp *services.ServiceProvider, stateLocker *StateLocker) *DemandCollector {
	subsystem := "demand"
	sublogger := logger.With(slog.String(keys.RoutineKey, "Demand Collector"))
	return &DemandCollector{
		depositPoolBalance: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "deposit_pool_balance"),
			"The amount of ETH currently in the Deposit Pool",
			nil, nil,
		),
		depositPoolExcess: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "deposit_pool_excess"),
			"The excess ETH balance of the Deposit Pool",
			nil, nil,
		),
		totalMinipoolCapacity: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "total_minipool_capacity"),
			"The total ETH capacity of the Minipool queue",
			nil, nil,
		),
		effectiveMinipoolCapacity: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "effective_minipool_capacity"),
			"The effective ETH capacity of the Minipool queue",
			nil, nil,
		),
		queueLength: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "queue_length"),
			"The number of minipools currently in the queue",
			nil, nil,
		),
		sp:          sp,
		logger:      sublogger,
		stateLocker: stateLocker,
	}
}

// Write metric descriptions to the Prometheus channel
func (c *DemandCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- c.depositPoolBalance
	channel <- c.depositPoolExcess
	channel <- c.totalMinipoolCapacity
	channel <- c.effectiveMinipoolCapacity
	channel <- c.queueLength
}

// Collect the latest metric values and pass them to Prometheus
func (c *DemandCollector) Collect(channel chan<- prometheus.Metric) {
	// Get the latest state
	state := c.stateLocker.GetState()
	if state == nil {
		return
	}

	balanceFloat := eth.WeiToEth(state.NetworkDetails.DepositPoolBalance)
	excessFloat := eth.WeiToEth(state.NetworkDetails.DepositPoolExcess)
	totalFloat := eth.WeiToEth(state.NetworkDetails.TotalQueueCapacity)
	effectiveFloat := eth.WeiToEth(state.NetworkDetails.EffectiveQueueCapacity)
	queueLength := float64(state.NetworkDetails.QueueLength.Uint64())

	channel <- prometheus.MustNewConstMetric(
		c.depositPoolBalance, prometheus.GaugeValue, balanceFloat)
	channel <- prometheus.MustNewConstMetric(
		c.depositPoolExcess, prometheus.GaugeValue, excessFloat)
	channel <- prometheus.MustNewConstMetric(
		c.totalMinipoolCapacity, prometheus.GaugeValue, totalFloat)
	channel <- prometheus.MustNewConstMetric(
		c.effectiveMinipoolCapacity, prometheus.GaugeValue, effectiveFloat)
	channel <- prometheus.MustNewConstMetric(
		c.queueLength, prometheus.GaugeValue, queueLength)
}
