package debug

import (
	"fmt"
	"math/bits"

	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/v3/crypto/hash/htr"
	state_native "github.com/prysmaticlabs/prysm/v5/beacon-chain/state/state-native"
	"github.com/prysmaticlabs/prysm/v5/beacon-chain/state/stateutil"
	"github.com/prysmaticlabs/prysm/v5/encoding/ssz"
	eth "github.com/prysmaticlabs/prysm/v5/proto/prysm/v1alpha1"
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

	beaconState.

	stateNative, ok := beaconState.(*state_native.BeaconState)
	if !ok {
		return nil, fmt.Errorf("failed while casting to state_native.BeaconState")
	}



	roots, err := stateutil.OptimizedValidatorRoots(stateNative.Validators())
	if err != nil {
		return nil, fmt.Errorf("failed getting validator roots: %w", err)
	}

	fmt.Println("%w", roots[1100000])

	// Return response
	return &response, nil
}

// ValidatorRegistryProof computes the merkle proof for a validator at a specific index
// in the validator registry.
func ValidatorRegistryProof(validators []*eth.Validator, index uint64) ([][32]byte, error) {
	validatorFieldRoots := 8
	if index >= uint64(len(validators)) {
		return nil, errors.New("validator index out of bounds")
	}

	// First get all validator roots
	roots, err := stateutil.OptimizedValidatorRoots(validators)
	if err != nil {
		return nil, errors.Wrap(err, "could not get validator roots")
	}

	// Since each validator has 8 fields, we need to get the root
	// at the correct position in the flattened array
	validatorPosition := index * uint64(validatorFieldRoots)

	// Get the individual validator root by hashing its field roots
	subRoots := make([][32]byte, validatorFieldRoots)
	copy(subRoots, roots[validatorPosition:validatorPosition+validatorFieldRoots])

	for i := 0; i < validatorTreeDepth; i++ {
		subRoots = htr.VectorizedSha256(subRoots)
	}

	// Create merkle tree from all validator roots
	depth := bits.Len64(uint64(len(validators) - 1)) // Calculate required depth
	tree := make([][32]byte, len(validators))

	// Fill tree with validator roots
	for i := 0; i < len(validators); i++ {
		pos := i * validatorFieldRoots
		vRoots := roots[pos : pos+validatorFieldRoots]
		// Hash up individual validator roots
		for j := 0; j < validatorTreeDepth; j++ {
			vRoots = htr.VectorizedSha256(vRoots)
		}
		tree[i] = vRoots[0]
	}

	// Generate proof
	proof := make([][32]byte, depth)
	tmp := tree
	for h := 0; h < depth; h++ {
		idx := (index >> h) ^ 1
		if idx < uint64(len(tmp)) {
			proof[h] = tmp[idx]
		}

		// Move up one level in the tree
		newSize := (len(tmp) + 1) / 2
		newTmp := make([][32]byte, newSize)
		for i := 0; i < len(tmp)-1; i += 2 {
			concat := append(tmp[i][:], tmp[i+1][:]...)
			newTmp[i/2] = ssz.HashWithDefaultHasher(concat)
		}
		// Handle odd number of elements
		if len(tmp)%2 == 1 {
			concat := append(tmp[len(tmp)-1][:], make([]byte, 32)...)
			newTmp[len(newTmp)-1] = ssz.HashWithDefaultHasher(concat)
		}
		tmp = newTmp
	}

	return proof, nil
}
