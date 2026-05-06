package megapool

import (
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// getLatestBlockWithdrawals returns the validator withdrawals processed in the
// latest beacon block that contains an execution payload. If the head slot has
// no execution payload (e.g. it was a missed slot), it walks backwards a few
// slots until it finds one.
func getLatestBlockWithdrawals(c *cli.Command) (*api.LatestBlockWithdrawalsResponse, error) {
	if err := services.RequireBeaconClientSynced(c); err != nil {
		return nil, err
	}
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	blockToRequest := "head"
	const maxAttempts = 8
	var (
		block  beacon.BeaconBlock
		exists bool
	)
	for attempts := 0; attempts < maxAttempts; attempts++ {
		block, exists, err = bc.GetBeaconBlock(blockToRequest)
		if err != nil {
			return nil, fmt.Errorf("error getting beacon block %s: %w", blockToRequest, err)
		}
		if exists && block.HasExecutionPayload {
			break
		}
		// Walk backwards by slot number; if we don't yet have one, fall back.
		var nextSlot uint64
		if block.Slot > 0 {
			nextSlot = block.Slot - 1
		} else if attempts == 0 {
			// We never resolved the head, give up
			return nil, fmt.Errorf("could not resolve the head beacon block")
		}
		if attempts == maxAttempts-1 {
			return nil, fmt.Errorf("could not find a beacon block with an execution payload after %d attempts", maxAttempts)
		}
		blockToRequest = fmt.Sprintf("%d", nextSlot)
	}

	response := &api.LatestBlockWithdrawalsResponse{
		Slot:        block.Slot,
		BlockNumber: block.ExecutionBlockNumber,
		Withdrawals: block.Withdrawals,
	}
	return response, nil
}
