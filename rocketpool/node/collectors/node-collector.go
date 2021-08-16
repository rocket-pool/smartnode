package collectors

import (
	"fmt"
	"log"
	"math"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/tokens"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/utils/rp"
	"golang.org/x/sync/errgroup"
)

// Settings
const MinipoolBalanceDetailsBatchSize = 20


// Beacon chain balance info for a minipool
type minipoolBalanceDetails struct {
    IsStaking bool
    NodeDeposit *big.Int
    NodeBalance *big.Int
    TotalBalance *big.Int
}


// Represents the collector for the user's node
type NodeCollector struct {
	// The total amount of RPL staked on the node
	totalStakedRpl 			*prometheus.Desc

	// The effective amount of RPL staked on the node (honoring the 150% collateral cap)
	effectiveStakedRpl		*prometheus.Desc

    // The RPL collateral level for the node
    rplCollateral           *prometheus.Desc

	// The cumulative RPL rewards earned by the node
	cumulativeRplRewards    *prometheus.Desc

	// The expected RPL rewards for the node at the next rewards checkpoint
	expectedRplRewards   	*prometheus.Desc

    // The estimated APR of RPL for the node from the next rewards checkpoint
    rplApr                  *prometheus.Desc

    // The token balances of your node wallet
    balances                *prometheus.Desc

    // The amount of ETH this node deposited into minipools
    depositedEth            *prometheus.Desc

    // The node's total share of its minipool's beacon chain balances 
    beaconShare             *prometheus.Desc

    // The total balances of all this node's validators on the beacon chain
    beaconBalance           *prometheus.Desc

    // The RPL rewards from the last period that have not been claimed yet
    unclaimedRewards        *prometheus.Desc

	// The Rocket Pool contract manager
	rp 					    *rocketpool.RocketPool

	// The beacon client
	bc 					    beacon.Client

    // The node's address
    nodeAddress             common.Address
}


// Create a new NodeCollector instance
func NewNodeCollector(rp *rocketpool.RocketPool, bc beacon.Client, nodeAddress common.Address) *NodeCollector {
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
		rplCollateral: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "rpl_collateral"),
			"The RPL collateral level for the node",
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
		rplApr: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "rpl_apr"),
			"The estimated APR of RPL for the node from the next rewards checkpoint",
			nil, nil,
		),
		balances: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "balance"),
			"How much ETH is in this node wallet",
			[]string{"Token"}, nil,
		),
		depositedEth: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "deposited_eth"),
			"The amount of ETH this node deposited into minipools",
			nil, nil,
		),
		beaconShare: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "beacon_share"),
			"The node's total share of its minipool's beacon chain balances",
			nil, nil,
		),
		beaconBalance: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "beacon_balance"),
			"The total balances of all this node's validators on the beacon chain",
			nil, nil,
        ),
        unclaimedRewards: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "unclaimed_rewards"),
            "The RPL rewards from the last period that have not been claimed yet",
            nil, nil,
		),
		rp: rp,
        bc: bc,
        nodeAddress: nodeAddress,
	}
}


