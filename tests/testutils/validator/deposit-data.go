package validator

import (
    "github.com/ethereum/go-ethereum/common"
    "github.com/prysmaticlabs/go-ssz"

    "github.com/rocket-pool/rocketpool-go/types"

    "github.com/rocket-pool/rocketpool-go/tests"
)


// Deposit settings
const depositAmount = 32000000000 // gwei


// Deposit data
type depositData struct {
    PublicKey []byte                `ssz-size:"48"`
    WithdrawalCredentials []byte    `ssz-size:"32"`
    Amount uint64
    Signature []byte                `ssz-size:"96"`
}


// Get the validator pubkey
func GetValidatorPubkey() (types.ValidatorPubkey, error) {
    return types.HexToValidatorPubkey(tests.ValidatorPubkey)
}


// Get the validator deposit signature
func GetValidatorSignature() (types.ValidatorSignature, error) {
    return types.HexToValidatorSignature(tests.ValidatorSignature)
}


// Get the validator deposit depositDataRoot
func GetDepositDataRoot(validatorPubkey types.ValidatorPubkey, withdrawalCredentials common.Hash, validatorSignature types.ValidatorSignature) (common.Hash, error) {
    return ssz.HashTreeRoot(depositData{
        PublicKey: validatorPubkey.Bytes(),
        WithdrawalCredentials: withdrawalCredentials[:],
        Amount: depositAmount,
        Signature: validatorSignature.Bytes(),
    })
}

