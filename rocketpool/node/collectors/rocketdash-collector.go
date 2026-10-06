package collectors

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/bindings/network"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/rocketpool/api/pdao"
	"github.com/rocket-pool/smartnode/shared/math"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/proposals"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Time to wait to make new RocketDash API calls
const hoursToWait float64 = 6

// Represents the collector for RocketDash metrics
type RocketDashCollector struct {
	// the number of active RocketDash proposals
	activeProposals *prometheus.Desc

	// the number of past RocketDash proposals
	closedProposals *prometheus.Desc

	// the number of votes on active RocketDash proposals
	votesActiveProposals *prometheus.Desc

	// the number of votes on closed RocketDash proposals
	votesClosedProposals *prometheus.Desc

	// The current node voting power on RocketDash
	nodeVotingPower *prometheus.Desc

	// The current delegate voting power on RocketDash
	delegateVotingPower *prometheus.Desc

	// The Rocket Pool Contract manager
	rp *rocketpool.RocketPool

	// The Rocket Pool config
	cfg *config.RocketPoolConfig

	// The Rocket Pool Execution Client manager
	ec *services.ExecutionClientManager

	// The Rocket Pool Beacon Client manager
	bc *services.BeaconClientManager

	// the node wallet address
	nodeAddress common.Address

	mutex sync.Mutex

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

// Create a new RocketDashCollector instance
func NewRocketDashCollector(rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, ec *services.ExecutionClientManager, bc *services.BeaconClientManager, nodeAddress common.Address) *RocketDashCollector {
	// Preserve metric names used by installed dashboards and alerts.
	subsystem := "snapshot"
	return &RocketDashCollector{
		activeProposals: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "proposals_active"),
			"The number of active RocketDash proposals",
			nil, nil,
		),
		closedProposals: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "proposals_closed"),
			"The number of closed RocketDash proposals",
			nil, nil,
		),
		votesActiveProposals: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "votes_active"),
			"The number of votes from user/delegate on active RocketDash proposals",
			nil, nil,
		),
		votesClosedProposals: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "votes_closed"),
			"The number of votes from user/delegate on closed RocketDash proposals",
			nil, nil,
		),
		nodeVotingPower: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "node_vp"),
			"The node current voting power on RocketDash",
			nil, nil,
		),
		delegateVotingPower: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "delegate_vp"),
			"The delegate current voting power on RocketDash",
			nil, nil,
		),
		rp:          rp,
		cfg:         cfg,
		ec:          ec,
		bc:          bc,
		nodeAddress: nodeAddress,
		logPrefix:   "RocketDash Collector",
	}
}

// Write metric descriptions to the Prometheus channel
func (collector *RocketDashCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- collector.activeProposals
	channel <- collector.closedProposals
	channel <- collector.votesActiveProposals
	channel <- collector.votesClosedProposals
	channel <- collector.nodeVotingPower
	channel <- collector.delegateVotingPower
}

// Collect the latest metric values and pass them to Prometheus
func (collector *RocketDashCollector) Collect(channel chan<- prometheus.Metric) {

	collector.mutex.Lock()
	defer collector.mutex.Unlock()
	if time.Since(collector.lastApiCallTimestamp).Hours() >= hoursToWait {
		if err := collector.refresh(); err != nil {
			collector.logError(err)
			return
		}
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
func (collector *RocketDashCollector) logError(err error) {
	fmt.Printf("[%s] %s\n", collector.logPrefix, err.Error())
}

func getVotingPower(propMgr *proposals.ProposalManager, blockNumber uint32, address common.Address) (float64, error) {
	// Get the total voting power
	totalDelegatedVP, _, _, err := propMgr.GetArtifactsForVoting(blockNumber, address)
	if err != nil {
		return 0, fmt.Errorf("error getting voting power: %w", err)
	}

	return math.WeiToEth(totalDelegatedVP), nil
}

func (collector *RocketDashCollector) refresh() error {
	var wg errgroup.Group
	var status api.SnapshotResponseStruct
	var propMgr *proposals.ProposalManager
	var blockNumber uint64
	var delegate common.Address
	wg.Go(func() error {
		var err error
		status, err = pdao.GetOffchainVotingStatus(collector.cfg, collector.rp, collector.nodeAddress, "")
		return err
	})
	wg.Go(func() error {
		var err error
		blockNumber, err = collector.ec.BlockNumber(context.Background())
		return err
	})
	wg.Go(func() error {
		var err error
		propMgr, err = proposals.NewProposalManager(nil, collector.cfg, collector.rp, collector.bc)
		return err
	})
	wg.Go(func() error {
		var err error
		delegate, err = network.GetCurrentVotingDelegate(collector.rp, collector.nodeAddress, nil)
		return err
	})
	if err := wg.Wait(); err != nil {
		return err
	}
	nodePower, err := getVotingPower(propMgr, uint32(blockNumber), collector.nodeAddress)
	if err != nil {
		return fmt.Errorf("getting node voting power: %w", err)
	}
	delegatePower := float64(0)
	if delegate == collector.nodeAddress {
		delegatePower = nodePower
	} else if delegate != (common.Address{}) {
		delegatePower, err = getVotingPower(propMgr, uint32(blockNumber), delegate)
		if err != nil {
			return fmt.Errorf("getting delegate voting power: %w", err)
		}
	}

	// Only publish a complete refresh; failed requests are retried on the next scrape.
	collector.cachedNodeVotingPower = nodePower
	collector.cachedDelegateVotingPower = delegatePower
	collector.collectProposalsAndVotes(status)
	collector.lastApiCallTimestamp = time.Now()
	return nil
}

func (collector *RocketDashCollector) collectProposalsAndVotes(status api.SnapshotResponseStruct) {
	collector.cachedActiveProposals = 0
	collector.cachedClosedProposals = 0
	collector.cachedVotesActiveProposals = 0
	collector.cachedVotesClosedProposals = 0
	voted := make(map[string]bool, len(status.ProposalVotes))
	for _, vote := range status.ProposalVotes {
		voted[vote.Proposal.Id] = true
	}
	for _, proposal := range status.ActiveSnapshotProposals {
		switch proposal.State {
		case "active":
			collector.cachedActiveProposals++
			if voted[proposal.Id] {
				collector.cachedVotesActiveProposals++
			}
		case "closed":
			collector.cachedClosedProposals++
			if voted[proposal.Id] {
				collector.cachedVotesClosedProposals++
			}
		}
	}
}
