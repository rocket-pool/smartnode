package collectors

import (
	"fmt"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/shared/services/beacon"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"golang.org/x/sync/errgroup"
)

// Represents the collector for the beaconchain metrics
type BeaconCollector struct {
	// The number of this node's validators is currently in a sync committee
	activeSyncCommittee *prometheus.Desc

	// The number of this node's validators on the next sync committee
	upcomingSyncCommittee *prometheus.Desc

	// The number of upcoming proposals for this node's validators
	upcomingProposals *prometheus.Desc

	// The number of recent proposals for this node's validators
	recentProposals *prometheus.Desc

	// The Rocket Pool contract manager
	rp *rocketpool.RocketPool

	// The beacon client
	bc beacon.Client

	// The execution client
	ec rocketpool.ExecutionClient

	// The node's address
	nodeAddress common.Address

	// The thread-safe locker for the network state
	stateLocker *StateLocker

	// Prefix for logging
	logPrefix string
}

// Create a new BeaconCollector instance
func NewBeaconCollector(rp *rocketpool.RocketPool, bc beacon.Client, ec rocketpool.ExecutionClient, nodeAddress common.Address, stateLocker *StateLocker) *BeaconCollector {
	subsystem := "beacon"
	return &BeaconCollector{
		activeSyncCommittee: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "active_sync_committee"),
			"The number of validators on a current sync committee",
			nil, nil,
		),
		upcomingSyncCommittee: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "upcoming_sync_committee"),
			"The number of validators on the next sync committee",
			nil, nil,
		),
		upcomingProposals: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "upcoming_proposals"),
			"The number of proposals assigned to validators in this epoch and the next",
			nil, nil,
		),
		recentProposals: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "recent_proposals"),
			"The number of block proposals made by validators in the most recent finalized epoch",
			nil, nil,
		),
		rp:          rp,
		bc:          bc,
		ec:          ec,
		nodeAddress: nodeAddress,
		stateLocker: stateLocker,
		logPrefix:   "Beacon Collector",
	}
}

// Write metric descriptions to the Prometheus channel
func (collector *BeaconCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- collector.activeSyncCommittee
	channel <- collector.upcomingSyncCommittee
	channel <- collector.upcomingProposals
	channel <- collector.recentProposals
}

// Collect the latest metric values and pass them to Prometheus
func (collector *BeaconCollector) Collect(channel chan<- prometheus.Metric) {
	// Get the latest state
	state := collector.stateLocker.GetState()
	if state == nil {
		return
	}

	var wg errgroup.Group
	activeSyncCommittee := float64(0)
	upcomingSyncCommittee := float64(0)
	upcomingProposals := float64(0)

	var validatorIndices []string
	var head beacon.BeaconHead

	// Get sync committee duties
	for _, mpd := range state.MinipoolDetailsByNode[collector.nodeAddress] {
		validator := state.ValidatorDetails[mpd.Pubkey]
		if validator.Exists {
			validatorIndices = append(validatorIndices, validator.Index)
		}
	}

	head, err := collector.bc.GetBeaconHead()
	if err != nil {
		collector.logError(fmt.Errorf("error getting Beacon chain head: %w", err))
		return
	}

	wg.Go(func() error {
		// Get current duties
		duties, err := collector.bc.GetValidatorSyncDuties(validatorIndices, head.Epoch)
		if err != nil {
			return fmt.Errorf("Error getting sync duties: %w", err)
		}

		for _, duty := range duties {
			if duty {
				activeSyncCommittee++
			}
		}

		return nil
	})

	wg.Go(func() error {
		// Get epochs per sync committee period config to query next period
		config := state.BeaconConfig

		// Get upcoming duties
		duties, err := collector.bc.GetValidatorSyncDuties(validatorIndices, head.Epoch+config.EpochsPerSyncCommitteePeriod)
		if err != nil {
			return fmt.Errorf("Error getting sync duties: %w", err)
		}

		for _, duty := range duties {
			if duty {
				upcomingSyncCommittee++
			}
		}

		return nil
	})

	wg.Go(func() error {
		// Get proposals in this epoch
		duties, err := collector.bc.GetValidatorProposerDuties(validatorIndices, head.Epoch)
		if err != nil {
			return fmt.Errorf("Error getting proposer duties: %w", err)
		}

		for _, duty := range duties {
			upcomingProposals += float64(duty)
		}

		// TODO: this seems to be illegal according to the official spec:
		// https://eth2book.info/altair/annotated-spec/#compute_proposer_index
		/*
			// Get proposals in the next epoch
			duties, err = collector.bc.GetValidatorProposerDuties(validatorIndices, head.Epoch + 1)
			if err != nil {
				return fmt.Errorf("Error getting proposer duties: %w", err)
			}

			for _, duty := range duties {
				upcomingProposals += float64(duty)
			}
		*/

		return nil
	})

	recentProposalCount := float64(0)
	wg.Go(func() error {
		// check the latest finalized epoch for proposals:
		count, err := collector.getProposedBlockCount(validatorIndices, head, state.BeaconConfig.SlotsPerEpoch)
		if err != nil {
			collector.logError(fmt.Errorf("error getting recent proposed block count: %w", err))
			return err
		}
		recentProposalCount = count
		return nil
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		collector.logError(err)
		return
	}

	channel <- prometheus.MustNewConstMetric(
		collector.activeSyncCommittee, prometheus.GaugeValue, activeSyncCommittee)
	channel <- prometheus.MustNewConstMetric(
		collector.upcomingSyncCommittee, prometheus.GaugeValue, upcomingSyncCommittee)
	channel <- prometheus.MustNewConstMetric(
		collector.upcomingProposals, prometheus.GaugeValue, upcomingProposals)
	channel <- prometheus.MustNewConstMetric(
		collector.recentProposals, prometheus.GaugeValue, recentProposalCount)
}

func (collector *BeaconCollector) getProposedBlockCount(validatorIndices []string, head beacon.BeaconHead, slotsPerEpoch uint64) (float64, error) {
	// prepare for quick lookups in event of many validators:
	indexLookup := make(map[string]string, len(validatorIndices))
	for _, index := range validatorIndices {
		indexLookup[index] = index
	}
	latestSlot := head.FinalizedEpoch*slotsPerEpoch + (slotsPerEpoch - 1)

	// check each block in the most recent epoch for our validators:
	proposedBlockCount := float64(0)

	for slot := latestSlot; slot > latestSlot-slotsPerEpoch; slot-- {
		block, hasBlock, err := collector.bc.GetBeaconBlockHeader(strconv.FormatUint(slot, 10))
		if err != nil {
			collector.logError(fmt.Errorf("error getting beacon block: %w", err))
			continue
		}
		if !hasBlock {
			continue
		}
		if _, ok := indexLookup[block.ProposerIndex]; !ok {
			continue
		}
		proposedBlockCount++
	}
	return proposedBlockCount, nil
}

// Log error messages
func (collector *BeaconCollector) logError(err error) {
	fmt.Printf("[%s] %s\n", collector.logPrefix, err.Error())
}