// Write metric descriptions to the Prometheus channel
func (collector *NodeCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- collector.totalStakedRpl
	channel <- collector.effectiveStakedRpl
	channel <- collector.cumulativeRplRewards
	channel <- collector.expectedRplRewards
	channel <- collector.rplApr
	channel <- collector.balances
	channel <- collector.depositedEth
	channel <- collector.beaconShare
	channel <- collector.unclaimedRewards
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
    ethBalance := float64(-1)
    oldRplBalance := float64(-1)
    newRplBalance := float64(-1)
    rethBalance := float64(-1)
    var activeMinipoolCount uint64
    var rplPrice float64
    collateralRatio := float64(-1)
    var addresses []common.Address
    var beaconHead beacon.BeaconHead
    unclaimedRewards := float64(-1)

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

    // Get the number of active minipools on the node
    wg.Go(func() error {
        _activeMinipoolCount, err := minipool.GetNodeActiveMinipoolCount(collector.rp, collector.nodeAddress, nil)
        if err != nil {
            return fmt.Errorf("Error getting node active minipool count: %w", err)
        }
        activeMinipoolCount = _activeMinipoolCount
        return nil
    })

    // Get the RPL price
    wg.Go(func() error {
        rplPriceWei, err := network.GetRPLPrice(collector.rp, nil)
        if err != nil {
            return fmt.Errorf("Error getting RPL price: %w", err)
        }
        rplPrice = eth.WeiToEth(rplPriceWei)
        return nil
    })

    // Get the list of minipool addresses for this node
    wg.Go(func() error {
        _addresses, err := minipool.GetNodeMinipoolAddresses(collector.rp, collector.nodeAddress, nil)
        if err != nil {
            return fmt.Errorf("Error getting node minipool addresses: %w", err)
        }
        addresses = _addresses
        return nil
    })

    // Get the beacon head
    wg.Go(func() error {
        _beaconHead, err := collector.bc.GetBeaconHead()
        if err != nil {
            return fmt.Errorf("Error getting beacon chain head: %w", err)
        }
        beaconHead = _beaconHead
        return nil
    })

    // Get the RPL price
    wg.Go(func() error {
        unclaimedRewardsWei, err := rewards.GetNodeClaimRewardsAmount(collector.rp, collector.nodeAddress, nil)
        if err != nil {
            return fmt.Errorf("Error getting RPL price: %w", err)
        }
        unclaimedRewards = eth.WeiToEth(unclaimedRewardsWei)
        return nil
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        log.Printf("%s\n", err.Error())
        return
    }

    // Calculate the estimated rewards
    rewardsIntervalDays := rewardsInterval.Seconds() / (60*60*24)
    inflationPerDay := eth.WeiToEth(inflationInterval)
    totalRplAtNextCheckpoint := (math.Pow(inflationPerDay, float64(rewardsIntervalDays)) - 1) * eth.WeiToEth(totalRplSupply)
    estimatedRewards := float64(0)
    if totalEffectiveStake.Cmp(big.NewInt(0)) == 1 {
        estimatedRewards = effectiveStakedRpl / eth.WeiToEth(totalEffectiveStake) * totalRplAtNextCheckpoint * nodeOperatorRewardsPercent
    }

    // Calculate the RPL APR
    rplApr := estimatedRewards / stakedRpl / rewardsInterval.Hours() * (24*365) * 100

    // Calculate the collateral ratio
    if activeMinipoolCount > 0 {
        collateralRatio = rplPrice * stakedRpl / (float64(activeMinipoolCount) * 16.0)
    }

    // Calculate the total deposits and corresponding beacon chain balance share
    minipoolDetails, err := collector.getBeaconBalances(addresses, beaconHead, nil)
    if err != nil {
        log.Printf("%s\n", err.Error())
        return
    }
    totalDepositBalance := float64(0)
    totalNodeShare := float64(0)
    totalBeaconBalance := float64(0)
    for _, minipool := range minipoolDetails {
        totalDepositBalance += eth.WeiToEth(minipool.NodeDeposit)
        totalNodeShare += eth.WeiToEth(minipool.NodeBalance)
        totalBeaconBalance += eth.WeiToEth(minipool.TotalBalance)
    }

    // Update all the metrics
    channel <- prometheus.MustNewConstMetric(
        collector.totalStakedRpl, prometheus.GaugeValue, stakedRpl)
    channel <- prometheus.MustNewConstMetric(
        collector.effectiveStakedRpl, prometheus.GaugeValue, effectiveStakedRpl)
    channel <- prometheus.MustNewConstMetric(
        collector.rplCollateral, prometheus.GaugeValue, collateralRatio)
    channel <- prometheus.MustNewConstMetric(
        collector.cumulativeRplRewards, prometheus.GaugeValue, cumulativeRewards)
    channel <- prometheus.MustNewConstMetric(
        collector.expectedRplRewards, prometheus.GaugeValue, estimatedRewards)
    channel <- prometheus.MustNewConstMetric(
        collector.rplApr, prometheus.GaugeValue, rplApr)
    channel <- prometheus.MustNewConstMetric(
        collector.balances, prometheus.GaugeValue, ethBalance, "ETH")
    channel <- prometheus.MustNewConstMetric(
        collector.balances, prometheus.GaugeValue, oldRplBalance, "Legacy RPL")
    channel <- prometheus.MustNewConstMetric(
        collector.balances, prometheus.GaugeValue, newRplBalance, "New RPL")
    channel <- prometheus.MustNewConstMetric(
        collector.balances, prometheus.GaugeValue, rethBalance, "rETH")
    channel <- prometheus.MustNewConstMetric(
        collector.depositedEth, prometheus.GaugeValue, totalDepositBalance)
    channel <- prometheus.MustNewConstMetric(
        collector.beaconShare, prometheus.GaugeValue, totalNodeShare)
    channel <- prometheus.MustNewConstMetric(
        collector.beaconBalance, prometheus.GaugeValue, totalBeaconBalance)
    channel <- prometheus.MustNewConstMetric(
        collector.unclaimedRewards, prometheus.GaugeValue, unclaimedRewards)
}


// Get the balances of the minipools on the beacon chain
func (collector *NodeCollector) getBeaconBalances(addresses []common.Address, beaconHead beacon.BeaconHead, opts *bind.CallOpts) ([]minipoolBalanceDetails, error) {

    // Get minipool validator statuses
    validators, err := rp.GetMinipoolValidators(collector.rp, collector.bc, addresses, opts, &beacon.ValidatorStatusOptions{Epoch: beaconHead.Epoch})
    if err != nil {
        return []minipoolBalanceDetails{}, err
    }

    // Load details in batches
    details := make([]minipoolBalanceDetails, len(addresses))
    for bsi := 0; bsi < len(addresses); bsi += MinipoolBalanceDetailsBatchSize {

        // Get batch start & end index
        msi := bsi
        mei := bsi + MinipoolBalanceDetailsBatchSize
        if mei > len(addresses) { mei = len(addresses) }

        // Load details
        var wg errgroup.Group
        for mi := msi; mi < mei; mi++ {
            mi := mi
            wg.Go(func() error {
                address := addresses[mi]
                validator := validators[address]
                mpDetails, err := collector.getMinipoolBalanceDetails(address, opts, validator, beaconHead.Epoch)
                if err == nil { details[mi] = mpDetails }
                return err
            })
        }
        if err := wg.Wait(); err != nil {
            return []minipoolBalanceDetails{}, err
        }

    }
    
    // Return
    return details, nil
}


// Get minipool balance details
func (collector *NodeCollector) getMinipoolBalanceDetails(minipoolAddress common.Address, opts *bind.CallOpts, validator beacon.ValidatorStatus, blockEpoch uint64) (minipoolBalanceDetails, error) {

    // Create minipool
    mp, err := minipool.NewMinipool(collector.rp, minipoolAddress)
    if err != nil {
        return minipoolBalanceDetails{}, err
    }
    blockBalance := eth.GweiToWei(float64(validator.Balance))

    // Data
    var wg errgroup.Group
    var status types.MinipoolStatus
    var nodeDepositBalance *big.Int

    // Load data
    wg.Go(func() error {
        var err error
        status, err = mp.GetStatus(opts)
        return err
    })
    wg.Go(func() error {
        var err error
        nodeDepositBalance, err = mp.GetNodeDepositBalance(opts)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return minipoolBalanceDetails{}, err
    }

    // Deal with pools that haven't received deposits yet so their balance is still 0
    if nodeDepositBalance == nil {
        nodeDepositBalance = big.NewInt(0)
    }

    // Use node deposit balance if initialized or prelaunch
    if status == types.Initialized || status == types.Prelaunch {
        return minipoolBalanceDetails{
            NodeDeposit: nodeDepositBalance,
            NodeBalance: nodeDepositBalance,
            TotalBalance: blockBalance,
        }, nil
    }

    // Use node deposit balance if validator not yet active on beacon chain at block
    if !validator.Exists || validator.ActivationEpoch >= blockEpoch {
        return minipoolBalanceDetails{
            NodeDeposit: nodeDepositBalance,
            NodeBalance: nodeDepositBalance,
            TotalBalance: blockBalance,
        }, nil
    }

    // Get node balance at block
    nodeBalance, err := mp.CalculateNodeShare(blockBalance, opts)
    if err != nil {
        return minipoolBalanceDetails{}, err
    }

    // Return
    return minipoolBalanceDetails{
        IsStaking: (validator.ExitEpoch > blockEpoch),
        NodeDeposit: nodeDepositBalance,
        NodeBalance: nodeBalance,
        TotalBalance: blockBalance,
    }, nil

}
