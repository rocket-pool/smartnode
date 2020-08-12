package validator

import (
    "github.com/prysmaticlabs/go-ssz"
    eth2types "github.com/wealdtech/go-eth2-types/v2"
)


// Deposit settings
const DepositAmount = 32000000000 // gwei


// Deposit data
type DepositData struct {
    PublicKey []byte                `ssz-size:"48"`
    WithdrawalCredentials []byte    `ssz-size:"32"`
    Amount uint64
    Signature []byte                `ssz-size:"96"`
}


// BLS signing root with domain
type signingRoot struct {
    ObjectRoot []byte               `ssz-size:"32"`
    Domain []byte                   `ssz-size:"32"`
}


// Get deposit data & root for a given validator key and withdrawal credentials
func GetDepositData(validatorKey *eth2types.BLSPrivateKey, withdrawalCredentials []byte) (DepositData, [32]byte, error) {

    // Build deposit data
    depositData := DepositData{
        PublicKey: validatorKey.PublicKey().Marshal(),
        WithdrawalCredentials: withdrawalCredentials,
        Amount: DepositAmount,
    }

    // Get signing root
    sr, err := ssz.SigningRoot(depositData)
    if err != nil {
        return DepositData{}, [32]byte{}, err
    }

    // Get signing root with domain
    srWithDomain, err := ssz.HashTreeRoot(signingRoot{
        ObjectRoot: sr[:],
        Domain: eth2types.Domain(eth2types.DomainDeposit, eth2types.ZeroForkVersion, eth2types.ZeroGenesisValidatorsRoot),
    })
    if err != nil {
        return DepositData{}, [32]byte{}, err
    }

    // Sign deposit data
    depositData.Signature = validatorKey.Sign(srWithDomain[:]).Marshal()

    // Get deposit data root
    depositDataRoot, err := ssz.HashTreeRoot(depositData)
    if err != nil {
        return DepositData{}, [32]byte{}, err
    }

    // Return
    return depositData, depositDataRoot, nil

}

