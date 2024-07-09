package collectors

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/rocketpool/api/pdao"
	"github.com/rocket-pool/smartnode/shared/services/config"
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

	// The Rocket Pool config
	cfg *config.RocketPoolConfig

	// the node wallet address
	nodeAddress common.Address

	// the delegate address
	delegateAddress common.Address

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
func NewSnapshotCollector(rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, nodeAddress common.Address, delegateAddress common.Address) *SnapshotCollector {
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
		cfg:             cfg,
		nodeAddress:     nodeAddress,
		delegateAddress: delegateAddress,
		logPrefix:       "Snapshot Collector",
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

	// Sync
	var wg errgroup.Group
	activeProposals := float64(0)
	closedProposals := float64(0)
	votesActiveProposals := float64(0)
	votesClosedProposals := float64(0)
	handledProposals := map[string]bool{}

	// Get the number of votes on Snapshot proposals
	wg.Go(func() error {
		if time.Since(collector.lastApiCallTimestamp).Hours() >= hoursToWait {
			votedProposals, err := pdao.GetSnapshotVotedProposals(collector.cfg.Smartnode.GetSnapshotApiDomain(), collector.cfg.Smartnode.GetSnapshotID(), collector.nodeAddress, collector.delegateAddress)
			if err != nil {
				return fmt.Errorf("Error getting Snapshot voted proposals: %w", err)
			}

			for _, votedProposal := range votedProposals.Data.Votes {
				_, exists := handledProposals[votedProposal.Proposal.Id]
				if !exists {
					if votedProposal.Proposal.State == "active" {
						votesActiveProposals += 1
					} else {
						votesClosedProposals += 1
					}
					handledProposals[votedProposal.Proposal.Id] = true
				}
			}
			collector.cachedVotesActiveProposals = votesActiveProposals
			collector.cachedVotesClosedProposals = votesClosedProposals

		}

		return nil
	})

	// Get the number of live Snapshot proposals
	wg.Go(func() error {
		if time.Since(collector.lastApiCallTimestamp).Hours() >= hoursToWait {
			proposals, err := pdao.GetSnapshotProposals(collector.cfg.Smartnode.GetSnapshotApiDomain(), collector.cfg.Smartnode.GetSnapshotID(), "")
			if err != nil {
				return fmt.Errorf("Error getting Snapshot voted proposals: %w", err)
			}

			for _, proposal := range proposals.Data.Proposals {
				if proposal.State == "active" {
					activeProposals += 1
				} else {
					closedProposals += 1
				}
			}
			collector.cachedActiveProposals = activeProposals
			collector.cachedClosedProposals = closedProposals
		}

		return nil
	})

	// Get the node's voting power
	wg.Go(func() error {
		if time.Since(collector.lastApiCallTimestamp).Hours() >= hoursToWait {

			votingPowerResponse, err := pdao.GetSnapshotVotingPower(collector.cfg.Smartnode.GetSnapshotApiDomain(), collector.cfg.Smartnode.GetSnapshotID(), collector.nodeAddress)
			if err != nil {
				return fmt.Errorf("Error getting Snapshot voted proposals for node address: %w", err)
			}

			collector.cachedNodeVotingPower = votingPowerResponse.Data.Vp.Vp
		}
		return nil
	})

	// Get the delegate's voting power
	wg.Go(func() error {
		if time.Since(collector.lastApiCallTimestamp).Hours() >= hoursToWait {
			votingPowerResponse, err := pdao.GetSnapshotVotingPower(collector.cfg.Smartnode.GetSnapshotApiDomain(), collector.cfg.Smartnode.GetSnapshotID(), collector.delegateAddress)
			if err != nil {
				return fmt.Errorf("Error getting Snapshot voted proposals for delegate address: %w", err)
			}

			collector.cachedDelegateVotingPower = votingPowerResponse.Data.Vp.Vp
		}
		return nil
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		collector.logError(err)
		return
	}
	if time.Since(collector.lastApiCallTimestamp).Hours() >= hoursToWait {
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
