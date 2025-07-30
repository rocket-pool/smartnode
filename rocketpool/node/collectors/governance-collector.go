package collectors

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rocket-pool/smartnode/bindings/dao/protocol"
	"github.com/rocket-pool/smartnode/bindings/network"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/rocket-pool/smartnode/rocketpool/api/pdao"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/contracts"
	"github.com/rocket-pool/smartnode/shared/services/proposals"
	"github.com/rocket-pool/smartnode/shared/types/api"
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

	// the number of active onchain proposals pending
	onchainPending *prometheus.Desc

	// the number of active onchain proposals in Phase 1
	onchainPhase1 *prometheus.Desc

	// the number of active onchain proposals in Phase 2
	onchainPhase2 *prometheus.Desc

	// The current node voting power on Snapshot
	nodeVotingPower *prometheus.Desc

	// The current delegate voting power on Snapshot
	delegateVotingPower *prometheus.Desc

	// The Rocket Pool Contract manager
	rp *rocketpool.RocketPool

	// The Rocket Pool config
	cfg *config.RocketPoolConfig

	// The Rocket Pool Execution Client manager
	ec *services.ExecutionClientManager

	// The Rocket Pool Beacon Client manager
	bc *services.BeaconClientManager

	// The RocketSignerRegistry Contract
	reg *contracts.RocketSignerRegistry

	// the node wallet address
	nodeAddress common.Address

	// the signalling address
	signallingAddress common.Address

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
func NewSnapshotCollector(rp *rocketpool.RocketPool, cfg *config.RocketPoolConfig, ec *services.ExecutionClientManager, bc *services.BeaconClientManager, reg *contracts.RocketSignerRegistry, nodeAddress common.Address, signallingAddress common.Address) *SnapshotCollector {
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
		nodeVotingPower: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "node_vp"),
			"The node current voting power on Snapshot",
			nil, nil,
		),
		delegateVotingPower: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "delegate_vp"),
			"The delegate current voting power on Snapshot",
			nil, nil,
		),
		rp:                rp,
		cfg:               cfg,
		ec:                ec,
		bc:                bc,
		reg:               reg,
		nodeAddress:       nodeAddress,
		signallingAddress: signallingAddress,
		logPrefix:         "Snapshot Collector",
	}
}

// Write metric descriptions to the Prometheus channel
func (collector *SnapshotCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- collector.activeProposals
	channel <- collector.closedProposals
	channel <- collector.votesActiveProposals
	channel <- collector.votesClosedProposals
	channel <- collector.onchainPending
	channel <- collector.onchainPhase1
	channel <- collector.onchainPhase2
	channel <- collector.nodeVotingPower
	channel <- collector.delegateVotingPower
}

