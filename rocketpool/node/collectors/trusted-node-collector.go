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
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/tokens"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
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

	// The balances submission participation of the ODAO members
	balancesParticipation *prometheus.Desc

	// The prices submission participation of the ODAO members
	pricesParticipation *prometheus.Desc

	// Whether or not ODAO collection is enabled
	enabled bool

	// The Rocket Pool contract manager
	rp *rocketpool.RocketPool

	// The beacon client
	bc beacon.Client

	// The node's address
	nodeAddress common.Address

	// Cached data
	cacheTime     time.Time
	cachedMetrics []prometheus.Metric

	// The event log interval for the current eth1 client
	eventLogInterval *big.Int

	// The thread-safe locker for the network state
	stateLocker *StateLocker

	// Prefix for logging
	logPrefix string
}

// Create a new NodeCollector instance
func NewTrustedNodeCollector(rp *rocketpool.RocketPool, bc beacon.Client, nodeAddress common.Address, cfg *config.RocketPoolConfig, stateLocker *StateLocker) *TrustedNodeCollector {

	// Get the event log interval
	eventLogInterval, err := cfg.GetEventLogInterval()
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
		balancesParticipation: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "balances_participation"),
			"Whether each member has participated in the current balances update interval",
			[]string{"member"}, nil,
		),
		pricesParticipation: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "prices_participation"),
			"Whether each member has participated in the current prices update interval",
			[]string{"member"}, nil,
		),
		enabled:          cfg.EnableODaoMetrics.Value.(bool),
		rp:               rp,
		bc:               bc,
		nodeAddress:      nodeAddress,
		eventLogInterval: big.NewInt(int64(eventLogInterval)),
		stateLocker:      stateLocker,
		logPrefix:        "ODAO Stats Collector",
	}
}

// Write metric descriptions to the Prometheus channel
func (collector *TrustedNodeCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- collector.proposalCount
	channel <- collector.unvotedProposalCount
	channel <- collector.proposalTable
	channel <- collector.ethBalance
	channel <- collector.balancesParticipation
	channel <- collector.pricesParticipation
}

// Caches slow to process metrics so it doesn't have to be processed every second
func (collector *TrustedNodeCollector) collectSlowMetrics(memberIds map[common.Address]string) {

	if !collector.enabled {
		return
	}

	// Create a new cached metrics array to populate
	collector.cachedMetrics = make([]prometheus.Metric, 0)

	// Sync
	var wg errgroup.Group

	var balancesParticipation map[common.Address]bool
	var pricesParticipation map[common.Address]bool

	// Wait for data
	if err := wg.Wait(); err != nil {
		collector.logError(err)
		return
	}

	// Balances participation
	for member, status := range balancesParticipation {
		value := float64(0)
		if status {
			value = 1
		}
		collector.cachedMetrics = append(collector.cachedMetrics, prometheus.MustNewConstMetric(collector.balancesParticipation, prometheus.GaugeValue, value, memberIds[member]))
	}

	// Prices participation
	for member, status := range pricesParticipation {
		value := float64(0)
		if status {
			value = 1
		}
		collector.cachedMetrics = append(collector.cachedMetrics, prometheus.MustNewConstMetric(collector.pricesParticipation, prometheus.GaugeValue, value, memberIds[member]))
	}
}

// Collect the latest metric values and pass them to Prometheus
func (collector *TrustedNodeCollector) Collect(channel chan<- prometheus.Metric) {

	if !collector.enabled {
		return
	}

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

	// Only collect fresh participation metrics from chain every 60 seconds as it updates infrequently and takes longer to collect
	now := time.Now()
	if now.Unix() > collector.cacheTime.Add(time.Second*60).Unix() {
		collector.collectSlowMetrics(memberIds)
		collector.cacheTime = now
	}

	// Wait for data
	if err := wg.Wait(); err != nil {
		collector.logError(err)
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
		collector.logError(err)
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

	// Include cached metrics
	for _, metric := range collector.cachedMetrics {
		channel <- metric
	}
}

// Log error messages
func (collector *TrustedNodeCollector) logError(err error) {
	fmt.Printf("[%s] %s\n", collector.logPrefix, err.Error())
}
