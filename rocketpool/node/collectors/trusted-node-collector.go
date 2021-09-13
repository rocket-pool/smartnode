package collectors

import (
	"fmt"
	"log"
	"math/big"
	"strconv"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rocket-pool/rocketpool-go/dao"
	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/tokens"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"golang.org/x/sync/errgroup"
)

// Represents the collector for the user's trusted node
type TrustedNodeCollector struct {
	// The number of proposals per state
	proposalCount *prometheus.Desc

	// The number of votable proposals this node has yet to vote on
	unvotedProposalCount *prometheus.Desc

	// Tabular data for the votes for and against each proposal
	proposalTable *prometheus.Desc

	// The ETH balance of each trusted node
	ethBalance *prometheus.Desc

	// This node's relative balances participation performance
	balancesPerformance *prometheus.Desc

	// This node's relative prices participation performance
	pricesPerformance *prometheus.Desc

	// The Rocket Pool contract manager
	rp *rocketpool.RocketPool

	// The beacon client
	bc beacon.Client

	// The node's address
	nodeAddress common.Address

	// Cached data
	cacheTime             time.Time
	balancesParticipation *node.TrustedNodeParticipation
	pricesParticipation   *node.TrustedNodeParticipation

    // The event log interval for the current eth1 client
    eventLogInterval        *big.Int
}

// Create a new NodeCollector instance
func NewTrustedNodeCollector(rp *rocketpool.RocketPool, bc beacon.Client, nodeAddress common.Address, cfg config.RocketPoolConfig) *TrustedNodeCollector {
	
    // Get the event log interval
    eventLogInterval, err := api.GetEventLogInterval(cfg)
    if err != nil {
        log.Printf("Error getting event log interval: %s\n", err.Error())
        return nil
    }
	
	subsystem := "trusted_node"
	return &TrustedNodeCollector{
		proposalCount: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "proposal_count"),
			"The number of proposals in each state",
			[]string{"state"}, nil,
		),
		unvotedProposalCount: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "unvoted_proposal_count"),
			"How many active proposals has this trusted node has not voted on",
			nil, nil,
		),
		proposalTable: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "proposal_table"),
			"Tabular data of each active proposal's for and against votes",
			[]string{"proposal", "category"}, nil,
		),
		ethBalance: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "eth_balance"),
			"The ETH balance of each trusted node",
			[]string{"member"}, nil,
		),
		balancesPerformance: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "balances_performance"),
			"This node's relative balances participation performance",
			nil, nil,
		),
		pricesPerformance: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "prices_performance"),
			"This node's relative prices participation performance",
			nil, nil,
		),
		rp:          rp,
		bc:          bc,
		nodeAddress: nodeAddress,
        eventLogInterval: eventLogInterval,
	}
}

// Write metric descriptions to the Prometheus channel
func (collector *TrustedNodeCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- collector.proposalCount
	channel <- collector.unvotedProposalCount
	channel <- collector.proposalTable
	channel <- collector.ethBalance
	channel <- collector.balancesPerformance
	channel <- collector.pricesPerformance
}

