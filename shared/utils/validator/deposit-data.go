package validator

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/shared/types/eth2/generic"
	eth2types "github.com/wealdtech/go-eth2-types/v2"

	"github.com/rocket-pool/smartnode/shared/services/beacon"
)

// Get deposit data & root for a given validator key and withdrawal credentials
func GetDepositData(validatorKey *eth2types.BLSPrivateKey, withdrawalCredentials common.Hash, eth2Config beacon.Eth2Config, depositAmount uint64) (generic.DepositData, common.Hash, error) {

	// Build deposit data
	dd := generic.DepositDataNoSignature{
		PublicKey:             validatorKey.PublicKey().Marshal(),
		WithdrawalCredentials: withdrawalCredentials[:],
		Amount:                depositAmount,
	}

	// Get signing root
	or, err := dd.HashTreeRoot()
	if err != nil {
		return generic.DepositData{}, common.Hash{}, err
	}

	sr := generic.SigningRoot{
		ObjectRoot: or[:],
		Domain:     eth2types.Domain(eth2types.DomainDeposit, eth2Config.GenesisForkVersion, eth2types.ZeroGenesisValidatorsRoot),
	}

	// Get signing root with domain
	srHash, err := sr.HashTreeRoot()
	if err != nil {
		return generic.DepositData{}, common.Hash{}, err
	}

	// Build deposit data struct (with signature)
	var depositData = generic.DepositData{
		PublicKey:             dd.PublicKey,
		WithdrawalCredentials: dd.WithdrawalCredentials,
		Amount:                dd.Amount,
		Signature:             validatorKey.Sign(srHash[:]).Marshal(),
	}

	// Get deposit data root
	depositDataRoot, err := depositData.HashTreeRoot()
	if err != nil {
		return generic.DepositData{}, common.Hash{}, err
	}

	// Return
	return depositData, depositDataRoot, nil

}
