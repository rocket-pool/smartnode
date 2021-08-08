package collectors

import (
	"log"
	"math/big"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Represents the collector for the Supply metrics
type SupplyCollector struct {
	// The total number of Rocket Pool nodes
	nodeCount 		*prometheus.Desc

	// The current commission rate for new minipools
	nodeFee	        *prometheus.Desc

	// The count of Rocket Pool minipools, broken down by status
	minipoolCount 	*prometheus.Desc

	// The total number of Rocket Pool minipools
	totalMinipools 	*prometheus.Desc

	// The number of active (non-finalized) Rocket Pool minipools
	activeMinipools	*prometheus.Desc

	// The breakdown of nodes by their registered timezone
	nodeTimezones 	*prometheus.Desc

	// The Rocket Pool contract manager
	rp 				*rocketpool.RocketPool
}


// Create a new PerformanceCollector instance
func NewSupplyCollector(rp *rocketpool.RocketPool) *SupplyCollector {
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
		nodeTimezones: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "node_timezones"),
			"The breakdown of nodes by their registered timezone",
			[]string{"timezone"}, nil,
		),
		rp: rp,
	}
}


// Write metric descriptions to the Prometheus channel
func (collector *SupplyCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- collector.nodeCount
	channel <- collector.nodeFee
	channel <- collector.minipoolCount
	channel <- collector.totalMinipools
	channel <- collector.activeMinipools
	channel <- collector.nodeTimezones
}


// Collect the latest metric values and pass them to Prometheus
func (collector *SupplyCollector) Collect(channel chan<- prometheus.Metric) {
 
	// Get total number of Rocket Pool nodes
	nodeCountUint, err := node.GetNodeCount(collector.rp, nil)
    nodeCount := float64(nodeCountUint)
	if err != nil {
		log.Printf("Error getting total number of Rocket Pool nodes: %s", err)
		nodeCount = -1
	}
    channel <- prometheus.MustNewConstMetric(
		collector.nodeCount, prometheus.GaugeValue, nodeCount)
	
	// Get the current commission rate for new minipools
	nodeFee, err := network.GetNodeFee(collector.rp, nil)
	if err != nil {
		log.Printf("Error getting current commission rate for new minipools: %s", err)
		nodeFee = -1
	}
    channel <- prometheus.MustNewConstMetric(
		collector.nodeFee, prometheus.GaugeValue, nodeFee)
	
    // Get the total number of Rocket Pool minipools
    minipoolCounts, err := minipool.GetMinipoolCountPerStatus(collector.rp, 0, 0, nil)
	initializedCount := float64(-1)
	prelaunchCount := float64(-1)
	stakingCount := float64(-1)
	withdrawableCount := float64(-1)
	dissolvedCount := float64(-1)
	finalizedCount := float64(-1)
    if err != nil {
        log.Printf("Error getting total number of Rocket Pool minipools: %s", err)
    } else {
		initializedCount = float64(minipoolCounts.Initialized.Uint64())
		prelaunchCount = float64(minipoolCounts.Prelaunch.Uint64())
		stakingCount = float64(minipoolCounts.Staking.Uint64())
		withdrawableCount = float64(minipoolCounts.Withdrawable.Uint64())
		dissolvedCount = float64(minipoolCounts.Dissolved.Uint64())
	}
	finalizedCountUint, err := minipool.GetFinalisedMinipoolCount(collector.rp, nil)
	if err != nil {
        log.Printf("Error getting total number of finalized minipools: %s", err)
	} else {
		finalizedCount = float64(finalizedCountUint)
		withdrawableCount -= finalizedCount
	}

    channel <- prometheus.MustNewConstMetric(
        collector.minipoolCount, prometheus.GaugeValue, initializedCount, "initialized")
	channel <- prometheus.MustNewConstMetric(
        collector.minipoolCount, prometheus.GaugeValue, prelaunchCount, "prelaunch")
    channel <- prometheus.MustNewConstMetric(
        collector.minipoolCount, prometheus.GaugeValue, stakingCount, "staking")
    channel <- prometheus.MustNewConstMetric(
        collector.minipoolCount, prometheus.GaugeValue, withdrawableCount, "withdrawable")
    channel <- prometheus.MustNewConstMetric(
        collector.minipoolCount, prometheus.GaugeValue, dissolvedCount, "dissolved")
	channel <- prometheus.MustNewConstMetric(
		collector.minipoolCount, prometheus.GaugeValue, finalizedCount, "finalized")

	// Set the total and active count
	totalMinipoolCount := initializedCount + prelaunchCount + stakingCount + withdrawableCount + dissolvedCount + finalizedCount
	activeMinipoolCount := totalMinipoolCount - finalizedCount
	channel <- prometheus.MustNewConstMetric(
		collector.totalMinipools, prometheus.GaugeValue, totalMinipoolCount)
	channel <- prometheus.MustNewConstMetric(
		collector.activeMinipools, prometheus.GaugeValue, activeMinipoolCount)

	zero := big.NewInt(0)
	timezoneCounts, err := node.GetNodeCountPerTimezone(collector.rp, zero, zero, nil)
	if err != nil {
        log.Printf("Error getting timezone counts: %s", err)
	} else {
		for _, timezoneCount := range timezoneCounts {
			channel <- prometheus.MustNewConstMetric(
				collector.nodeTimezones, prometheus.GaugeValue, float64(timezoneCount.Count.Uint64()), timezoneCount.Timezone)
		}
	}

}
