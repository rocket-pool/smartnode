package collectors

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/rocketpool/api/node"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/state"
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

	// The manager for the network state in Atlas
	m *state.NetworkStateManager

	// Prefix for logging
	logPrefix string
}

// Create a new SmoothingPoolCollector instance
func NewSmoothingPoolCollector(rp *rocketpool.RocketPool, ec *services.ExecutionClientManager, m *state.NetworkStateManager) *SmoothingPoolCollector {
	subsystem := "smoothing_pool"
	return &SmoothingPoolCollector{
		ethBalanceOnSmoothingPool: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "eth_balance"),
			"The ETH balance on the smoothing pool",
			nil, nil,
		),
		rp:        rp,
		ec:        ec,
		m:         m,
		logPrefix: "SP Collector",
	}
}

// Write metric descriptions to the Prometheus channel
func (collector *SmoothingPoolCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- collector.ethBalanceOnSmoothingPool
}

// Collect the latest metric values and pass them to Prometheus
func (collector *SmoothingPoolCollector) Collect(channel chan<- prometheus.Metric) {
	latestState := collector.m.GetLatestState()
	if latestState == nil {
		collector.collectImpl_Legacy(channel)
	} else {
		collector.collectImpl_Atlas(latestState, channel)
	}
}

// Collect the latest metric values and pass them to Prometheus
func (collector *SmoothingPoolCollector) collectImpl_Legacy(channel chan<- prometheus.Metric) {

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
		collector.logError(err)
		return
	}

	channel <- prometheus.MustNewConstMetric(
		collector.ethBalanceOnSmoothingPool, prometheus.GaugeValue, ethBalanceOnSmoothingPool)
}

// Collect the latest metric values and pass them to Prometheus
func (collector *SmoothingPoolCollector) collectImpl_Atlas(state *state.NetworkState, channel chan<- prometheus.Metric) {

	ethBalanceOnSmoothingPool := eth.WeiToEth(state.NetworkDetails.SmoothingPoolBalance)

	channel <- prometheus.MustNewConstMetric(
		collector.ethBalanceOnSmoothingPool, prometheus.GaugeValue, ethBalanceOnSmoothingPool)
}

// Log error messages
func (collector *SmoothingPoolCollector) logError(err error) {
	fmt.Printf("[%s] %s\n", collector.logPrefix, err.Error())
}