// Collect the latest metric values and pass them to Prometheus
func (collector *TrustedNodeCollector) Collect(channel chan<- prometheus.Metric) {

	// Sync
	var wg errgroup.Group
	var err error

	var proposals []dao.ProposalDetails
	memberIds := make(map[common.Address]string)
	ethBalances := make(map[string]float64)

	unvotedCount := float64(0)
	pendingCount := float64(0)
	activeCount := float64(0)
	succeededCount := float64(0)
	executedCount := float64(0)
	cancelledCount := float64(0)
	defeatedCount := float64(0)
	expiredCount := float64(0)

	// Get the total staked RPL
	wg.Go(func() error {
		proposals, err = dao.GetProposalsWithMember(collector.rp, collector.nodeAddress, nil)
		if err != nil {
			return fmt.Errorf("Error getting proposals: %w", err)
		}
		return nil
	})

	// Get member IDs
	wg.Go(func() error {
		members, err := trustednode.GetMembers(collector.rp, nil)
		if err != nil {
			return fmt.Errorf("Error getting members: %w", err)
		}
		for _, member := range members {
			memberIds[member.Address] = member.ID
		}
		return nil
	})

	// Only collect fresh participation from chain every 30 minutes
	now := time.Now()
	if now.Unix() > collector.cacheTime.Add(time.Minute*30).Unix() {
		// Get the balances participation data
		wg.Go(func() error {
			var err error
			collector.balancesParticipation, err = node.CalculateTrustedNodeBalancesParticipation(collector.rp, collector.eventLogInterval, nil)
			if err != nil {
				return fmt.Errorf("Error getting trusted node balances participation data: %w", err)
			}
			return nil
		})

		// Get the prices participation data
		wg.Go(func() error {
			var err error
			collector.pricesParticipation, err = node.CalculateTrustedNodePricesParticipation(collector.rp, collector.eventLogInterval, nil)
			if err != nil {
				return fmt.Errorf("Error getting trusted node prices participation data: %w", err)
			}
			return nil
		})

		collector.cacheTime = now
	}

	// Wait for data
	if err := wg.Wait(); err != nil {
		log.Printf("%s\n", err.Error())
		return
	}

	lock := sync.Mutex{}

	// Get member ETH balances
	for address, memberId := range memberIds {
		wg.Go(func(address common.Address, id string) func() error {
			return func() error {
				// Get balances
				balances, err := tokens.GetBalances(collector.rp, address, nil)
				if err != nil {
					return fmt.Errorf("Error getting node balances: %w", err)
				}
				lock.Lock()
				ethBalances[id] = eth.WeiToEth(balances.ETH)
				lock.Unlock()
				return nil
			}
		}(address, memberId))
	}

	// Wait for data
	if err := wg.Wait(); err != nil {
		log.Printf("%s\n", err.Error())
		return
	}

	// Calculate metrics
	for _, proposal := range proposals {
		switch proposal.State {
		case types.Pending:
			pendingCount++
		case types.Active:
			activeCount++
			if !proposal.MemberVoted {
				unvotedCount++
			}
		case types.Succeeded:
			succeededCount++
		case types.Executed:
			executedCount++
		case types.Cancelled:
			cancelledCount++
		case types.Defeated:
			defeatedCount++
		case types.Expired:
			expiredCount++
		}
	}

	// Update all the metrics
	channel <- prometheus.MustNewConstMetric(
		collector.unvotedProposalCount, prometheus.GaugeValue, unvotedCount)
	channel <- prometheus.MustNewConstMetric(
		collector.proposalCount, prometheus.GaugeValue, pendingCount, "pending")
	channel <- prometheus.MustNewConstMetric(
		collector.proposalCount, prometheus.GaugeValue, activeCount, "active")
	channel <- prometheus.MustNewConstMetric(
		collector.proposalCount, prometheus.GaugeValue, succeededCount, "succeeded")
	channel <- prometheus.MustNewConstMetric(
		collector.proposalCount, prometheus.GaugeValue, executedCount, "executed")
	channel <- prometheus.MustNewConstMetric(
		collector.proposalCount, prometheus.GaugeValue, cancelledCount, "cancelled")
	channel <- prometheus.MustNewConstMetric(
		collector.proposalCount, prometheus.GaugeValue, defeatedCount, "defeated")
	channel <- prometheus.MustNewConstMetric(
		collector.proposalCount, prometheus.GaugeValue, expiredCount, "expired")

	// Update balance metrics
	for memberId, balance := range ethBalances {
		channel <- prometheus.MustNewConstMetric(
			collector.ethBalance, prometheus.GaugeValue, balance, memberId)
	}

	// Update proposal metrics
	for _, proposal := range proposals {
		if proposal.State != types.Active {
			continue
		}
		channel <- prometheus.MustNewConstMetric(
			collector.proposalTable, prometheus.GaugeValue, proposal.VotesFor, strconv.FormatUint(proposal.ID, 10), "for")
		channel <- prometheus.MustNewConstMetric(
			collector.proposalTable, prometheus.GaugeValue, proposal.VotesAgainst, strconv.FormatUint(proposal.ID, 10), "against")
		channel <- prometheus.MustNewConstMetric(
			collector.proposalTable, prometheus.GaugeValue, proposal.VotesRequired, strconv.FormatUint(proposal.ID, 10), "required")
	}

	// Update participation metrics
	balancesPerformance := float64(1)
	if collector.balancesParticipation != nil && collector.balancesParticipation.ExpectedSubmissions > 0 {
		if count, ok := collector.balancesParticipation.ActualSubmissions[collector.nodeAddress]; ok {
			balancesPerformance = count / collector.balancesParticipation.ExpectedSubmissions
		}
	}
	channel <- prometheus.MustNewConstMetric(
		collector.balancesPerformance, prometheus.GaugeValue, balancesPerformance)

	pricesPerformance := float64(1)
	if collector.pricesParticipation != nil && collector.pricesParticipation.ExpectedSubmissions > 0 {
		if count, ok := collector.pricesParticipation.ActualSubmissions[collector.nodeAddress]; ok {
			pricesPerformance = count / collector.pricesParticipation.ExpectedSubmissions
		}
	}
	channel <- prometheus.MustNewConstMetric(
		collector.pricesPerformance, prometheus.GaugeValue, pricesPerformance)
}
