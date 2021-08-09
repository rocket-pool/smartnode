package collectors

import (
	"fmt"
	"log"
	"math"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/tokens"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"golang.org/x/sync/errgroup"
)

// Represents the collector for the user's node
type NodeCollector struct {
	// The total amount of RPL staked on the node
	totalStakedRpl 			*prometheus.Desc

	// The effective amount of RPL staked on the node (honoring the 150% collateral cap)
	effectiveStakedRpl		*prometheus.Desc

	// The cumulative RPL rewards earned by the node
	cumulativeRplRewards    *prometheus.Desc

	// The expected RPL rewards for the node at the next rewards checkpoint
	expectedRplRewards   	*prometheus.Desc

    // The estimated APY of RPL for the node
    rplApy                  *prometheus.Desc

    // The token balances of your node wallet
    balances              *prometheus.Desc

    // The public key of a minipool
    pubkeys                 *prometheus.Desc

    // The balance of a validator on the Beacon Chain
    beaconBalance           *prometheus.Desc

    // The node's share of the Beacon Chain rewards for a minipool
    yourShare               *prometheus.Desc

    // The ETH APY for a minipool
    ethApy                  *prometheus.Desc

	// The Rocket Pool contract manager
	rp 					    *rocketpool.RocketPool

    nodeAddress             common.Address
}


// Create a new NodeCollector instance
func NewNodeCollector(rp *rocketpool.RocketPool, nodeAddress common.Address) *NodeCollector {
	subsystem := "node"
	return &NodeCollector{
		totalStakedRpl: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "total_staked_rpl"),
			"The total amount of RPL staked on the node",
			nil, nil,
		),
		effectiveStakedRpl: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "effective_staked_rpl"),
			"The effective amount of RPL staked on the node (honoring the 150% collateral cap)",
			nil, nil,
		),
		cumulativeRplRewards: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "cumulative_rpl_rewards"),
			"The cumulative RPL rewards earned by the node",
			nil, nil,
		),
		expectedRplRewards: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "expected_rpl_rewards"),
			"The expected RPL rewards for the node at the next rewards checkpoint",
			nil, nil,
		),
		rplApy: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "rpl_apy"),
			"The estimated APY of RPL for the node",
			nil, nil,
		),
		balances: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "balance"),
			"How much ETH is in this node wallet",
			[]string{"token"}, nil,
		),
		rp: rp,
        nodeAddress: nodeAddress,
	}
}


// Write metric descriptions to the Prometheus channel
func (collector *NodeCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- collector.totalStakedRpl
	channel <- collector.effectiveStakedRpl
	channel <- collector.cumulativeRplRewards
	channel <- collector.expectedRplRewards
	channel <- collector.rplApy
	channel <- collector.balances
}


