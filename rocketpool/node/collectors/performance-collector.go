package collectors

import (
	"context"
	"log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/tokens"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
)

// Represents the collector for the Performance metrics
type PerformanceCollector struct {
	// The ETH utilization rate (%)
	ethUtilizationRate 		*prometheus.Desc

	// The total amount of ETH staked
	totalStakingBalanceEth	*prometheus.Desc

	// The ETH / rETH ratio
	ethRethExchangeRate 	*prometheus.Desc

	// The total amount of ETH locked (TVL)
	totalValueLockedEth     *prometheus.Desc

	// The total rETH supply
	totalRethSupply      	*prometheus.Desc

	// The ETH balance of the rETH contract address
	rethContractBalance     *prometheus.Desc

	// The Rocket Pool contract manager
	rp 						*rocketpool.RocketPool
}


// Create a new PerformanceCollector instance
func NewPerformanceCollector(rp *rocketpool.RocketPool) *PerformanceCollector {
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
		rp: rp,
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
 
	// Get the ETH utilization rate
	ethUtilizationRate, err := network.GetETHUtilizationRate(collector.rp, nil)
	if err != nil {
		log.Printf("Error getting ETH utilization rate: %s", err)
		ethUtilizationRate = -1
	}
    channel <- prometheus.MustNewConstMetric(
		collector.ethUtilizationRate, prometheus.GaugeValue, ethUtilizationRate)
	
	// Get the total ETH staking balance
	totalStakingBalance, err := network.GetStakingETHBalance(collector.rp, nil)
	balanceFloat := float64(-1)
	if err != nil {
		log.Printf("Error getting total ETH staking balance: %s", err)
	} else {
		balanceFloat = eth.WeiToEth(totalStakingBalance)
	}
    channel <- prometheus.MustNewConstMetric(
		collector.totalStakingBalanceEth, prometheus.GaugeValue, balanceFloat)
	
    // Get the ETH-rETH exchange rate
    exchangeRate, err := tokens.GetRETHExchangeRate(collector.rp, nil)
    if err != nil {
        log.Printf("Error getting ETH-rETH exchange rate: %s", err)
        exchangeRate = -1
    }
    channel <- prometheus.MustNewConstMetric(
        collector.ethRethExchangeRate, prometheus.GaugeValue, exchangeRate)
	
    // Get the total ETH balance (TVL)
    tvl, err := network.GetTotalETHBalance(collector.rp, nil)
	tvlFloat := float64(-1)
    if err != nil {
        log.Printf("Error getting total ETH balance (TVL): %s", err)
    } else {
		tvlFloat = eth.WeiToEth(tvl)
	}
    channel <- prometheus.MustNewConstMetric(
        collector.totalValueLockedEth, prometheus.GaugeValue, tvlFloat)

	// Get the ETH balance of the rETH contract
	rETHContract, err := collector.rp.GetContract("rocketTokenRETH")
	rETHBalance := float64(-1)
	if err != nil {
        log.Printf("Error getting ETH balance of rETH staking contract: %s", err)
	} else {
		balance, err := collector.rp.Client.BalanceAt(
			context.Background(), *rETHContract.Address, nil)
		if err != nil {
			log.Printf("Error getting ETH balance of rETH staking contract: %s", err)
		} else {
			rETHBalance = eth.WeiToEth(balance)
		}
	}
	channel <- prometheus.MustNewConstMetric(
		collector.rethContractBalance, prometheus.GaugeValue, rETHBalance)

    // Get the total rETH supply
    totalRethSupply, err := tokens.GetRETHTotalSupply(collector.rp, nil)
    rethFloat := float64(-1)
    if err != nil {
        log.Printf("Error getting total rETH supply: %s", err)
    } else {
		rethFloat = eth.WeiToEth(totalRethSupply)
	}
    channel <- prometheus.MustNewConstMetric(
        collector.totalRethSupply, prometheus.GaugeValue, rethFloat)

}
