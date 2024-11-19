package debug

import (
	"fmt"
	"math"

	state_native "github.com/prysmaticlabs/prysm/v5/beacon-chain/state/state-native"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/state/stateutil"
	"github.com/prysmaticlabs/prysm/v5/container/trie"
	"github.com/prysmaticlabs/prysm/v5/encoding/ssz"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getBeaconStateForSlot(c *cli.Context, slot uint64) (*api.BeaconStateResponse, error) {
	// Create a new response
	response := api.BeaconStateResponse{}

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	// Get beacon state
	beaconState, err := bc.GetBeaconState(slot)
	if err != nil {
		return nil, err
	}

	stateNative, ok := beaconState.(*state_native.BeaconState)
	if !ok {
		return nil, fmt.Errorf("failed while casting to state_native.BeaconState")
	}

	roots, err := stateutil.OptimizedValidatorRoots(stateNative.Validators())
	if err != nil {
		return nil, fmt.Errorf("failed getting validator roots: %w", err)
	}

	//trie.GenerateTrieFromItems(roots, ssz.Depth(roots))

	fmt.Println("%w", roots[1100000])

	// Return response
	return &response, nil
}
