package validator

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prysmaticlabs/go-ssz"

	"github.com/rocket-pool/rocketpool-go/types"

	"github.com/rocket-pool/rocketpool-go/tests"
)

// Deposit settings
const depositAmount = 16000000000 // gwei

// Deposit data
type depositData struct {
	PublicKey             []byte `ssz-size:"48"`
	WithdrawalCredentials []byte `ssz-size:"32"`
	Amount                uint64
	Signature             []byte `ssz-size:"96"`
}

// Get the validator pubkey
func GetValidatorPubkey(pubkey int) (types.ValidatorPubkey, error) {
	if pubkey == 1 {
		return types.HexToValidatorPubkey(tests.ValidatorPubkey)
	} else if pubkey == 2 {
		return types.HexToValidatorPubkey(tests.ValidatorPubkey2)
	} else if pubkey == 3 {
		return types.HexToValidatorPubkey(tests.ValidatorPubkey3)
	} else {
		return types.ValidatorPubkey{}, fmt.Errorf("Invalid pubkey index %d", pubkey)
	}
}

// Get the validator deposit signature
func GetValidatorSignature(pubkey int) (types.ValidatorSignature, error) {
	if pubkey == 1 {
		return types.HexToValidatorSignature(tests.ValidatorSignature)
	} else if pubkey == 2 {
		return types.HexToValidatorSignature(tests.ValidatorSignature2)
	} else if pubkey == 3 {
		return types.HexToValidatorSignature(tests.ValidatorSignature3)
	} else {
		return types.ValidatorSignature{}, fmt.Errorf("Invalid pubkey index %d", pubkey)
	}
}

// Get the validator deposit depositDataRoot
func GetDepositDataRoot(validatorPubkey types.ValidatorPubkey, withdrawalCredentials common.Hash, validatorSignature types.ValidatorSignature) (common.Hash, error) {
	return ssz.HashTreeRoot(depositData{
		PublicKey:             validatorPubkey.Bytes(),
		WithdrawalCredentials: withdrawalCredentials[:],
		Amount:                depositAmount,
		Signature:             validatorSignature.Bytes(),
	})
}
