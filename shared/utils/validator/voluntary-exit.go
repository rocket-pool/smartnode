package validator

import (
    "github.com/rocket-pool/rocketpool-go/types"
    "github.com/rocket-pool/smartnode/shared/types/eth2"
    eth2types "github.com/wealdtech/go-eth2-types/v2"
)


// Get a voluntary exit message signature for a given validator key and index
func GetSignedExitMessage(validatorKey *eth2types.BLSPrivateKey, validatorIndex uint64, epoch uint64, signatureDomain []byte) (types.ValidatorSignature, error) {

    // Build voluntary exit message
    exitMessage := eth2.VoluntaryExit{
        Epoch: epoch,
        ValidatorIndex: validatorIndex,
    }

    // Get object root
    or, err := exitMessage.HashTreeRoot()
    if err != nil {
        return types.ValidatorSignature{}, err
    }

    // Get signing root
    sr := eth2.SigningRoot{
        ObjectRoot: or[:],
        Domain: signatureDomain,
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

