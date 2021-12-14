package collectors

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/utils/rp"
	"log"

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

	// The Rocket Pool contract manager
	rp *rocketpool.RocketPool

	// The beacon client
	bc beacon.Client

	// The eth1 client
	ec *ethclient.Client

	// The node's address
	nodeAddress common.Address
}


// Create a new PerformanceCollector instance
func NewBeaconCollector(rp *rocketpool.RocketPool, bc beacon.Client, ec *ethclient.Client, nodeAddress common.Address) *BeaconCollector {
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
		rp: rp,
		bc: bc,
		ec: ec,
		nodeAddress: nodeAddress,
	}
}


// Write metric descriptions to the Prometheus channel
func (collector *BeaconCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- collector.activeSyncCommittee
	channel <- collector.upcomingSyncCommittee
}


// Collect the latest metric values and pass them to Prometheus
func (collector *BeaconCollector) Collect(channel chan<- prometheus.Metric) {

    // Sync
    var wg errgroup.Group

    activeSyncCommittee := float64(0)
	upcomingSyncCommittee := float64(0)

	// Get sync committee duties
    wg.Go(func() error {
		validatorIndices, err := rp.GetNodeValidatorIndices(collector.rp, collector.ec, collector.bc, collector.nodeAddress)
		if err != nil {
			return fmt.Errorf("Error getting validator indices: %w", err)
		}

		head, err := collector.bc.GetBeaconHead()
		if err != nil {
			return fmt.Errorf("Error getting beaconchain head: %w", err)
		}

		var wg2 errgroup.Group

		wg2.Go(func() error {
			// Get current duties
			duties, err := collector.bc.GetValidatorSyncDuties(validatorIndices, head.Epoch)
			if err != nil {
				return fmt.Errorf("Error getting sync duties: %w", err)
			}

			for _, duty := range duties {
				if duty {
					activeSyncCommittee ++
				}
			}

			return nil
		})

		wg2.Go(func() error {
			// Get epochs per sync committee period config to query next period
			config, err := collector.bc.GetEth2Config()
			if err != nil {
				return fmt.Errorf("Error getting ETH2 config: %w", err)
			}

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

		return wg2.Wait()
	})

    // Wait for data
    if err := wg.Wait(); err != nil {
        log.Printf("%s\n", err.Error())
        return
    }

    channel <- prometheus.MustNewConstMetric(
		collector.activeSyncCommittee, prometheus.GaugeValue, activeSyncCommittee)
	channel <- prometheus.MustNewConstMetric(
		collector.upcomingSyncCommittee, prometheus.GaugeValue, upcomingSyncCommittee)
}
