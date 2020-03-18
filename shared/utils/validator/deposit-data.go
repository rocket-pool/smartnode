package validator

import (
    "github.com/prysmaticlabs/go-ssz"

    "github.com/rocket-pool/smartnode/shared/utils/bls"
    bytesutil "github.com/rocket-pool/smartnode/shared/utils/bytes"
)


// Deposit settings
const DEPOSIT_AMOUNT uint64 = 32000000000

// BLS settings
const BLS_DOMAIN_DEPOSIT uint64 = 3
var GenesisForkVersion []byte = []byte{1,3,3,7}


// Deposit data
type DepositData struct {
    PublicKey [48]byte
    WithdrawalCredentials [32]byte
    Amount uint64
    Signature [96]byte
}


// Deposit data BLS signing root
type SigningRoot struct {
    ObjectRoot [32]byte
    Domain [8]byte
}


// Get deposit data & root for a given validator key and withdrawal credentials
func GetDepositData(validatorKey *bls.Key, withdrawalCredentials []byte) (*DepositData, [32]byte, error) {

    // Compute domain
    domain := bls.ComputeDomain(bytesutil.ToBytes4(bytesutil.Bytes4(BLS_DOMAIN_DEPOSIT)), GenesisForkVersion)

    // Build deposit data
    depositData := &DepositData{}
    copy(depositData.PublicKey[:], validatorKey.PublicKey.Marshal())
    copy(depositData.WithdrawalCredentials[:], withdrawalCredentials)
    depositData.Amount = DEPOSIT_AMOUNT

    // Get deposit data signing root
    signingRoot, err := ssz.SigningRoot(depositData)
    if err != nil { return nil, [32]byte{}, err }

    // Get deposit data signing root with domain
    signingRootObject := &SigningRoot{}
    copy(signingRootObject.ObjectRoot[:], signingRoot[:])
    copy(signingRootObject.Domain[:], domain)
    signingRootWithDomain, err := ssz.HashTreeRoot(signingRootObject)
    if err != nil { return nil, [32]byte{}, err }

    // Sign deposit data
    signature := validatorKey.SecretKey.Sign(signingRootWithDomain[:]).Marshal()
    copy(depositData.Signature[:], signature)

    // Get deposit data root
    depositDataRoot, err := ssz.HashTreeRoot(depositData)
    if err != nil { return nil, [32]byte{}, err }

    // Return
    return depositData, depositDataRoot, nil

}

