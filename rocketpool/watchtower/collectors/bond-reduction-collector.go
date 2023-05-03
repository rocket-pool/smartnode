package collectors

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

// Represents the collector for the bond reduction check metrics
type BondReductionCollector struct {

	// The total number of minipools waiting for bond reduction
	totalMinipoolsDesc *prometheus.Desc

	// The number of bond reductions cancelled due to balance too low
	balanceTooLowDesc *prometheus.Desc

	// The number of bond reductions cancelled due to being in an invalid state
	invalidStateDesc *prometheus.Desc

	// The time of the latest block that the check was run against
	latestBlockTimeDesc *prometheus.Desc

	// Counters
	TotalMinipools  float64
	BalanceTooLow   float64
	InvalidState    float64
	LatestBlockTime float64

	// Mutex
	UpdateLock *sync.Mutex
}

// Create a new ScrubCollector instance
func NewBondReductionCollector() *BondReductionCollector {
	subsystem := "bond_reduction"
	return &BondReductionCollector{
		totalMinipoolsDesc: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "total_minipools"),
			"The total number of minipools waiting for bond reduction",
			nil, nil,
		),
		balanceTooLowDesc: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "balance_too_low"),
			"The number of bond reductions cancelled due to balance too low",
			nil, nil,
		),
		invalidStateDesc: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "invalid_state"),
			"The number of bond reductions cancelled due to being in an invalid state",
			nil, nil,
		),
		latestBlockTimeDesc: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "latest_block_time"),
			"The time of the latest block that the check was run against",
			nil, nil,
		),
		UpdateLock: &sync.Mutex{},
	}
}

// Write metric descriptions to the Prometheus channel
func (collector *BondReductionCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- collector.totalMinipoolsDesc
	channel <- collector.balanceTooLowDesc
	channel <- collector.invalidStateDesc
	channel <- collector.latestBlockTimeDesc
}

// Collect the latest metric values and pass them to Prometheus
func (collector *BondReductionCollector) Collect(channel chan<- prometheus.Metric) {

	// Sync
	collector.UpdateLock.Lock()
	defer collector.UpdateLock.Unlock()

	// Update all of the metrics
	channel <- prometheus.MustNewConstMetric(
		collector.totalMinipoolsDesc, prometheus.GaugeValue, collector.TotalMinipools)
	channel <- prometheus.MustNewConstMetric(
		collector.balanceTooLowDesc, prometheus.GaugeValue, collector.BalanceTooLow)
	channel <- prometheus.MustNewConstMetric(
		collector.invalidStateDesc, prometheus.GaugeValue, collector.InvalidState)
	channel <- prometheus.MustNewConstMetric(
		collector.latestBlockTimeDesc, prometheus.GaugeValue, collector.LatestBlockTime)
}
