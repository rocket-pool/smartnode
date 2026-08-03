package beacon

import (
	"context"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// The subset of the Execution Client needed to resolve a payload hash into a block number.
// Kept minimal so that Execution Client managers, the raw client interface, and test stubs
// all satisfy it without any of them having to be imported here.
type ExecutionHeaderSource interface {
	// HeaderByHash returns the block header with the given hash.
	HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error)
}

// ResolveExecutionBlockNumber returns the Execution Layer block number associated with a Beacon
// block.
//
// Pre-Gloas blocks embed the number in the execution payload. Gloas (EIP-7732) blocks no longer
// carry the payload at all - the body only has a bid committing to the payload's block hash - so
// the number has to be resolved through the Execution Client.
func ResolveExecutionBlockNumber(ctx context.Context, ec ExecutionHeaderSource, block BeaconBlock) (uint64, bool, error) {

	// Pre-Gloas: the payload was embedded in the block, so the number is already known.
	if block.ExecutionBlockNumber != 0 {
		return block.ExecutionBlockNumber, true, nil
	}

	// No payload association at all
	if block.ExecutionBlockHash == (common.Hash{}) {
		return 0, false, nil
	}

	header, err := ec.HeaderByHash(ctx, block.ExecutionBlockHash)
	if err != nil {
		if errors.Is(err, ethereum.NotFound) {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("error resolving EL block for payload hash %s at slot %d: %w", block.ExecutionBlockHash.Hex(), block.Slot, err)
	}
	if header == nil || header.Number == nil {
		return 0, false, nil
	}

	return header.Number.Uint64(), true, nil
}
