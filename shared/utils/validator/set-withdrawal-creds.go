package validator

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared/types/eth2"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
)

// Get a voluntary exit message signature for a given validator key and index
func GetSignedWithdrawalCredsChangeMessage(validatorKey *eth2types.BLSPrivateKey, validatorIndex uint64, fromBlsPubkey types.ValidatorPubkey, newWithdrawalAddress common.Address, signatureDomain []byte) (types.ValidatorSignature, error) {

	// Build withdrawal creds change message
	message := eth2.WithdrawalCredentialsChange{
		ValidatorIndex:     fmt.Sprintf("%d", validatorIndex),
		FromBLSPubkey:      fromBlsPubkey.Hex(),
		ToExecutionAddress: newWithdrawalAddress.Hex(),
	}

	// Get object root
	or, err := message.HashTreeRoot()
	if err != nil {
		return types.ValidatorSignature{}, err
	}

	// Get signing root
	sr := eth2.SigningRoot{
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
