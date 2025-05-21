package minipool

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/minipool"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/urfave/cli"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
	util "github.com/wealdtech/go-eth2-util"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/validator"
)

const (
	validatorLimit uint = 2000
)

func canChangeWithdrawalCreds(c *cli.Context, minipoolAddress common.Address, mnemonic string) (*api.CanChangeWithdrawalCredentialsResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}
	err = services.RequireEthClientSynced(c)
	if err != nil {
		return nil, err
	}
	err = services.RequireBeaconClientSynced(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.CanChangeWithdrawalCredentialsResponse{}

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Validate minipool owner
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}
	if err := validateMinipoolOwner(mp, nodeAccount.Address); err != nil {
		return nil, err
	}

	// Check minipool status
	mpv3, success := minipool.GetMinipoolAsV3(mp)
	if !success {
		return nil, fmt.Errorf("minipool delegate is too old - it must be upgraded before you can change the withdrawal credentials to this minipool")
	}
	details, err := mpv3.GetStatusDetails(nil)
	if err != nil {
		return nil, fmt.Errorf("error getting status details for minipool %s: %w", minipoolAddress.Hex(), err)
	}
	if !details.IsVacant {
		return nil, fmt.Errorf("minipool %s is not vacant", minipoolAddress.Hex())
	}
	if details.Status != types.Prelaunch {
		return nil, fmt.Errorf("minipool %s is not in prelaunch state", minipoolAddress.Hex())
	}

	// Check the validator's status and current creds
	pubkey, err := minipool.GetMinipoolPubkey(rp, minipoolAddress, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't get the pubkey for minipool %s: %w", minipoolAddress.Hex(), err)
	}
	beaconStatus, err := bc.GetValidatorStatus(pubkey, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting Beacon status for minipool %s (pubkey %s): %w", minipoolAddress.Hex(), pubkey.Hex(), err)
	}
	if beaconStatus.Status != beacon.ValidatorState_ActiveOngoing {
		return nil, fmt.Errorf("minipool %s (pubkey %s) was in state %v, but is required to be active_ongoing for migration", minipoolAddress.Hex(), pubkey.Hex(), beaconStatus.Status)
	}
	if beaconStatus.WithdrawalCredentials[0] != 0x00 {
		return nil, fmt.Errorf("minipool %s (pubkey %s) has already been migrated - its withdrawal credentials are %s", minipoolAddress.Hex(), pubkey.Hex(), beaconStatus.WithdrawalCredentials.Hex())
	}

	// Get the index for this validator based on the mnemonic
	index := uint(0)
	validatorKeyPath := validator.ValidatorKeyPath
	var validatorKey *eth2types.BLSPrivateKey
	for index < validatorLimit {
		key, err := validator.GetPrivateKey(mnemonic, index, validatorKeyPath)
		if err != nil {
			return nil, fmt.Errorf("error deriving key for index %d: %w", index, err)
		}
		candidatePubkey := key.PublicKey().Marshal()
		if bytes.Equal(pubkey[:], candidatePubkey) {
			validatorKey = key
			break
		}
		index++
	}
	if validatorKey == nil {
		return nil, fmt.Errorf("couldn't find the validator key for this mnemonic after %d tries", validatorLimit)
	}

	// Get the withdrawal creds from this index
	withdrawalKey, err := validator.GetWithdrawalKey(mnemonic, index, validatorKeyPath)
	if err != nil {
		return nil, err
	}
	withdrawalPubkey := withdrawalKey.PublicKey().Marshal()
	withdrawalPubkeyHashBytes := util.SHA256(withdrawalPubkey) // Withdrawal creds use sha256, *not* Keccak
	withdrawalPubkeyHash := common.BytesToHash(withdrawalPubkeyHashBytes)
	withdrawalPubkeyHash[0] = 0x00 // BLS prefix

	// Make sure they match what's on Beacon
	if beaconStatus.WithdrawalCredentials != withdrawalPubkeyHash {
		return nil, fmt.Errorf("withdrawal credentials mismatch for minipool %s (pubkey %s): should be %s but matching index %d provided %s", minipoolAddress.Hex(), pubkey.Hex(), beaconStatus.WithdrawalCredentials.Hex(), index, withdrawalPubkeyHash.Hex())
	}

	// Update & return response
	response.CanChange = true
	return &response, nil

}

func changeWithdrawalCreds(c *cli.Context, minipoolAddress common.Address, mnemonic string) (*api.ChangeWithdrawalCredentialsResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	if err := services.RequireBeaconClientSynced(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ChangeWithdrawalCredentialsResponse{}

	// Get minipool validator pubkey
	pubkey, err := minipool.GetMinipoolPubkey(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Get the index for this validator based on the mnemonic
	index := uint(0)
	validatorKeyPath := validator.ValidatorKeyPath
	var validatorKey *eth2types.BLSPrivateKey
	for index < validatorLimit {
		key, err := validator.GetPrivateKey(mnemonic, index, validatorKeyPath)
		if err != nil {
			return nil, fmt.Errorf("error deriving key for index %d: %w", index, err)
		}
		candidatePubkey := key.PublicKey().Marshal()
		if bytes.Equal(pubkey[:], candidatePubkey) {
			validatorKey = key
			break
		}
		index++
	}
	if validatorKey == nil {
		return nil, fmt.Errorf("couldn't find the validator key for this mnemonic after %d tries", validatorLimit)
	}

	// Get the withdrawal creds from this index
	withdrawalKey, err := validator.GetWithdrawalKey(mnemonic, index, validatorKeyPath)
	if err != nil {
		return nil, err
	}

	// Get beacon head
	head, err := bc.GetBeaconHead()
	if err != nil {
		return nil, err
	}

	// Get the BlsToExecutionChange signature domain
	signatureDomain, err := bc.GetDomainData(eth2types.DomainBlsToExecutionChange[:], head.Epoch, true)
	if err != nil {
		return nil, err
	}

	// Get validator index
	validatorIndex, err := bc.GetValidatorIndex(pubkey)
	if err != nil {
		return nil, err
	}

	// Get signed withdrawal creds change message
	signature, err := validator.GetSignedWithdrawalCredsChangeMessage(withdrawalKey, validatorIndex, minipoolAddress, signatureDomain)
	if err != nil {
		return nil, err
	}

	// Broadcast withdrawal creds change message
	withdrawalPubkey := types.BytesToValidatorPubkey(withdrawalKey.PublicKey().Marshal())
	if err := bc.ChangeWithdrawalCredentials(validatorIndex, withdrawalPubkey, minipoolAddress, signature); err != nil {
		return nil, err
	}

	// Return response
	return &response, nil
}
