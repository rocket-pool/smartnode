package collectors

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/v2/shared/keys"
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

	// The logger
	logger *slog.Logger

	// The thread-safe locker for the network state
	stateLocker *StateLocker
}

// Create a new PerformanceCollector instance
func NewPerformanceCollector(logger *log.Logger, sp *services.ServiceProvider, stateLocker *StateLocker) *PerformanceCollector {
	subsystem := "performance"
	sublogger := logger.With(slog.String(keys.RoutineKey, "Performance Collector"))
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
		logger:      sublogger,
		stateLocker: stateLocker,
	}
}

// Write metric descriptions to the Prometheus channel
func (c *PerformanceCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- c.ethUtilizationRate
	channel <- c.totalStakingBalanceEth
	channel <- c.ethRethExchangeRate
	channel <- c.totalValueLockedEth
	channel <- c.rethContractBalance
	channel <- c.totalRethSupply
}

// Collect the latest metric values and pass them to Prometheus
func (c *PerformanceCollector) Collect(channel chan<- prometheus.Metric) {
	// Get the latest state
	state := c.stateLocker.GetState()
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
		c.ethUtilizationRate, prometheus.GaugeValue, ethUtilizationRate)
	channel <- prometheus.MustNewConstMetric(
		c.totalStakingBalanceEth, prometheus.GaugeValue, balanceFloat)
	channel <- prometheus.MustNewConstMetric(
		c.ethRethExchangeRate, prometheus.GaugeValue, exchangeRate)
	channel <- prometheus.MustNewConstMetric(
		c.totalValueLockedEth, prometheus.GaugeValue, tvlFloat)
	channel <- prometheus.MustNewConstMetric(
		c.rethContractBalance, prometheus.GaugeValue, rETHBalance)
	channel <- prometheus.MustNewConstMetric(
		c.totalRethSupply, prometheus.GaugeValue, rethFloat)
}
