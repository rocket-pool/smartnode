package collectors

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prometheus/client_golang/prometheus"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/voting"
	"github.com/rocket-pool/smartnode/v2/shared/keys"
	"github.com/rocket-pool/smartnode/v2/shared/types"
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

	// The logger
	logger *slog.Logger

	// Store values from the latest API call
	cachedNodeVotingPower      float64
	cachedDelegateVotingPower  float64
	cachedVotesClosedProposals float64
	cachedVotesActiveProposals float64
	cachedActiveProposals      float64
	cachedClosedProposals      float64

	// Store the last execution time
	lastApiCallTimestamp time.Time
}

// Create a new SnapshotCollector instance
func NewSnapshotCollector(logger *log.Logger, sp *services.ServiceProvider) *SnapshotCollector {
	subsystem := "snapshot"
	sublogger := logger.With(slog.String(keys.RoutineKey, "Snapshot Collector"))
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
		sp:     sp,
		logger: sublogger,
	}
}

// Write metric descriptions to the Prometheus channel
func (c *SnapshotCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- c.activeProposals
	channel <- c.closedProposals
	channel <- c.votesActiveProposals
	channel <- c.votesClosedProposals
	channel <- c.nodeVotingPower
	channel <- c.delegateVotingPower
}

// Collect the latest metric values and pass them to Prometheus
func (c *SnapshotCollector) Collect(channel chan<- prometheus.Metric) {
	// Update everything if there's an update due
	if time.Since(c.lastApiCallTimestamp).Hours() >= hoursToWait {
		var wg errgroup.Group
		activeProposals := float64(0)
		closedProposals := float64(0)
		votesActiveProposals := float64(0)
		votesClosedProposals := float64(0)

		// Services
		rp := c.sp.GetRocketPool()
		cfg := c.sp.GetConfig()
		snapshotID := cfg.GetVotingSnapshotID()
		nodeAddress, hasNodeAddress := c.sp.GetWallet().GetAddress()
		if !hasNodeAddress {
			return
		}
		snapshot := c.sp.GetSnapshotDelegation()
		if snapshot == nil {
			return
		}

		var delegateAddress common.Address
		err := rp.Query(func(mc *batch.MultiCaller) error {
			snapshot.Delegation(mc, &delegateAddress, nodeAddress, snapshotID)
			return nil
		}, nil)
		if err != nil {
			c.logger.Error("Error getting voting delegate for node", log.Err(err))
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

			c.cachedActiveProposals = activeProposals
			c.cachedClosedProposals = closedProposals
			c.cachedVotesActiveProposals = votesActiveProposals
			c.cachedVotesClosedProposals = votesClosedProposals
			return nil
		})

		// Get the node's voting power
		wg.Go(func() error {
			votingPower, err := voting.GetSnapshotVotingPower(cfg, nodeAddress)
			if err != nil {
				return fmt.Errorf("Error getting Snapshot voted proposals for node address: %w", err)
			}
			c.cachedNodeVotingPower = votingPower
			return nil
		})

		// Get the delegate's voting power
		wg.Go(func() error {
			delegateVotingPower, err := voting.GetSnapshotVotingPower(cfg, delegateAddress)
			if err != nil {
				return fmt.Errorf("Error getting Snapshot voted proposals for delegate address: %w", err)
			}
			c.cachedDelegateVotingPower = delegateVotingPower
			return nil
		})

		// Wait for data
		if err := wg.Wait(); err != nil {
			c.logger.Error(err.Error())
			return
		}

		c.lastApiCallTimestamp = time.Now()
	}

	channel <- prometheus.MustNewConstMetric(
		c.votesActiveProposals, prometheus.GaugeValue, c.cachedVotesActiveProposals)
	channel <- prometheus.MustNewConstMetric(
		c.votesClosedProposals, prometheus.GaugeValue, c.cachedVotesClosedProposals)
	channel <- prometheus.MustNewConstMetric(
		c.activeProposals, prometheus.GaugeValue, c.cachedActiveProposals)
	channel <- prometheus.MustNewConstMetric(
		c.closedProposals, prometheus.GaugeValue, c.cachedClosedProposals)
	channel <- prometheus.MustNewConstMetric(
		c.nodeVotingPower, prometheus.GaugeValue, c.cachedNodeVotingPower)
	channel <- prometheus.MustNewConstMetric(
		c.delegateVotingPower, prometheus.GaugeValue, c.cachedDelegateVotingPower)
}
