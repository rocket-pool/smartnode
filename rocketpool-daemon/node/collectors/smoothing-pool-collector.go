package collectors

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/services"
)

// Represents the collector for Smoothing Pool metrics
type SmoothingPoolCollector struct {
	// the ETH balance on the smoothing pool
	ethBalanceOnSmoothingPool *prometheus.Desc

	// The Smartnode service provider
	sp *services.ServiceProvider

	// The thread-safe locker for the network state
	stateLocker *StateLocker

	// Prefix for logging
	logPrefix string
}

// Create a new SmoothingPoolCollector instance
func NewSmoothingPoolCollector(sp *services.ServiceProvider, stateLocker *StateLocker) *SmoothingPoolCollector {
	subsystem := "smoothing_pool"
	return &SmoothingPoolCollector{
		ethBalanceOnSmoothingPool: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "eth_balance"),
			"The ETH balance on the smoothing pool",
			nil, nil,
		),
		sp:          sp,
		stateLocker: stateLocker,
		logPrefix:   "SP Collector",
	}
}

// Write metric descriptions to the Prometheus channel
func (collector *SmoothingPoolCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- collector.ethBalanceOnSmoothingPool
}

// Collect the latest metric values and pass them to Prometheus
func (collector *SmoothingPoolCollector) Collect(channel chan<- prometheus.Metric) {
	// Get the latest state
	state := collector.stateLocker.GetState()
	if state == nil {
		return
	}

	ethBalanceOnSmoothingPool := eth.WeiToEth(state.NetworkDetails.SmoothingPoolBalance)

	channel <- prometheus.MustNewConstMetric(
		collector.ethBalanceOnSmoothingPool, prometheus.GaugeValue, ethBalanceOnSmoothingPool)
}

// Log error messages
func (collector *SmoothingPoolCollector) logError(err error) {
	fmt.Printf("[%s] %s\n", collector.logPrefix, err.Error())
}
