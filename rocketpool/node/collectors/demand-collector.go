package collectors

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
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

	// The Rocket Pool contract manager
	rp *rocketpool.RocketPool

	// The thread-safe locker for the network state
	stateLocker *StateLocker

	// Prefix for logging
	logPrefix string
}

// Create a new DemandCollector instance
func NewDemandCollector(rp *rocketpool.RocketPool, stateLocker *StateLocker) *DemandCollector {
	subsystem := "demand"
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
		rp:          rp,
		stateLocker: stateLocker,
		logPrefix:   "Demand Collector",
	}
}

// Write metric descriptions to the Prometheus channel
func (collector *DemandCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- collector.depositPoolBalance
	channel <- collector.depositPoolExcess
	channel <- collector.totalMinipoolCapacity
	channel <- collector.effectiveMinipoolCapacity
}

// Collect the latest metric values and pass them to Prometheus
func (collector *DemandCollector) Collect(channel chan<- prometheus.Metric) {
	// Get the latest state
	state := collector.stateLocker.GetState()

	balanceFloat := eth.WeiToEth(state.NetworkDetails.DepositPoolBalance)
	excessFloat := eth.WeiToEth(state.NetworkDetails.DepositPoolExcess)
	totalFloat := eth.WeiToEth(state.NetworkDetails.QueueCapacity.Total)
	effectiveFloat := eth.WeiToEth(state.NetworkDetails.QueueCapacity.Effective)

	channel <- prometheus.MustNewConstMetric(
		collector.depositPoolBalance, prometheus.GaugeValue, balanceFloat)
	channel <- prometheus.MustNewConstMetric(
		collector.depositPoolExcess, prometheus.GaugeValue, excessFloat)
	channel <- prometheus.MustNewConstMetric(
		collector.totalMinipoolCapacity, prometheus.GaugeValue, totalFloat)
	channel <- prometheus.MustNewConstMetric(
		collector.effectiveMinipoolCapacity, prometheus.GaugeValue, effectiveFloat)
}

// Log error messages
func (collector *DemandCollector) logError(err error) {
	fmt.Printf("[%s] %s\n", collector.logPrefix, err.Error())
}