// Collect the latest metric values and pass them to Prometheus
func (collector *NodeCollector) Collect(channel chan<- prometheus.Metric) {
 
    // Sync
    var wg errgroup.Group
    stakedRpl := float64(-1)
    effectiveStakedRpl := float64(-1)
    cumulativeRewards := float64(-1)
    var rewardsInterval time.Duration
    var inflationInterval *big.Int
    var totalRplSupply *big.Int
    var totalEffectiveStake *big.Int
    var nodeOperatorRewardsPercent float64
    var nodeRegistrationTime time.Time
    ethBalance := float64(-1)
    oldRplBalance := float64(-1)
    newRplBalance := float64(-1)
    rethBalance := float64(-1)

    // Get the total staked RPL
    wg.Go(func() error {
        stakedRplWei, err := node.GetNodeRPLStake(collector.rp, collector.nodeAddress, nil)
        if err != nil {
            return fmt.Errorf("Error getting total staked RPL: %w", err)
        } else {
            stakedRpl = eth.WeiToEth(stakedRplWei)
        }
        return nil
    })
    
    // Get the effective staked RPL
    wg.Go(func() error {
        effectiveStakedRplWei, err := node.GetNodeEffectiveRPLStake(collector.rp, collector.nodeAddress, nil)
        if err != nil {
            return fmt.Errorf("Error getting effective staked RPL: %w", err)
        } else {
            effectiveStakedRpl = eth.WeiToEth(effectiveStakedRplWei)
        }
        return nil
    })

    // Get the cumulative RPL rewards
    wg.Go(func() error {
        cumulativeRewardsWei, err := rewards.CalculateLifetimeNodeRewards(collector.rp, collector.nodeAddress)
        if err != nil {
            return fmt.Errorf("Error getting cumulative RPL rewards: %w", err)
        } else {
            cumulativeRewards = eth.WeiToEth(cumulativeRewardsWei)
        }
        return nil
    })

    // Get the rewards checkpoint interval
    wg.Go(func() error {
        _rewardsInterval, err := rewards.GetClaimIntervalTime(collector.rp, nil)
        if err != nil {
            return fmt.Errorf("Error getting rewards checkpoint interval: %w", err) 
        }
        rewardsInterval = _rewardsInterval
        return nil
    })
    
    // Get the RPL inflation interval
    wg.Go(func() error {
        _inflationInterval, err := tokens.GetRPLInflationIntervalRate(collector.rp, nil)
        if err != nil {
            return fmt.Errorf("Error getting RPL inflation interval: %w", err)
        }
        inflationInterval = _inflationInterval
        return nil
    })

    // Get the total RPL supply
    wg.Go(func() error {
        _totalRplSupply, err := tokens.GetRPLTotalSupply(collector.rp, nil)
        if err != nil {
            return fmt.Errorf("Error getting total RPL supply: %w", err)
        }
        totalRplSupply = _totalRplSupply
        return nil
    })

    // Get the total network effective stake
    wg.Go(func() error {
        _totalEffectiveStake, err := node.GetTotalEffectiveRPLStake(collector.rp, nil)
        if err != nil {
            return fmt.Errorf("Error getting total network effective stake: %w", err)
        }
        totalEffectiveStake = _totalEffectiveStake
        return nil
    })

    // Get the node operator rewards percent
    wg.Go(func() error {
        _nodeOperatorRewardsPercent, err := rewards.GetNodeOperatorRewardsPercent(collector.rp, nil)
        if err != nil {
            return fmt.Errorf("Error getting node operator rewards percent: %w", err)
        }
        nodeOperatorRewardsPercent = _nodeOperatorRewardsPercent
        return nil
    })

    // Get the node registration time
    wg.Go(func() error {
        _nodeRegistrationTime, err := rewards.GetNodeRegistrationTime(collector.rp, collector.nodeAddress, nil)
        if err != nil {
            return fmt.Errorf("Error getting node registration time: %w", err)
        }
        nodeRegistrationTime = _nodeRegistrationTime
        return nil
    })

    // Get the node balances
    wg.Go(func() error {
        balances, err := tokens.GetBalances(collector.rp, collector.nodeAddress, nil)
        if err != nil {
            return fmt.Errorf("Error getting node balances: %w", err)
        }
        ethBalance = eth.WeiToEth(balances.ETH)
        oldRplBalance = eth.WeiToEth(balances.FixedSupplyRPL)
        newRplBalance = eth.WeiToEth(balances.RPL)
        rethBalance = eth.WeiToEth(balances.RETH)
        return nil
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        log.Printf("%s\n", err.Error())
    }

    // Calculate the estimated rewards
    rewardsIntervalDays := rewardsInterval.Seconds() / (60*60*24)
    inflationPerDay := eth.WeiToEth(inflationInterval)
    totalRplAtNextCheckpoint := (math.Pow(inflationPerDay, float64(rewardsIntervalDays)) - 1) * eth.WeiToEth(totalRplSupply)
    estimatedRewards := float64(0)
    if totalEffectiveStake.Cmp(big.NewInt(0)) == 1 {
        estimatedRewards = effectiveStakedRpl / eth.WeiToEth(totalEffectiveStake) * totalRplAtNextCheckpoint * nodeOperatorRewardsPercent
    }

    // Calculate the RPL APY
    timeSinceRegistration := time.Since(nodeRegistrationTime)
    rplPerYear := (cumulativeRewards + estimatedRewards) / timeSinceRegistration.Hours() * (24*365) 
    rplApy := rplPerYear / effectiveStakedRpl * 100;

    channel <- prometheus.MustNewConstMetric(
        collector.totalStakedRpl, prometheus.GaugeValue, stakedRpl)
    channel <- prometheus.MustNewConstMetric(
        collector.effectiveStakedRpl, prometheus.GaugeValue, effectiveStakedRpl)
    channel <- prometheus.MustNewConstMetric(
        collector.cumulativeRplRewards, prometheus.GaugeValue, cumulativeRewards)
    channel <- prometheus.MustNewConstMetric(
        collector.expectedRplRewards, prometheus.GaugeValue, estimatedRewards)
    channel <- prometheus.MustNewConstMetric(
        collector.rplApy, prometheus.GaugeValue, rplApy)
    channel <- prometheus.MustNewConstMetric(
        collector.balances, prometheus.GaugeValue, ethBalance, "ETH")
    channel <- prometheus.MustNewConstMetric(
        collector.balances, prometheus.GaugeValue, oldRplBalance, "Legacy RPL")
    channel <- prometheus.MustNewConstMetric(
        collector.balances, prometheus.GaugeValue, newRplBalance, "New RPL")
    channel <- prometheus.MustNewConstMetric(
        collector.balances, prometheus.GaugeValue, rethBalance, "rETH")

}
