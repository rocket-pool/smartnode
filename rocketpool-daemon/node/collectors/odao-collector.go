package collectors

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/v2/shared/keys"
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

	// The logger
	logger *slog.Logger

	// The thread-safe locker for the network state
	stateLocker *StateLocker
}

// Create a new OdaoCollector instance
func NewOdaoCollector(logger *log.Logger, sp *services.ServiceProvider, stateLocker *StateLocker) *OdaoCollector {
	subsystem := "odao"
	sublogger := logger.With(slog.String(keys.RoutineKey, "ODAO Collector"))
	return &OdaoCollector{
		currentEth1Block: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "current_eth1_block"),
			"The latest block reported by the Execution client at the time of collecting the metrics",
			nil, nil,
		),
		pricesBlock: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "prices_block"),
			"The Execution block that was used when reporting the latest prices",
			nil, nil,
		),
		effectiveRplStakeBlock: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "effective_rpl_stake_block"),
			"The Execution block where the Effective RPL Stake was last updated",
			nil, nil,
		),
		latestReportableBlock: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "latest_reportable_block"),
			"The latest Execution block where network prices were reportable by the ODAO",
			nil, nil,
		),
		sp:          sp,
		logger:      sublogger,
		stateLocker: stateLocker,
	}
}

// Write metric descriptions to the Prometheus channel
func (c *OdaoCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- c.currentEth1Block
	channel <- c.pricesBlock
	channel <- c.effectiveRplStakeBlock
	channel <- c.latestReportableBlock
}

// Collect the latest metric values and pass them to Prometheus
func (c *OdaoCollector) Collect(channel chan<- prometheus.Metric) {
	// Get the latest state
	state := c.stateLocker.GetState()
	if state == nil {
		return
	}

	blockNumberFloat := float64(state.ElBlockNumber)
	pricesBlockFloat := float64(state.NetworkDetails.PricesBlock)
	effectiveRplStakeBlockFloat := pricesBlockFloat
	channel <- prometheus.MustNewConstMetric(
		c.currentEth1Block, prometheus.GaugeValue, blockNumberFloat)
	channel <- prometheus.MustNewConstMetric(
		c.pricesBlock, prometheus.GaugeValue, pricesBlockFloat)
	channel <- prometheus.MustNewConstMetric(
		c.effectiveRplStakeBlock, prometheus.GaugeValue, effectiveRplStakeBlockFloat)
}
