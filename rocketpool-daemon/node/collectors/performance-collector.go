package collectors

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/services"
)

// Represents the collector for the Performance metrics
type PerformanceCollector struct {
	// The ETH utilization rate (%)
	ethUtilizationRate *prometheus.Desc

	// The total amount of ETH staked
	totalStakingBalanceEth *prometheus.Desc

	// The ETH / rETH ratio
	ethRethExchangeRate *prometheus.Desc

	// The total amount of ETH locked (TVL)
	totalValueLockedEth *prometheus.Desc

	// The total rETH supply
	totalRethSupply *prometheus.Desc

	// The ETH balance of the rETH contract address
	rethContractBalance *prometheus.Desc

	// The Smartnode service provider
	sp *services.ServiceProvider

	// The thread-safe locker for the network state
	stateLocker *StateLocker

	// Prefix for logging
	logPrefix string
}

// Create a new PerformanceCollector instance
func NewPerformanceCollector(sp *services.ServiceProvider, stateLocker *StateLocker) *PerformanceCollector {
	subsystem := "performance"
	return &PerformanceCollector{
		ethUtilizationRate: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "eth_utilization_rate"),
			"The ETH utilization rate (%)",
			nil, nil,
		),
		totalStakingBalanceEth: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "total_staking_balance_eth"),
			"The total amount of ETH staked",
			nil, nil,
		),
		ethRethExchangeRate: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "eth_reth_exchange_rate"),
			"The ETH / rETH ratio",
			nil, nil,
		),
		totalValueLockedEth: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "total_value_locked_eth"),
			"The total amount of ETH locked (TVL)",
			nil, nil,
		),
		rethContractBalance: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "reth_contract_balance"),
			"The ETH balance of the rETH contract address",
			nil, nil,
		),
		totalRethSupply: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "total_reth_supply"),
			"The total rETH supply",
			nil, nil,
		),
		sp:          sp,
		stateLocker: stateLocker,
		logPrefix:   "Performance Collector",
	}
}

// Write metric descriptions to the Prometheus channel
func (collector *PerformanceCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- collector.ethUtilizationRate
	channel <- collector.totalStakingBalanceEth
	channel <- collector.ethRethExchangeRate
	channel <- collector.totalValueLockedEth
	channel <- collector.rethContractBalance
	channel <- collector.totalRethSupply
}

// Collect the latest metric values and pass them to Prometheus
func (collector *PerformanceCollector) Collect(channel chan<- prometheus.Metric) {
	// Get the latest state
	state := collector.stateLocker.GetState()
	if state == nil {
		return
	}

	ethUtilizationRate := state.NetworkDetails.ETHUtilizationRate
	balanceFloat := eth.WeiToEth(state.NetworkDetails.StakingETHBalance)
	exchangeRate := state.NetworkDetails.RETHExchangeRate
	tvlFloat := eth.WeiToEth(state.NetworkDetails.TotalETHBalance)
	rETHBalance := eth.WeiToEth(state.NetworkDetails.RETHBalance)
	rethFloat := eth.WeiToEth(state.NetworkDetails.TotalRETHSupply)

	channel <- prometheus.MustNewConstMetric(
		collector.ethUtilizationRate, prometheus.GaugeValue, ethUtilizationRate)
	channel <- prometheus.MustNewConstMetric(
		collector.totalStakingBalanceEth, prometheus.GaugeValue, balanceFloat)
	channel <- prometheus.MustNewConstMetric(
		collector.ethRethExchangeRate, prometheus.GaugeValue, exchangeRate)
	channel <- prometheus.MustNewConstMetric(
		collector.totalValueLockedEth, prometheus.GaugeValue, tvlFloat)
	channel <- prometheus.MustNewConstMetric(
		collector.rethContractBalance, prometheus.GaugeValue, rETHBalance)
	channel <- prometheus.MustNewConstMetric(
		collector.totalRethSupply, prometheus.GaugeValue, rethFloat)
}

// Log error messages
func (collector *PerformanceCollector) logError(err error) {
	fmt.Printf("[%s] %s\n", collector.logPrefix, err.Error())
}
