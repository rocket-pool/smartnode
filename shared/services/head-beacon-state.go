package services

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/types/eth2"
	"github.com/rocket-pool/smartnode/shared/types/eth2/generic"
)

// GetHeadBeaconState selects the latest beacon root already exposed by EIP-4788.
// The execution header's timestamp anchors proofs without waiting for finality
// or assuming that the next beacon slot has an execution payload.
func GetHeadBeaconState(bc beacon.Client, ec rocketpool.ExecutionClient) (eth2.BeaconState, uint64, error) {
	header, err := ec.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, 0, fmt.Errorf("error getting execution head: %w", err)
	}
	if header == nil || header.ParentBeaconRoot == nil || *header.ParentBeaconRoot == (common.Hash{}) {
		return nil, 0, fmt.Errorf("execution head has no parent beacon block root")
	}
	block, found, err := bc.GetBeaconBlockHeader(header.ParentBeaconRoot.Hex())
	if err != nil {
		return nil, 0, fmt.Errorf("error getting head proof beacon block: %w", err)
	}
	if !found {
		return nil, 0, fmt.Errorf("beacon block %s is not available", header.ParentBeaconRoot.Hex())
	}
	if block.Root != *header.ParentBeaconRoot {
		return nil, 0, fmt.Errorf("head proof beacon block root does not match execution head")
	}
	response, err := bc.GetBeaconStateSSZ(block.Slot)
	if err != nil {
		return nil, 0, fmt.Errorf("error getting head proof beacon state: %w", err)
	}
	beaconState, err := eth2.NewBeaconState(response.Data, response.Size, response.Fork)
	if err != nil {
		return nil, 0, err
	}
	if beaconState.GetSlot() != block.Slot {
		return nil, 0, fmt.Errorf("head proof beacon state slot does not match beacon block")
	}
	stateRoot, err := generic.SSZ.HashTreeRoot(beaconState)
	if err != nil {
		return nil, 0, fmt.Errorf("error hashing head proof beacon state: %w", err)
	}
	if common.Hash(stateRoot) != block.StateRoot {
		return nil, 0, fmt.Errorf("head proof beacon state root does not match beacon block; retry after beacon chain reorganisation")
	}
	return beaconState, header.Time, nil
}
