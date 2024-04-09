package collectors

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/v2/shared/keys"
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

	// Context for graceful shutdowns
	ctx context.Context

	// The Smartnode service provider
	sp *services.ServiceProvider

	// The logger
	logger *slog.Logger

	// The thread-safe locker for the network state
	stateLocker *StateLocker
}

// Create a new BeaconCollector instance
func NewBeaconCollector(logger *log.Logger, ctx context.Context, sp *services.ServiceProvider, stateLocker *StateLocker) *BeaconCollector {
	subsystem := "beacon"
	sublogger := logger.With(slog.String(keys.RoutineKey, "Beacon Collector"))
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
		ctx:         ctx,
		sp:          sp,
		logger:      sublogger,
		stateLocker: stateLocker,
	}
}

// Write metric descriptions to the Prometheus channel
func (c *BeaconCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- c.activeSyncCommittee
	channel <- c.upcomingSyncCommittee
	channel <- c.upcomingProposals
	channel <- c.recentProposals
}

// Collect the latest metric values and pass them to Prometheus
func (c *BeaconCollector) Collect(channel chan<- prometheus.Metric) {
	// Get the latest state
	state := c.stateLocker.GetState()
	if state == nil {
		return
	}
	epoch := state.BeaconSlotNumber / state.BeaconConfig.SlotsPerEpoch

	// Get services
	bc := c.sp.GetBeaconClient()
	nodeAddress, hasNodeAddress := c.sp.GetWallet().GetAddress()

	activeSyncCommittee := float64(0)
	upcomingSyncCommittee := float64(0)
	upcomingProposals := float64(0)
	validatorIndices := []string{}
	recentProposalCount := float64(0)

	// Get sync committee duties
	if hasNodeAddress {
		for _, mpd := range state.MinipoolDetailsByNode[nodeAddress] {
			validator := state.ValidatorDetails[mpd.Pubkey]
			if validator.Exists {
				validatorIndices = append(validatorIndices, validator.Index)
			}
		}
	}

	if len(validatorIndices) > 0 {
		var wg errgroup.Group

		wg.Go(func() error {
			// Get current duties
			duties, err := bc.GetValidatorSyncDuties(c.ctx, validatorIndices, epoch)
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
			duties, err := bc.GetValidatorSyncDuties(c.ctx, validatorIndices, epoch+config.EpochsPerSyncCommitteePeriod)
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
			duties, err := bc.GetValidatorProposerDuties(c.ctx, validatorIndices, epoch)
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

		wg.Go(func() error {
			// check the latest finalized epoch for proposals:
			count, err := c.getProposedBlockCount(validatorIndices, bc, state.BeaconConfig.SlotsPerEpoch)
			if err != nil {
				c.logger.Error("Error getting recent proposed block count", log.Err(err))
				return err
			}
			recentProposalCount = count
			return nil
		})

		// Wait for data
		if err := wg.Wait(); err != nil {
			c.logger.Error(err.Error())
			return
		}
	}

	channel <- prometheus.MustNewConstMetric(
		c.activeSyncCommittee, prometheus.GaugeValue, activeSyncCommittee)
	channel <- prometheus.MustNewConstMetric(
		c.upcomingSyncCommittee, prometheus.GaugeValue, upcomingSyncCommittee)
	channel <- prometheus.MustNewConstMetric(
		c.upcomingProposals, prometheus.GaugeValue, upcomingProposals)
	channel <- prometheus.MustNewConstMetric(
		c.recentProposals, prometheus.GaugeValue, recentProposalCount)
}

func (c *BeaconCollector) getProposedBlockCount(validatorIndices []string, bc beacon.IBeaconClient, slotsPerEpoch uint64) (float64, error) {
	// Get the Beacon head
	head, err := bc.GetBeaconHead(c.ctx)
	if err != nil {
		c.logger.Error("Error getting Beacon chain head", log.Err(err))
		return 0, nil
	}

	// prepare for quick lookups in event of many validators:
	indexLookup := make(map[string]string, len(validatorIndices))
	for _, index := range validatorIndices {
		indexLookup[index] = index
	}
	latestSlot := head.FinalizedEpoch*slotsPerEpoch + (slotsPerEpoch - 1)

	// check each block in the most recent epoch for our validators:
	proposedBlockCount := float64(0)

	for slot := latestSlot; slot > latestSlot-slotsPerEpoch; slot-- {
		block, hasBlock, err := bc.GetBeaconBlockHeader(c.ctx, strconv.FormatUint(slot, 10))
		if err != nil {
			c.logger.Error("Error getting Beacon block", log.Err(err))
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
