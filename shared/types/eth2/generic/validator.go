package generic

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/utils/math"
)

const beaconStateValidatorWithdrawalCredentialsPubkeyGeneralizedIndex uint64 = 4 // Container with 8 fields, so gid 8 is the first field. We want the parent of 1st field, so gid 8 / 2 = 4
const BeaconStateValidatorWithdrawableEpochGeneralizedIndex uint64 = 15          // Container with 8 fields, so gid 8 is the first field. We want the 8th field, so gid 8 + 7 = 15

func GetGeneralizedIndexForValidator(index uint64, validatorsArrayIndex uint64) uint64 {
	root := validatorsArrayIndex

	// Now, grab the validator index within the list
	// `start` is `index * 32` and `pos` is `start / 32` so pos is just `index`
	pos := index
	baseIndex := uint64(2) // Lists have a base index of 2
	root = root*baseIndex*math.GetPowerOfTwoCeil(beaconStateValidatorsMaxLength) + pos

	// root is now the generalized index for the validator
	return root
}

func (validator *Validator) ValidatorCredentialsPubkeyProof() ([][]byte, error) {
	// Just get the portion of the proof for the validator's credentials.
	generalizedIndex := beaconStateValidatorWithdrawalCredentialsPubkeyGeneralizedIndex
	root, err := validator.GetTree()
	if err != nil {
		return nil, fmt.Errorf("could not get validator tree: %w", err)
	}
	proof, err := root.Prove(int(generalizedIndex))
	if err != nil {
		return nil, fmt.Errorf("could not get proof for validator credentials: %w", err)
	}
	return proof.Hashes, nil
}

func (validator *Validator) ValidatorWithdrawableEpochProof() ([][]byte, error) {
	// Just get the portion of the proof for the validator ExitEpoch.
	generalizedIndex := BeaconStateValidatorWithdrawableEpochGeneralizedIndex
	root, err := validator.GetTree()
	if err != nil {
		return nil, fmt.Errorf("could not get validator tree: %w", err)
	}
	proof, err := root.Prove(int(generalizedIndex))
	if err != nil {
		return nil, fmt.Errorf("could not get proof for validator withdrawable epoch: %w", err)
	}
	return proof.Hashes, nil
}
