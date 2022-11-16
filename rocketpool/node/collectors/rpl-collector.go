package collectors

import (
	"fmt"
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"golang.org/x/sync/errgroup"
)

// Represents the collector for the RPL metrics
type RplCollector struct {
	// The RPL price (in terms of ETH)
	rplPrice *prometheus.Desc

	// The total amount of RPL staked on the network
	totalValueStaked *prometheus.Desc

	// The total effective amount of RPL staked on the network
	totalEffectiveStaked *prometheus.Desc

	// The date and time of the next RPL rewards checkpoint
	checkpointTime *prometheus.Desc

	// The Rocket Pool contract manager
	rp *rocketpool.RocketPool
}

// Create a new RplCollector instance
func NewRplCollector(rp *rocketpool.RocketPool) *RplCollector {
	subsystem := "rpl"
	return &RplCollector{
		rplPrice: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "rpl_price"),
			"The RPL price (in terms of ETH)",
			nil, nil,
		),
		totalValueStaked: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "total_value_staked"),
			"The total amount of RPL staked on the network",
			nil, nil,
		),
		totalEffectiveStaked: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "total_effective_staked"),
			"The total effective amount of RPL staked on the network",
			nil, nil,
		),
		checkpointTime: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "checkpoint_time"),
			"The date and time of the next RPL rewards checkpoint",
			nil, nil,
		),
		rp: rp,
	}
}

// Write metric descriptions to the Prometheus channel
func (collector *RplCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- collector.rplPrice
	channel <- collector.totalValueStaked
	channel <- collector.totalEffectiveStaked
	channel <- collector.checkpointTime
}

// Collect the latest metric values and pass them to Prometheus
func (collector *RplCollector) Collect(channel chan<- prometheus.Metric) {

	// Sync
	var wg errgroup.Group
	rplPriceFloat := float64(-1)
	totalValueStakedFloat := float64(-1)
	totalEffectiveStakedFloat := float64(-1)
	var lastCheckpoint time.Time
	var rewardsInterval time.Duration

	// Get the RPL price (in terms of ETH)
	wg.Go(func() error {
		rplPrice, err := network.GetRPLPrice(collector.rp, nil)
		if err != nil {
			return fmt.Errorf("Error getting RPL price: %w", err)
		}

		rplPriceFloat = eth.WeiToEth(rplPrice)
		return nil
	})

	// Get the total amount of RPL staked on the network
	wg.Go(func() error {
		totalValueStaked, err := node.GetTotalRPLStake(collector.rp, nil)
		if err != nil {
			return fmt.Errorf("Error getting total amount of RPL staked on the network: %w", err)
		}

		totalValueStakedFloat = eth.WeiToEth(totalValueStaked)
		return nil
	})

	// Get the total effective amount of RPL staked on the network
	wg.Go(func() error {
		totalEffectiveStaked, err := node.GetTotalEffectiveRPLStake(collector.rp, nil)
		if err != nil {
			return fmt.Errorf("Error getting total effective amount of RPL staked on the network: %w", err)
		}

		totalEffectiveStakedFloat = eth.WeiToEth(totalEffectiveStaked)
		return nil
	})

	// Get the start of the rewards checkpoint
	wg.Go(func() error {
		_lastCheckpoint, err := rewards.GetClaimIntervalTimeStart(collector.rp, nil)
		if err != nil {
			return fmt.Errorf("Error getting the previous rewards claim time: %w", err)
		}

		lastCheckpoint = _lastCheckpoint
		return err
	})

	// Get the rewards checkpoint interval
	wg.Go(func() error {
		_rewardsInterval, err := rewards.GetClaimIntervalTime(collector.rp, nil)
		if err != nil {
			return fmt.Errorf("Error getting the rewards interval: %w", err)
		}

		rewardsInterval = _rewardsInterval
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		log.Printf("%s\n", err.Error())
		return
	}

	nextRewardsTime := float64(lastCheckpoint.Add(rewardsInterval).Unix()) * 1000

	channel <- prometheus.MustNewConstMetric(
		collector.rplPrice, prometheus.GaugeValue, rplPriceFloat)
	channel <- prometheus.MustNewConstMetric(
		collector.totalValueStaked, prometheus.GaugeValue, totalValueStakedFloat)
	channel <- prometheus.MustNewConstMetric(
		collector.totalEffectiveStaked, prometheus.GaugeValue, totalEffectiveStakedFloat)
	channel <- prometheus.MustNewConstMetric(
		collector.checkpointTime, prometheus.GaugeValue, nextRewardsTime)
}
