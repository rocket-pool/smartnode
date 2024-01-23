package collectors

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/services"
)

// Represents the collector for the Supply metrics
type SupplyCollector struct {
	// The total number of Rocket Pool nodes
	nodeCount *prometheus.Desc

	// The current commission rate for new minipools
	nodeFee *prometheus.Desc

	// The count of Rocket Pool minipools, broken down by status
	minipoolCount *prometheus.Desc

	// The total number of Rocket Pool minipools
	totalMinipools *prometheus.Desc

	// The number of active (non-finalized) Rocket Pool minipools
	activeMinipools *prometheus.Desc

	// The Smartnode service provider
	sp *services.ServiceProvider

	// The thread-safe locker for the network state
	stateLocker *StateLocker

	// Prefix for logging
	logPrefix string
}

// Create a new PerformanceCollector instance
func NewSupplyCollector(sp *services.ServiceProvider, stateLocker *StateLocker) *SupplyCollector {
	subsystem := "supply"
	return &SupplyCollector{
		nodeCount: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "node_count"),
			"The total number of Rocket Pool nodes",
			nil, nil,
		),
		nodeFee: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "node_fee"),
			"The current commission rate for new minipools",
			nil, nil,
		),
		minipoolCount: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "minipool_count"),
			"The count of Rocket Pool minipools, broken down by status",
			[]string{"status"}, nil,
		),
		totalMinipools: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "total_minipools"),
			"The total number of Rocket Pool minipools",
			nil, nil,
		),
		activeMinipools: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "active_minipools"),
			"The number of active (non-finalized) Rocket Pool minipools",
			nil, nil,
		),
		sp:          sp,
		stateLocker: stateLocker,
		logPrefix:   "Supply Collector",
	}
}

// Write metric descriptions to the Prometheus channel
func (collector *SupplyCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- collector.nodeCount
	channel <- collector.nodeFee
	channel <- collector.minipoolCount
	channel <- collector.totalMinipools
	channel <- collector.activeMinipools
}

// Collect the latest metric values and pass them to Prometheus
func (collector *SupplyCollector) Collect(channel chan<- prometheus.Metric) {
	// Get the latest state
	state := collector.stateLocker.GetState()
	if state == nil {
		return
	}

	// Sync
	nodeCount := float64(-1)
	nodeFee := state.NetworkDetails.NodeFee
	initializedCount := float64(-1)
	prelaunchCount := float64(-1)
	stakingCount := float64(-1)
	dissolvedCount := float64(-1)
	finalizedCount := float64(-1)

	// Get total number of Rocket Pool nodes
	nodeCount = float64(len(state.NodeDetails))

	// Get the total number of Rocket Pool minipools
	for _, mpd := range state.MinipoolDetails {
		if mpd.Finalised {
			finalizedCount++
		} else {
			switch mpd.Status {
			case types.MinipoolStatus_Initialized:
				initializedCount++
			case types.MinipoolStatus_Prelaunch:
				prelaunchCount++
			case types.MinipoolStatus_Staking:
				stakingCount++
			case types.MinipoolStatus_Dissolved:
				dissolvedCount++
			}
		}
	}

	channel <- prometheus.MustNewConstMetric(
		collector.nodeCount, prometheus.GaugeValue, nodeCount)
	channel <- prometheus.MustNewConstMetric(
		collector.nodeFee, prometheus.GaugeValue, nodeFee)
	channel <- prometheus.MustNewConstMetric(
		collector.minipoolCount, prometheus.GaugeValue, initializedCount, "initialized")
	channel <- prometheus.MustNewConstMetric(
		collector.minipoolCount, prometheus.GaugeValue, prelaunchCount, "prelaunch")
	channel <- prometheus.MustNewConstMetric(
		collector.minipoolCount, prometheus.GaugeValue, stakingCount, "staking")
	channel <- prometheus.MustNewConstMetric(
		collector.minipoolCount, prometheus.GaugeValue, dissolvedCount, "dissolved")
	channel <- prometheus.MustNewConstMetric(
		collector.minipoolCount, prometheus.GaugeValue, finalizedCount, "finalized")

	// Set the total and active count
	totalMinipoolCount := initializedCount + prelaunchCount + stakingCount + dissolvedCount + finalizedCount
	activeMinipoolCount := totalMinipoolCount - finalizedCount
	channel <- prometheus.MustNewConstMetric(
		collector.totalMinipools, prometheus.GaugeValue, totalMinipoolCount)
	channel <- prometheus.MustNewConstMetric(
		collector.activeMinipools, prometheus.GaugeValue, activeMinipoolCount)
}

// Log error messages
func (collector *SupplyCollector) logError(err error) {
	fmt.Printf("[%s] %s\n", collector.logPrefix, err.Error())
}
