package collectors

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prometheus/client_golang/prometheus"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/voting"
	"github.com/rocket-pool/smartnode/shared/types"
	"golang.org/x/sync/errgroup"
)

// Time to wait to make new Snapshot API calls
const hoursToWait float64 = 6

// Represents the collector for Snapshot metrics
type SnapshotCollector struct {
	// the number of active Snashot proposals
	activeProposals *prometheus.Desc

	// the number of past Snapshot proposals
	closedProposals *prometheus.Desc

	// the number of votes on active Snapshot proposals
	votesActiveProposals *prometheus.Desc

	// the number of votes on closed Snapshot proposals
	votesClosedProposals *prometheus.Desc

	// The current node voting power on Snapshot
	nodeVotingPower *prometheus.Desc

	// The current delegate voting power on Snapshot
	delegateVotingPower *prometheus.Desc

	// The Smartnode service provider
	sp *services.ServiceProvider

	// Store values from the latest API call
	cachedNodeVotingPower      float64
	cachedDelegateVotingPower  float64
	cachedVotesClosedProposals float64
	cachedVotesActiveProposals float64
	cachedActiveProposals      float64
	cachedClosedProposals      float64

	// Store the last execution time
	lastApiCallTimestamp time.Time

	// Prefix for logging
	logPrefix string
}

// Create a new SnapshotCollector instance
func NewSnapshotCollector(sp *services.ServiceProvider) *SnapshotCollector {
	subsystem := "snapshot"
	return &SnapshotCollector{
		activeProposals: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "proposals_active"),
			"The number of active Snapshot proposals",
			nil, nil,
		),
		closedProposals: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "proposals_closed"),
			"The number of closed Snapshot proposals",
			nil, nil,
		),
		votesActiveProposals: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "votes_active"),
			"The number of votes from user/delegate on active Snapshot proposals",
			nil, nil,
		),
		votesClosedProposals: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "votes_closed"),
			"The number of votes from user/delegate on closed Snapshot proposals",
			nil, nil,
		),
		nodeVotingPower: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "node_vp"),
			"The node current voting power on Snapshot",
			nil, nil,
		),
		delegateVotingPower: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "delegate_vp"),
			"The delegate current voting power on Snapshot",
			nil, nil,
		),
		sp:        sp,
		logPrefix: "Snapshot Collector",
	}
}

// Write metric descriptions to the Prometheus channel
func (collector *SnapshotCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- collector.activeProposals
	channel <- collector.closedProposals
	channel <- collector.votesActiveProposals
	channel <- collector.votesClosedProposals
	channel <- collector.nodeVotingPower
	channel <- collector.delegateVotingPower
}

// Collect the latest metric values and pass them to Prometheus
func (collector *SnapshotCollector) Collect(channel chan<- prometheus.Metric) {
	// Update everything if there's an update due
	if time.Since(collector.lastApiCallTimestamp).Hours() >= hoursToWait {
		var wg errgroup.Group
		activeProposals := float64(0)
		closedProposals := float64(0)
		votesActiveProposals := float64(0)
		votesClosedProposals := float64(0)

		// Services
		rp := collector.sp.GetRocketPool()
		cfg := collector.sp.GetConfig()
		snapshotID := cfg.Smartnode.GetVotingSnapshotID()
		nodeAddress, hasNodeAddress := collector.sp.GetWallet().GetAddress()
		if !hasNodeAddress {
			return
		}
		snapshot := collector.sp.GetSnapshotDelegation()
		if snapshot == nil {
			return
		}

		var delegateAddress common.Address
		err := rp.Query(func(mc *batch.MultiCaller) error {
			snapshot.Delegation(mc, &delegateAddress, nodeAddress, snapshotID)
			return nil
		}, nil)
		if err != nil {
			collector.logError(fmt.Errorf("error getting voting delegate for node: %w", err))
			return
		}

		// Get the number of Snapshot proposals and votes
		wg.Go(func() error {
			proposals, err := voting.GetSnapshotProposals(cfg, nodeAddress, delegateAddress, false)
			if err != nil {
				return fmt.Errorf("error getting Snapshot proposals: %w", err)
			}

			for _, proposal := range proposals {
				switch proposal.State {
				case types.ProposalState_Active:
					activeProposals++
					if len(proposal.UserVotes) > 0 || len(proposal.DelegateVotes) > 0 {
						votesActiveProposals++
					}
				case types.ProposalState_Closed:
					closedProposals++
					if len(proposal.UserVotes) > 0 || len(proposal.DelegateVotes) > 0 {
						votesClosedProposals++
					}
				}
			}

			collector.cachedActiveProposals = activeProposals
			collector.cachedClosedProposals = closedProposals
			collector.cachedVotesActiveProposals = votesActiveProposals
			collector.cachedVotesClosedProposals = votesClosedProposals
			return nil
		})

		// Get the node's voting power
		wg.Go(func() error {
			votingPower, err := voting.GetSnapshotVotingPower(cfg, nodeAddress)
			if err != nil {
				return fmt.Errorf("Error getting Snapshot voted proposals for node address: %w", err)
			}
			collector.cachedNodeVotingPower = votingPower
			return nil
		})

		// Get the delegate's voting power
		wg.Go(func() error {
			delegateVotingPower, err := voting.GetSnapshotVotingPower(cfg, delegateAddress)
			if err != nil {
				return fmt.Errorf("Error getting Snapshot voted proposals for delegate address: %w", err)
			}
			collector.cachedDelegateVotingPower = delegateVotingPower
			return nil
		})

		// Wait for data
		if err := wg.Wait(); err != nil {
			collector.logError(err)
			return
		}

		collector.lastApiCallTimestamp = time.Now()
	}

	channel <- prometheus.MustNewConstMetric(
		collector.votesActiveProposals, prometheus.GaugeValue, collector.cachedVotesActiveProposals)
	channel <- prometheus.MustNewConstMetric(
		collector.votesClosedProposals, prometheus.GaugeValue, collector.cachedVotesClosedProposals)
	channel <- prometheus.MustNewConstMetric(
		collector.activeProposals, prometheus.GaugeValue, collector.cachedActiveProposals)
	channel <- prometheus.MustNewConstMetric(
		collector.closedProposals, prometheus.GaugeValue, collector.cachedClosedProposals)
	channel <- prometheus.MustNewConstMetric(
		collector.nodeVotingPower, prometheus.GaugeValue, collector.cachedNodeVotingPower)
	channel <- prometheus.MustNewConstMetric(
		collector.delegateVotingPower, prometheus.GaugeValue, collector.cachedDelegateVotingPower)
}

// Log error messages
func (collector *SnapshotCollector) logError(err error) {
	fmt.Printf("[%s] %s\n", collector.logPrefix, err.Error())
}
