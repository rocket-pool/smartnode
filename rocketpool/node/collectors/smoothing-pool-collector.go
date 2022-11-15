package collectors

import (
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/rocketpool/api/node"
	"github.com/rocket-pool/smartnode/shared/services"
	"golang.org/x/sync/errgroup"
)

// Represents the collector for Smoothing Pool metrics
type SmoothingPoolCollector struct {
	// the ETH balance on the smoothing pool
	ethBalanceOnSmoothingPool *prometheus.Desc

	// The Rocket Pool contract manager
	rp *rocketpool.RocketPool

	// The EC client
	ec *services.ExecutionClientManager

	// The node address
	nodeAddress common.Address
}

// Create a new SmoothingPoolCollector instance
func NewSmoothingPoolCollector(rp *rocketpool.RocketPool, ec *services.ExecutionClientManager) *SmoothingPoolCollector {
	subsystem := "smoothing_pool"
	return &SmoothingPoolCollector{
		ethBalanceOnSmoothingPool: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "eth_balance"),
			"The ETH balance on the smoothing pool",
			nil, nil,
		),
		rp: rp,
		ec: ec,
	}
}

// Write metric descriptions to the Prometheus channel
func (collector *SmoothingPoolCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- collector.ethBalanceOnSmoothingPool
}

// Collect the latest metric values and pass them to Prometheus
func (collector *SmoothingPoolCollector) Collect(channel chan<- prometheus.Metric) {

	// Sync
	var wg errgroup.Group
	ethBalanceOnSmoothingPool := float64(0)

	// Get the ETH balance in the smoothing pool
	wg.Go(func() error {
		balanceResponse, err := node.GetSmoothingPoolBalance(collector.rp, collector.ec)
		if err != nil {
			return fmt.Errorf("Error getting smoothing pool balance: %w", err)
		}
		ethBalanceOnSmoothingPool = eth.WeiToEth(balanceResponse.EthBalance)

		return nil
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		log.Printf("%s\n", err.Error())
		return
	}

	channel <- prometheus.MustNewConstMetric(
		collector.ethBalanceOnSmoothingPool, prometheus.GaugeValue, ethBalanceOnSmoothingPool)
}
