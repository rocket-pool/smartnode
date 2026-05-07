package megapool

import (
	"fmt"
	"math"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

const (
	farFutureEpoch                   uint64 = math.MaxUint64
	perEpochActivationExitChurnLimit uint64 = 256_000_000_000
)

// getBeaconWithdrawalQueueEstimate estimates how long the current beacon-chain
// exit queue will take to be processed.
func getBeaconWithdrawalQueueEstimate(c *cli.Command) (*api.BeaconWithdrawalQueueEstimateResponse, error) {
	if err := services.RequireBeaconClientSynced(c); err != nil {
		return nil, err
	}
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	eth2Config, err := bc.GetEth2Config()
	if err != nil {
		return nil, fmt.Errorf("error getting eth2 config: %w", err)
	}

	head, err := bc.GetBeaconHead()
	if err != nil {
		return nil, fmt.Errorf("error getting beacon head: %w", err)
	}
	currentEpoch := head.Epoch

	validators, err := bc.GetAllValidators()
	if err != nil {
		return nil, fmt.Errorf("error getting validator set: %w", err)
	}

	// Walk the validator set once and collect the effective balance of validators currently waiting to exit
	var exitQueueGwei uint64
	for _, v := range validators {

		// In the exit queue if exit_epoch is set and still in the future.
		if v.ExitEpoch != farFutureEpoch && v.ExitEpoch > currentEpoch {
			exitQueueGwei += v.EffectiveBalance
		}
	}

	churnPerEpochGwei := perEpochActivationExitChurnLimit

	// epochs needed to process the queue, rounded up
	var estimatedEpochs uint64
	if churnPerEpochGwei > 0 && exitQueueGwei > 0 {
		estimatedEpochs = (exitQueueGwei + churnPerEpochGwei - 1) / churnPerEpochGwei
	}
	estimatedSeconds := estimatedEpochs * eth2Config.SecondsPerEpoch

	return &api.BeaconWithdrawalQueueEstimateResponse{
		ExitQueueGwei:         exitQueueGwei,
		ChurnPerEpochGwei:     churnPerEpochGwei,
		SecondsPerEpoch:       eth2Config.SecondsPerEpoch,
		EstimatedQueueEpochs:  estimatedEpochs,
		EstimatedQueueSeconds: estimatedSeconds,
	}, nil
}
