package validator

import (
    "github.com/prysmaticlabs/go-ssz"
    "github.com/rocket-pool/rocketpool-go/types"
    eth2types "github.com/wealdtech/go-eth2-types/v2"
)


// Voluntary exit message
type VoluntaryExit struct {
    Epoch uint64
    ValidatorIndex uint64
}


// Get a voluntary exit message signature for a given validator key and index
func GetSignedExitMessage(validatorKey *eth2types.BLSPrivateKey, validatorIndex uint64, epoch uint64, signatureDomain []byte) (types.ValidatorSignature, error) {

    // Build voluntary exit message
    exitMessage := VoluntaryExit{
        Epoch: epoch,
        ValidatorIndex: validatorIndex,
    }

    // Get object root
    or, err := ssz.HashTreeRoot(exitMessage)
    if err != nil {
        return types.ValidatorSignature{}, err
    }

    // Get signing root
    sr, err := ssz.HashTreeRoot(signingRoot{
        ObjectRoot: or[:],
        Domain: signatureDomain,
    })
    if err != nil {
        return types.ValidatorSignature{}, err
    }

    // Sign message
    signature := validatorKey.Sign(sr[:]).Marshal()

    // Return
    return types.BytesToValidatorSignature(signature), nil

}

