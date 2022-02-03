package collectors

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "rocketpool"

// Represents the collector for the scrub check metrics
type ScrubCollector struct {

	// The total number of minipools in prelaunch status
	prelaunchMinipoolsDesc *prometheus.Desc

	// The number of minipools that had verified good credentials on the Beacon Chain
	beaconPassesDesc *prometheus.Desc

	// The number of minipools that had invalid withdrawal credentials on the Beacon Chain, and were scrubbed
	beaconScrubsDesc *prometheus.Desc

	// The number of minipools that had a good prestake signature
	prestakePassesDesc *prometheus.Desc

	// The number of minipools that had an invalid prestake signature and were scrubbed
	prestakeScrubsDesc *prometheus.Desc

	// The number of minipools that passed the deposit contract checks
	depositContractPassesDesc *prometheus.Desc

	// The number of minipools that had invalid / malicious deposits or credentials in the deposit contract and were scrubbed
	depositContractScrubsDesc *prometheus.Desc

	// The number of minipools without any deposits in the deposit contract
	poolsWithoutDepositsDesc *prometheus.Desc

	// The number of minipools that weren't covered by any of the other checks
	uncoveredMinipoolsDesc *prometheus.Desc

	// The number of minipools that were scrubbed for safety because they failed the sanity checks
	safetyScrubsDesc *prometheus.Desc

	// The time of the latest block that the check was run against
	latestBlockTimeDesc *prometheus.Desc

	// Counters
	TotalMinipools        float64
	GoodOnBeaconCount     float64
	BadOnBeaconCount      float64
	GoodPrestakeCount     float64
	BadPrestakeCount      float64
	GoodOnDepositContract float64
	BadOnDepositContract  float64
	DepositlessMinipools  float64
	UncoveredMinipools    float64
	SafetyScrubs          float64
	LatestBlockTime       float64

	// Mutex
	UpdateLock sync.Mutex
}

// Create a new ScrubCollector instance
func NewScrubCollector() *ScrubCollector {
	subsystem := "scrub"
	return &ScrubCollector{
		prelaunchMinipoolsDesc: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "prelaunch_minipools"),
			"The total number of minipools in prelaunch status",
			nil, nil,
		),
		beaconPassesDesc: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "beacon_passes"),
			"The number of minipools that had verified good credentials on the Beacon Chain",
			nil, nil,
		),
		beaconScrubsDesc: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "beacon_scrubs"),
			"The number of minipools that had invalid withdrawal credentials on the Beacon Chain, and were scrubbed",
			nil, nil,
		),
		prestakePassesDesc: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "prestake_passes"),
			"The number of minipools that had a good prestake signature",
			nil, nil,
		),
		prestakeScrubsDesc: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "prestake_scrubs"),
			"The number of minipools that had an invalid prestake signature and were scrubbed",
			nil, nil,
		),
		depositContractPassesDesc: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "deposit_contract_passes"),
			"The number of minipools that passed the deposit contract checks",
			nil, nil,
		),
		depositContractScrubsDesc: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "deposit_contract_scrubs"),
			"The number of minipools that had invalid / malicious deposits or credentials in the deposit contract and were scrubbed",
			nil, nil,
		),
		poolsWithoutDepositsDesc: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "pools_without_deposits"),
			"The number of minipools without any deposits in the deposit contract",
			nil, nil,
		),
		uncoveredMinipoolsDesc: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "uncovered_minipools"),
			"The number of minipools that weren't covered by any of the other checks",
			nil, nil,
		),
		safetyScrubsDesc: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "safety_scrubs"),
			"The number of minipools that were scrubbed for safety because they failed the sanity checks",
			nil, nil,
		),
		latestBlockTimeDesc: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "latest_block_time"),
			"The time of the latest block that the check was run against",
			nil, nil,
		),
	}
}

// Write metric descriptions to the Prometheus channel
func (collector *ScrubCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- collector.prelaunchMinipoolsDesc
	channel <- collector.beaconPassesDesc
	channel <- collector.beaconScrubsDesc
	channel <- collector.prestakePassesDesc
	channel <- collector.prestakeScrubsDesc
	channel <- collector.depositContractPassesDesc
	channel <- collector.depositContractScrubsDesc
	channel <- collector.poolsWithoutDepositsDesc
	channel <- collector.uncoveredMinipoolsDesc
	channel <- collector.safetyScrubsDesc
}

// Collect the latest metric values and pass them to Prometheus
func (collector *ScrubCollector) Collect(channel chan<- prometheus.Metric) {

	// Sync
	collector.UpdateLock.Lock()
	defer collector.UpdateLock.Unlock()

	// Update all of the metrics
	channel <- prometheus.MustNewConstMetric(
		collector.prelaunchMinipoolsDesc, prometheus.GaugeValue, collector.TotalMinipools)
	channel <- prometheus.MustNewConstMetric(
		collector.beaconPassesDesc, prometheus.GaugeValue, collector.GoodOnBeaconCount)
	channel <- prometheus.MustNewConstMetric(
		collector.beaconScrubsDesc, prometheus.GaugeValue, collector.BadOnBeaconCount)
	channel <- prometheus.MustNewConstMetric(
		collector.prestakePassesDesc, prometheus.GaugeValue, collector.GoodPrestakeCount)
	channel <- prometheus.MustNewConstMetric(
		collector.prestakeScrubsDesc, prometheus.GaugeValue, collector.BadPrestakeCount)
	channel <- prometheus.MustNewConstMetric(
		collector.depositContractPassesDesc, prometheus.GaugeValue, collector.GoodOnDepositContract)
	channel <- prometheus.MustNewConstMetric(
		collector.depositContractScrubsDesc, prometheus.GaugeValue, collector.BadOnDepositContract)
	channel <- prometheus.MustNewConstMetric(
		collector.poolsWithoutDepositsDesc, prometheus.GaugeValue, collector.DepositlessMinipools)
	channel <- prometheus.MustNewConstMetric(
		collector.uncoveredMinipoolsDesc, prometheus.GaugeValue, collector.UncoveredMinipools)
	channel <- prometheus.MustNewConstMetric(
		collector.safetyScrubsDesc, prometheus.GaugeValue, collector.SafetyScrubs)
	channel <- prometheus.MustNewConstMetric(
		collector.latestBlockTimeDesc, prometheus.GaugeValue, collector.LatestBlockTime)

}
