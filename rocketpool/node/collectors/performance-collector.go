package collectors

import (
	"context"
	"fmt"
	"log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/tokens"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"golang.org/x/sync/errgroup"
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

	// The Rocket Pool contract manager
	rp *rocketpool.RocketPool
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

	// Sync
	var wg errgroup.Group
	ethUtilizationRate := float64(-1)
	balanceFloat := float64(-1)
	exchangeRate := float64(-1)
	tvlFloat := float64(-1)
	rETHBalance := float64(-1)
	rethFloat := float64(-1)

	// Get the ETH utilization rate
	wg.Go(func() error {
		_ethUtilizationRate, err := network.GetETHUtilizationRate(collector.rp, nil)
		if err != nil {
			return fmt.Errorf("Error getting ETH utilization rate: %w", err)
		}

		ethUtilizationRate = _ethUtilizationRate
		return nil
	})

	// Get the total ETH staking balance
	wg.Go(func() error {
		totalStakingBalance, err := network.GetStakingETHBalance(collector.rp, nil)
		if err != nil {
			return fmt.Errorf("Error getting total ETH staking balance: %w", err)
		}

		balanceFloat = eth.WeiToEth(totalStakingBalance)
		return nil
	})

	// Get the ETH-rETH exchange rate
	wg.Go(func() error {
		_exchangeRate, err := tokens.GetRETHExchangeRate(collector.rp, nil)
		if err != nil {
			return fmt.Errorf("Error getting ETH-rETH exchange rate: %w", err)
		}

		exchangeRate = _exchangeRate
		return nil
	})

	// Get the total ETH balance (TVL)
	wg.Go(func() error {
		tvl, err := network.GetTotalETHBalance(collector.rp, nil)
		if err != nil {
			return fmt.Errorf("Error getting total ETH balance (TVL): %w", err)
		}

		tvlFloat = eth.WeiToEth(tvl)
		return nil
	})

	// Get the ETH balance of the rETH contract
	wg.Go(func() error {
		rETHContract, err := collector.rp.GetContract("rocketTokenRETH", nil)
		if err != nil {
			return fmt.Errorf("Error getting ETH balance of rETH staking contract: %w", err)
		}

		balance, err := collector.rp.Client.BalanceAt(context.Background(), *rETHContract.Address, nil)
		if err != nil {
			return fmt.Errorf("Error getting ETH balance of rETH staking contract: %w", err)
		}

		rETHBalance = eth.WeiToEth(balance)
		return nil
	})

	// Get the total rETH supply
	wg.Go(func() error {
		totalRethSupply, err := tokens.GetRETHTotalSupply(collector.rp, nil)
		if err != nil {
			return fmt.Errorf("Error getting total rETH supply: %w", err)
		}

		rethFloat = eth.WeiToEth(totalRethSupply)
		return nil
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		log.Printf("%s\n", err.Error())
		return
	}

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
