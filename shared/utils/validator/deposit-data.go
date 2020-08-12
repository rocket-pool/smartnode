package validator

import (
    "github.com/prysmaticlabs/go-ssz"
    eth2types "github.com/wealdtech/go-eth2-types/v2"

    "github.com/rocket-pool/smartnode/shared/services/beacon"
    bytesutil "github.com/rocket-pool/smartnode/shared/utils/bytes"
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
func GetDepositData(validatorKey *eth2types.BLSPrivateKey, withdrawalCredentials []byte, eth2Config beacon.Eth2Config) (DepositData, [32]byte, error) {

    // Build deposit data
    depositData := DepositData{
        PublicKey: validatorKey.PublicKey().Marshal(),
        WithdrawalCredentials: withdrawalCredentials,
        Amount: DepositAmount,
    }

    // Get deposit data signing root
    sr, err := ssz.SigningRoot(depositData)
    if err != nil {
        return DepositData{}, [32]byte{}, err
    }

    // Compute domain
    domain := eth2types.Domain(bytesutil.ToBytes4(eth2Config.DomainDeposit), eth2types.ZeroForkVersion, eth2types.ZeroGenesisValidatorsRoot)

    // Get deposit data signing root with domain
    srWithDomain, err := ssz.HashTreeRoot(signingRoot{
        ObjectRoot: sr[:],
        Domain: domain,
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

