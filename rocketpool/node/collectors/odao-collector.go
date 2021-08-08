package collectors

import (
	"context"
	"log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Represents the collector for the ODAO metrics
type OdaoCollector struct {
	// The latest block reported by the ETH1 client at the time of collecting the metrics
	currentEth1Block 		*prometheus.Desc

	// The ETH1 block that was used when reporting the latest prices
	pricesBlock	            *prometheus.Desc

	// The ETH1 block where the Effective RPL Stake was last updated
	effectiveRplStakeBlock  *prometheus.Desc

	// The latest ETH1 block where network prices were reportable by the ODAO
	latestReportableBlock   *prometheus.Desc

	// The Rocket Pool contract manager
	rp 						*rocketpool.RocketPool
}


// Create a new DemandCollector instance
func NewOdaoCollector(rp *rocketpool.RocketPool) *OdaoCollector {
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
		rp: rp,
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
 
	// Get the latest block reported by the ETH1 client
    blockNumber, err := collector.rp.Client.BlockNumber(context.Background())
    blockNumberFloat := float64(blockNumber)
    if err != nil {
		log.Printf("Error getting latest ETH1 block: %s", err)
		blockNumberFloat = -1
    }
    channel <- prometheus.MustNewConstMetric(
		collector.currentEth1Block, prometheus.GaugeValue, blockNumberFloat)
	
	// Get ETH1 block that was used when reporting the latest prices
	pricesBlock, err := network.GetPricesBlock(collector.rp, nil)
	pricesBlockFloat := float64(pricesBlock)
	if err != nil {
		log.Printf("Error getting ETH1 prices block: %s", err)
		pricesBlockFloat = -1
	}
    channel <- prometheus.MustNewConstMetric(
		collector.pricesBlock, prometheus.GaugeValue, pricesBlockFloat)

	// Get the ETH1 block where the Effective RPL Stake was last updated
	effectiveRplStakeBlock, err := network.GetPricesBlock(collector.rp, nil)
	effectiveRplStakeBlockFloat := float64(effectiveRplStakeBlock)
	if err != nil {
		log.Printf("Error getting ETH1 effective RPL stake block: %s", err)
		effectiveRplStakeBlockFloat = -1
	}
    channel <- prometheus.MustNewConstMetric(
		collector.effectiveRplStakeBlock, prometheus.GaugeValue, effectiveRplStakeBlockFloat)

    // Get the latest ETH1 block where network prices were reportable by the ODAO
    latestReportableBlock, err := network.GetLatestReportablePricesBlock(collector.rp, nil)
    latestReportableBlockFloat := float64(-1) 
    if err != nil {
        log.Printf("Error getting ETH1 latest reportable block: %s", err)
    } else {
        latestReportableBlockFloat = float64(latestReportableBlock.Uint64())
    }
    channel <- prometheus.MustNewConstMetric(
        collector.latestReportableBlock, prometheus.GaugeValue, latestReportableBlockFloat)
}
