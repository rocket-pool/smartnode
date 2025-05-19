package validator

import (
	"fmt"
	"strconv"

	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared/types/eth2/generic"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
)

// Get a voluntary exit message signature for a given validator key and index
func GetSignedExitMessage(validatorKey *eth2types.BLSPrivateKey, validatorIndex string, epoch uint64, signatureDomain []byte) (types.ValidatorSignature, error) {

	// Parse the validator index
	indexNum, err := strconv.ParseUint(validatorIndex, 10, 64)
	if err != nil {
		return types.ValidatorSignature{}, fmt.Errorf("error parsing validator index (%s): %w", validatorIndex, err)
	}

	// Build voluntary exit message
	exitMessage := generic.VoluntaryExit{
		Epoch:          epoch,
		ValidatorIndex: indexNum,
	}

	// Get object root
	or, err := exitMessage.HashTreeRoot()
	if err != nil {
		return types.ValidatorSignature{}, err
	}

	// Get signing root
	sr := generic.SigningRoot{
		ObjectRoot: or[:],
		Domain:     signatureDomain,
	}

	srHash, err := sr.HashTreeRoot()
	if err != nil {
		return types.ValidatorSignature{}, err
	}

	// Sign message
	signature := validatorKey.Sign(srHash[:]).Marshal()

	// Return
	return types.BytesToValidatorSignature(signature), nil

}
