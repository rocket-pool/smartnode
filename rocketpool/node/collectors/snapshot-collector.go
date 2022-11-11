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
	// the number of open Snashot proposals
	openProposals *prometheus.Desc

	// the number of past Snapshot proposals
	closedProposals *prometheus.Desc

	// the number of votes on open Snapshot proposals
	votesOpenProposals *prometheus.Desc

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
		openProposals: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "proposals_open"),
			"The number of open Snapshot proposals",
			nil, nil,
		),
		closedProposals: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "proposals_closed"),
			"The number of close Snapshot proposals",
			nil, nil,
		),
		votesOpenProposals: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "votes_open"),
			"The number of votes on open Snapshot proposals",
			nil, nil,
		),
		votesClosedProposals: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "votes_closed"),
			"The number of votes on closed Snapshot proposals",
			nil, nil,
		),
		cfg:             cfg,
		nodeAddress:     nodeAddress,
		delegateAddress: delegateAddres,
	}
}

// Write metric descriptions to the Prometheus channel
func (collector *SnapshotCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- collector.openProposals
	channel <- collector.closedProposals
	channel <- collector.votesOpenProposals
	channel <- collector.votesClosedProposals
}

// Collect the latest metric values and pass them to Prometheus
func (collector *SnapshotCollector) Collect(channel chan<- prometheus.Metric) {

	// Sync
	var wg errgroup.Group
	openProposals := float64(0)
	closedProposals := float64(0)
	votesOpenProposals := float64(0)
	votesClosedProposals := float64(0)

	// Get the number of votes on Snapshot proposals
	wg.Go(func() error {
		votedProposals, err := node.GetSnapshotVotedProposals(collector.cfg.Smartnode.GetSnapshotApiDomain(), collector.cfg.Smartnode.GetSnapshotID(), collector.nodeAddress, collector.delegateAddress)
		if err != nil {
			return fmt.Errorf("Error getting Snapshot voted proposals: %w", err)
		}

		for _, votedProposal := range votedProposals.Data.Votes {
			if votedProposal.Proposal.State != "closed" {
				votesClosedProposals += 1
			} else {
				votesOpenProposals += 1
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
			if proposal.State == "open" {
				openProposals += 1
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
		collector.votesOpenProposals, prometheus.GaugeValue, votesOpenProposals)
	channel <- prometheus.MustNewConstMetric(
		collector.votesClosedProposals, prometheus.GaugeValue, votesClosedProposals)
	channel <- prometheus.MustNewConstMetric(
		collector.openProposals, prometheus.GaugeValue, openProposals)
	channel <- prometheus.MustNewConstMetric(
		collector.closedProposals, prometheus.GaugeValue, closedProposals)
}
