package collectors

import (
	"context"
	"fmt"
	"log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"golang.org/x/sync/errgroup"
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

	// The Rocket Pool contract manager
	rp *rocketpool.RocketPool
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

	// Sync
	var wg errgroup.Group
	blockNumberFloat := float64(-1)
	pricesBlockFloat := float64(-1)
	effectiveRplStakeBlockFloat := float64(-1)
	latestReportableBlockFloat := float64(-1)

	// Get the latest block reported by the ETH1 client
	wg.Go(func() error {
		blockNumber, err := collector.rp.Client.BlockNumber(context.Background())
		if err != nil {
			return fmt.Errorf("Error getting latest ETH1 block: %w", err)
		}

		blockNumberFloat = float64(blockNumber)
		return nil
	})

	// Get ETH1 block that was used when reporting the latest prices
	wg.Go(func() error {
		pricesBlock, err := network.GetPricesBlock(collector.rp, nil)
		if err != nil {
			return fmt.Errorf("Error getting ETH1 prices block: %w", err)
		}

		pricesBlockFloat = float64(pricesBlock)
		effectiveRplStakeBlockFloat = float64(pricesBlock)
		return nil
	})

	// Get the latest ETH1 block where network prices were reportable by the ODAO
	wg.Go(func() error {
		latestReportableBlock, err := network.GetLatestReportablePricesBlock(collector.rp, nil)
		if err != nil {
			return fmt.Errorf("Error getting ETH1 latest reportable block: %w", err)
		}

		latestReportableBlockFloat = float64(latestReportableBlock.Uint64())
		return nil
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		log.Printf("%s\n", err.Error())
		return
	}

	channel <- prometheus.MustNewConstMetric(
		collector.currentEth1Block, prometheus.GaugeValue, blockNumberFloat)
	channel <- prometheus.MustNewConstMetric(
		collector.pricesBlock, prometheus.GaugeValue, pricesBlockFloat)
	channel <- prometheus.MustNewConstMetric(
		collector.effectiveRplStakeBlock, prometheus.GaugeValue, effectiveRplStakeBlockFloat)
	channel <- prometheus.MustNewConstMetric(
		collector.latestReportableBlock, prometheus.GaugeValue, latestReportableBlockFloat)

}
