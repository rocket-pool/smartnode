package collectors

import (
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/rocketpool/api/node"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"golang.org/x/sync/errgroup"
)

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

	// The Rocket Pool config
	cfg *config.RocketPoolConfig

	// the node wallet address
	nodeAddress common.Address

	// the delegate address
	delegateAddress common.Address
}

// Create a new SnapshotCollector instance
func NewSnapshotCollector(rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, nodeAddress common.Address, delegateAddres common.Address) *SnapshotCollector {
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
		cfg:             cfg,
		nodeAddress:     nodeAddress,
		delegateAddress: delegateAddres,
	}
}

// Write metric descriptions to the Prometheus channel
func (collector *SnapshotCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- collector.activeProposals
	channel <- collector.closedProposals
	channel <- collector.votesActiveProposals
	channel <- collector.votesClosedProposals
}

// Collect the latest metric values and pass them to Prometheus
func (collector *SnapshotCollector) Collect(channel chan<- prometheus.Metric) {

	// Sync
	var wg errgroup.Group
	activeProposals := float64(0)
	closedProposals := float64(0)
	votesActiveProposals := float64(0)
	votesClosedProposals := float64(0)

	// Get the number of votes on Snapshot proposals
	wg.Go(func() error {
		votedProposals, err := node.GetSnapshotVotedProposals(collector.cfg.Smartnode.GetSnapshotApiDomain(), collector.cfg.Smartnode.GetSnapshotID(), collector.nodeAddress, collector.delegateAddress)
		if err != nil {
			return fmt.Errorf("Error getting Snapshot voted proposals: %w", err)
		}

		for _, votedProposal := range votedProposals.Data.Votes {
			if votedProposal.Proposal.State == "active" {
				votesActiveProposals += 1
			} else {
				votesClosedProposals += 1
			}
		}

		return nil
	})

	// Get the number of live Snapshot proposals
	wg.Go(func() error {
		proposals, err := node.GetSnapshotProposals(collector.cfg.Smartnode.GetSnapshotApiDomain(), collector.cfg.Smartnode.GetSnapshotID(), "")
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

		return nil
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		log.Printf("%s\n", err.Error())
		return
	}

	channel <- prometheus.MustNewConstMetric(
		collector.votesActiveProposals, prometheus.GaugeValue, votesActiveProposals)
	channel <- prometheus.MustNewConstMetric(
		collector.votesClosedProposals, prometheus.GaugeValue, votesClosedProposals)
	channel <- prometheus.MustNewConstMetric(
		collector.activeProposals, prometheus.GaugeValue, activeProposals)
	channel <- prometheus.MustNewConstMetric(
		collector.closedProposals, prometheus.GaugeValue, closedProposals)
}