// Collect the latest metric values and pass them to Prometheus
func (collector *SnapshotCollector) Collect(channel chan<- prometheus.Metric) {

	// Sync
	var wg errgroup.Group
	var err error
	var propMgr *proposals.ProposalManager
	var blockNumber uint64
	var onchainVotingDelegate common.Address
	var isVotingInitialized bool
	var onchainProposals []protocol.ProtocolDaoProposalDetails
	onchainPending := float64(0)
	onchainPhase1 := float64(0)
	onchainPhase2 := float64(0)
	activeProposals := float64(0)
	closedProposals := float64(0)
	blankAddress := common.Address{}

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

	// Get latest block number
	wg.Go(func() error {
		if time.Since(collector.lastApiCallTimestamp).Hours() >= hoursToWait {
			blockNumber, err = collector.ec.BlockNumber(context.Background())
			if err != nil {
				return fmt.Errorf("Error getting block number: %w", err)
			}
		}
		return nil

	})

	// Get the propMgr
	wg.Go(func() error {
		if time.Since(collector.lastApiCallTimestamp).Hours() >= hoursToWait {
			propMgr, err = proposals.NewProposalManager(nil, collector.cfg, collector.rp, collector.bc)
			if err != nil {
				return fmt.Errorf("Error getting the prop manager: %w", err)
			}
		}
		return nil
	})

	// Get the node onchain voting delegate
	wg.Go(func() error {
		if time.Since(collector.lastApiCallTimestamp).Hours() >= hoursToWait {
			onchainVotingDelegate, err = network.GetCurrentVotingDelegate(collector.rp, collector.nodeAddress, nil)
			if err != nil {
				return fmt.Errorf("Error getting the on-chain voting delegate: %w", err)
			}
		}
		return err
	})

	// Get Voting Initialized status
	wg.Go(func() error {
		if time.Since(collector.lastApiCallTimestamp).Hours() >= hoursToWait {
			isVotingInitialized, err = network.GetVotingInitialized(collector.rp, collector.nodeAddress, nil)
			if err != nil {
				return fmt.Errorf("Error checking if voting is initialized: %w", err)
			}
		}
		return err
	})

	// Get onchain proposals
	wg.Go(func() error {
		if time.Since(collector.lastApiCallTimestamp).Hours() >= hoursToWait {
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
				}
			}
		}
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		collector.logError(err)
		return
	}

	// Check if sufficient time has passed and voting is initialized
	if time.Since(collector.lastApiCallTimestamp).Hours() >= hoursToWait && isVotingInitialized {
		// Get voting power for the node
		nodeVotingPower, err := getVotingPower(propMgr, uint32(blockNumber), collector.nodeAddress)
		if err != nil {
			collector.logError(fmt.Errorf("error getting node voting power: %w", err))
			collector.cachedNodeVotingPower = 0
		} else {
			collector.cachedNodeVotingPower = nodeVotingPower
		}
		// Get voting power for the delegate
		delegateVotingPower, err := getVotingPower(propMgr, uint32(blockNumber), onchainVotingDelegate)
		if err != nil {
			collector.logError(fmt.Errorf("error getting delegate voting power: %w", err))
			collector.cachedDelegateVotingPower = 0
		} else {
			collector.cachedDelegateVotingPower = delegateVotingPower
		}
	}

	// Get the number of votes on Snapshot proposals
	if time.Since(collector.lastApiCallTimestamp).Hours() >= hoursToWait {
		// Check if there is a delegate voting on behalf of the node
		if onchainVotingDelegate != blankAddress || onchainVotingDelegate != collector.nodeAddress {
			delegateSignallingAddress, err := collector.reg.NodeToSigner(&bind.CallOpts{}, onchainVotingDelegate)
			if err != nil {
				collector.logError(fmt.Errorf("Error getting the signalling address: %w", err))
			}
			votedProposals, err := pdao.GetSnapshotVotedProposals(collector.cfg.Smartnode.GetSnapshotApiDomain(), collector.cfg.Smartnode.GetSnapshotID(), onchainVotingDelegate, delegateSignallingAddress)
			if err != nil {
				collector.logError(fmt.Errorf("Error getting Snapshot voted proposals: %w", err))
			}
			collector.collectVotes(votedProposals)
		} else {
			votedProposals, err := pdao.GetSnapshotVotedProposals(collector.cfg.Smartnode.GetSnapshotApiDomain(), collector.cfg.Smartnode.GetSnapshotID(), collector.nodeAddress, collector.signallingAddress)
			if err != nil {
				collector.logError(fmt.Errorf("Error getting Snapshot voted proposals: %w", err))
			}
			collector.collectVotes(votedProposals)
		}
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
		collector.onchainPending, prometheus.GaugeValue, onchainPending)
	channel <- prometheus.MustNewConstMetric(
		collector.onchainPhase1, prometheus.GaugeValue, onchainPhase1)
	channel <- prometheus.MustNewConstMetric(
		collector.onchainPhase2, prometheus.GaugeValue, onchainPhase2)
	channel <- prometheus.MustNewConstMetric(
		collector.nodeVotingPower, prometheus.GaugeValue, collector.cachedNodeVotingPower)
	channel <- prometheus.MustNewConstMetric(
		collector.delegateVotingPower, prometheus.GaugeValue, collector.cachedDelegateVotingPower)
}

// Log error messages
func (collector *SnapshotCollector) logError(err error) {
	fmt.Printf("[%s] %s\n", collector.logPrefix, err.Error())
}

func getVotingPower(propMgr *proposals.ProposalManager, blockNumber uint32, address common.Address) (float64, error) {
	// Get the total voting power
	totalDelegatedVP, _, _, err := propMgr.GetArtifactsForVoting(blockNumber, address)
	if err != nil {
		return 0, fmt.Errorf("error getting voting power: %w", err)
	}

	return eth.WeiToEth(totalDelegatedVP), nil
}

func (collector *SnapshotCollector) collectVotes(votedProposals *api.SnapshotVotedProposals) {
	handledProposals := map[string]bool{}
	votesActiveProposals := float64(0)
	votesClosedProposals := float64(0)

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
