package collectors

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rocket-pool/smartnode/bindings/dao/protocol"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/types"
	"golang.org/x/sync/errgroup"
)

// Represents the collector for onchain governance metrics
type GovernanceCollector struct {
	// the number of active onchain proposals pending
	onchainPending *prometheus.Desc

	// the number of active onchain proposals in Phase 1
	onchainPhase1 *prometheus.Desc

	// the number of active onchain proposals in Phase 2
	onchainPhase2 *prometheus.Desc

	// the number of closed onchain proposals
	onchainClosed *prometheus.Desc

	// The Rocket Pool Contract manager
	rp *rocketpool.RocketPool

	// Prefix for logging
	logPrefix string
}

// Create a new SnapshotCollector instance
func NewGovernanceCollector(rp *rocketpool.RocketPool) *GovernanceCollector {
	subsystem := "governance"
	return &GovernanceCollector{
		onchainPending: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "onchain_pending"),
			"The number of pending onchain proposals",
			nil, nil,
		),
		onchainPhase1: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "onchain_phase1"),
			"The number of onchain proposals in Phase 1",
			nil, nil,
		),
		onchainPhase2: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "onchain_phase2"),
			"The number of onchain proposals in Phase 2",
			nil, nil,
		),
		onchainClosed: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "onchain_closed"),
			"The number of closed onchain proposals",
			nil, nil,
		),
		rp:        rp,
		logPrefix: "Governance Collector",
	}
}

// Write metric descriptions to the Prometheus channel
func (collector *GovernanceCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- collector.onchainPending
	channel <- collector.onchainPhase1
	channel <- collector.onchainPhase2
	channel <- collector.onchainClosed
}

// Collect the latest metric values and pass them to Prometheus
func (collector *GovernanceCollector) Collect(channel chan<- prometheus.Metric) {

	// Sync
	var wg errgroup.Group
	var err error
	var onchainProposals []protocol.ProtocolDaoProposalDetails
	onchainPending := float64(0)
	onchainPhase1 := float64(0)
	onchainPhase2 := float64(0)
	onchainClosed := float64(0)

	// Get onchain proposals
	wg.Go(func() error {
		onchainProposals, err = protocol.GetProposals(collector.rp, nil)
		if err != nil {
			return fmt.Errorf("error fetching onchain proposals: %w", err)
		}
		for _, proposal := range onchainProposals {
			if proposal.State == types.ProtocolDaoProposalState_Pending {
				onchainPending += 1
			} else if proposal.State == types.ProtocolDaoProposalState_ActivePhase1 {
				onchainPhase1 += 1
			} else if proposal.State == types.ProtocolDaoProposalState_ActivePhase2 {
				onchainPhase2 += 1
			} else {
				onchainClosed += 1
			}
		}
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		collector.logError(err)
		return
	}

	channel <- prometheus.MustNewConstMetric(
		collector.onchainPending, prometheus.GaugeValue, onchainPending)
	channel <- prometheus.MustNewConstMetric(
		collector.onchainPhase1, prometheus.GaugeValue, onchainPhase1)
	channel <- prometheus.MustNewConstMetric(
		collector.onchainPhase2, prometheus.GaugeValue, onchainPhase2)
}

// Log error messages
func (collector *GovernanceCollector) logError(err error) {
	fmt.Printf("[%s] %s\n", collector.logPrefix, err.Error())
}
