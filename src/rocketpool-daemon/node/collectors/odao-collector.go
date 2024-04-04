package collectors

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/services"
)

// Represents the collector for the ODAO metrics
type OdaoCollector struct {
	// The latest block reported by the ETH1 client at the time of collecting the metrics
	currentEth1Block *prometheus.Desc

	// The ETH1 block that was used when reporting the latest prices
	pricesBlock *prometheus.Desc

	// The ETH1 block where the Effective RPL Stake was last updated
	effectiveRplStakeBlock *prometheus.Desc

	// The latest ETH1 block where network prices were reportable by the ODAO
	latestReportableBlock *prometheus.Desc

	// The Smartnode service provider
	sp *services.ServiceProvider

	// The thread-safe locker for the network state
	stateLocker *StateLocker

	// Prefix for logging
	logPrefix string
}

// Create a new DemandCollector instance
func NewOdaoCollector(sp *services.ServiceProvider, stateLocker *StateLocker) *OdaoCollector {
	subsystem := "odao"
	return &OdaoCollector{
		currentEth1Block: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "current_eth1_block"),
			"The latest block reported by the ETH1 client at the time of collecting the metrics",
			nil, nil,
		),
		pricesBlock: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "prices_block"),
			"The ETH1 block that was used when reporting the latest prices",
			nil, nil,
		),
		effectiveRplStakeBlock: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "effective_rpl_stake_block"),
			"The ETH1 block where the Effective RPL Stake was last updated",
			nil, nil,
		),
		latestReportableBlock: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "latest_reportable_block"),
			"The latest ETH1 block where network prices were reportable by the ODAO",
			nil, nil,
		),
		sp:          sp,
		stateLocker: stateLocker,
		logPrefix:   "ODAO Collector",
	}
}

// Write metric descriptions to the Prometheus channel
func (collector *OdaoCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- collector.currentEth1Block
	channel <- collector.pricesBlock
	channel <- collector.effectiveRplStakeBlock
	channel <- collector.latestReportableBlock
}

// Collect the latest metric values and pass them to Prometheus
func (collector *OdaoCollector) Collect(channel chan<- prometheus.Metric) {
	// Get the latest state
	state := collector.stateLocker.GetState()
	if state == nil {
		return
	}

	blockNumberFloat := float64(state.ElBlockNumber)
	pricesBlockFloat := float64(state.NetworkDetails.PricesBlock)
	effectiveRplStakeBlockFloat := pricesBlockFloat
	channel <- prometheus.MustNewConstMetric(
		collector.currentEth1Block, prometheus.GaugeValue, blockNumberFloat)
	channel <- prometheus.MustNewConstMetric(
		collector.pricesBlock, prometheus.GaugeValue, pricesBlockFloat)
	channel <- prometheus.MustNewConstMetric(
		collector.effectiveRplStakeBlock, prometheus.GaugeValue, effectiveRplStakeBlockFloat)
}

// Log error messages
func (collector *OdaoCollector) logError(err error) {
	fmt.Printf("[%s] %s\n", collector.logPrefix, err.Error())
}
