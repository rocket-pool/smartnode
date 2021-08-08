package collectors

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
)

// Represents the collector for the RPL metrics
type RplCollector struct {
	// The RPL price (in terms of ETH)
	rplPrice 			    *prometheus.Desc

	// The total amount of RPL staked on the network
	totalValueStaked	    *prometheus.Desc

	// The total effective amount of RPL staked on the network
	totalEffectiveStaked    *prometheus.Desc

	// The Rocket Pool contract manager
    rp 						*rocketpool.RocketPool
}


// Create a new DemandCollector instance
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
		rp: rp,
	}
}


// Write metric descriptions to the Prometheus channel
func (collector *RplCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- collector.rplPrice
	channel <- collector.totalValueStaked
	channel <- collector.totalEffectiveStaked
}


// Collect the latest metric values and pass them to Prometheus
func (collector *RplCollector) Collect(channel chan<- prometheus.Metric) {
 
	// Get the RPL price (in terms of ETH)
	rplPrice, err := network.GetRPLPrice(collector.rp, nil)
	rplPriceFloat := float64(-1)
	if err != nil {
		log.Printf("Error getting RPL price: %s", err)
	} else {
        rplPriceFloat = eth.WeiToEth(rplPrice)
    }
    channel <- prometheus.MustNewConstMetric(
		collector.rplPrice, prometheus.GaugeValue, rplPriceFloat)
	
	// Get the total amount of RPL staked on the network
	totalValueStaked, err := node.GetTotalRPLStake(collector.rp, nil)
	totalValueStakedFloat := float64(-1)
	if err != nil {
		log.Printf("Error getting total amount of RPL staked on the network: %s", err)
	} else {
        totalValueStakedFloat = eth.WeiToEth(totalValueStaked)
    }
    channel <- prometheus.MustNewConstMetric(
		collector.totalValueStaked, prometheus.GaugeValue, totalValueStakedFloat)

	// Get the total effective amount of RPL staked on the network
	totalEffectiveStaked, err := node.GetTotalEffectiveRPLStake(collector.rp, nil)
	totalEffectiveStakedFloat := float64(-1)
	if err != nil {
		log.Printf("Error getting total effective amount of RPL staked on the network: %s", err)
	} else {
        totalEffectiveStakedFloat = eth.WeiToEth(totalEffectiveStaked)
    }
    channel <- prometheus.MustNewConstMetric(
		collector.totalEffectiveStaked, prometheus.GaugeValue, totalEffectiveStakedFloat)
}
