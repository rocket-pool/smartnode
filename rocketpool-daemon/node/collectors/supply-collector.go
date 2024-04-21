package collectors

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/rocketpool-go/v2/types"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/v2/shared/keys"
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

	// The logger
	logger *slog.Logger

	// The thread-safe locker for the network state
	stateLocker *StateLocker
}

// Create a new PerformanceCollector instance
func NewSupplyCollector(logger *log.Logger, sp *services.ServiceProvider, stateLocker *StateLocker) *SupplyCollector {
	subsystem := "supply"
	sublogger := logger.With(slog.String(keys.TaskKey, "Supply Collector"))
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
		logger:      sublogger,
		stateLocker: stateLocker,
	}
}

// Write metric descriptions to the Prometheus channel
func (c *SupplyCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- c.nodeCount
	channel <- c.nodeFee
	channel <- c.minipoolCount
	channel <- c.totalMinipools
	channel <- c.activeMinipools
}

// Collect the latest metric values and pass them to Prometheus
func (c *SupplyCollector) Collect(channel chan<- prometheus.Metric) {
	// Get the latest state
	state := c.stateLocker.GetState()
	if state == nil {
		return
	}

	// Sync
	nodeFee := state.NetworkDetails.NodeFee
	initializedCount := float64(-1)
	prelaunchCount := float64(-1)
	stakingCount := float64(-1)
	dissolvedCount := float64(-1)
	finalizedCount := float64(-1)

	// Get total number of Rocket Pool nodes
	nodeCount := float64(len(state.NodeDetails))

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
		c.nodeCount, prometheus.GaugeValue, nodeCount)
	channel <- prometheus.MustNewConstMetric(
		c.nodeFee, prometheus.GaugeValue, nodeFee)
	channel <- prometheus.MustNewConstMetric(
		c.minipoolCount, prometheus.GaugeValue, initializedCount, "initialized")
	channel <- prometheus.MustNewConstMetric(
		c.minipoolCount, prometheus.GaugeValue, prelaunchCount, "prelaunch")
	channel <- prometheus.MustNewConstMetric(
		c.minipoolCount, prometheus.GaugeValue, stakingCount, "staking")
	channel <- prometheus.MustNewConstMetric(
		c.minipoolCount, prometheus.GaugeValue, dissolvedCount, "dissolved")
	channel <- prometheus.MustNewConstMetric(
		c.minipoolCount, prometheus.GaugeValue, finalizedCount, "finalized")

	// Set the total and active count
	totalMinipoolCount := initializedCount + prelaunchCount + stakingCount + dissolvedCount + finalizedCount
	activeMinipoolCount := totalMinipoolCount - finalizedCount
	channel <- prometheus.MustNewConstMetric(
		c.totalMinipools, prometheus.GaugeValue, totalMinipoolCount)
	channel <- prometheus.MustNewConstMetric(
		c.activeMinipools, prometheus.GaugeValue, activeMinipoolCount)
}
