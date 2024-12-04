package debug

import (
	"bytes"
	"fmt"

	"math/bits"

	"github.com/pkg/errors"
	state_native "github.com/prysmaticlabs/prysm/v5/beacon-chain/state/state-native"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/state/stateutil"
	"github.com/prysmaticlabs/prysm/v5/crypto/hash"
	"github.com/prysmaticlabs/prysm/v5/crypto/hash/htr"
	ethpb "github.com/prysmaticlabs/prysm/v5/proto/prysm/v1alpha1"
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
	proof, err := ValidatorProof(stateNative.Validators(), 1)
	if err != nil {
		return nil, fmt.Errorf("error calculating validator proof")
	}

	val, err := stateNative.ValidatorAtIndex(1)
	if err != nil {
		return nil, fmt.Errorf("error getting validator by index")
	}

	root, err := stateutil.ValidatorRegistryRoot(stateNative.Validators())

	verified, err := VerifyValidatorProof(val, 1, proof, root)

	fmt.Println("%w", verified)

	// Return response
	return &response, nil
}

// ValidatorProof computes the merkle proof for a validator at a specific index
// in the validator registry.
func ValidatorProof(validators []*ethpb.Validator, index uint64) ([][32]byte, error) {
	if index >= uint64(len(validators)) {
		return nil, errors.New("validator index out of bounds")
	}

	// First get all validator roots
	roots, err := stateutil.OptimizedValidatorRoots(validators)
	if err != nil {
		return nil, errors.Wrap(err, "could not get validator roots")
	}

	// Create merkle tree from all validator roots
	depth := bits.Len64(uint64(len(validators) - 1)) // Calculate required depth

	// Generate proof
	proof := make([][32]byte, depth)
	tmp := roots
	for h := 0; h < depth; h++ {
		// Get the sibling index at height "h"
		idx := (index >> h) ^ 1
		if idx < uint64(len(tmp)) {
			proof[h] = tmp[idx]
		}

		// Move up one level in the tree
		newSize := (len(tmp) + 1) / 2
		newTmp := make([][32]byte, newSize)
		for i := 0; i < len(tmp)-1; i += 2 {
			concat := append(tmp[i][:], tmp[i+1][:]...)
			newTmp[i/2] = hash.Hash(concat)
		}
		// Handle odd number of elements
		if len(tmp)%2 == 1 {
			concat := append(tmp[len(tmp)-1][:], make([]byte, 32)...)
			newTmp[len(newTmp)-1] = hash.Hash(concat)
		}
		tmp = newTmp
	}

	return proof, nil
}
