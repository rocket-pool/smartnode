package minipool

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/minipool"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/urfave/cli"
	eth2types "github.com/wealdtech/go-eth2-types/v2"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/validator"
)

func importKey(c *cli.Context, minipoolAddress common.Address, mnemonic string) (*api.ImportKeyResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	if err := services.RequireBeaconClientSynced(c); err != nil {
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

	// Response
	response := api.ImportKeyResponse{}

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

	// Get minipool validator pubkey
	pubkey, err := minipool.GetMinipoolPubkey(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}
	emptyPubkey := types.ValidatorPubkey{}
	if pubkey == emptyPubkey {
		return nil, fmt.Errorf("minipool %s does not have a validator pubkey associated with it", minipoolAddress.Hex())
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

	// Save the keystore to disk
	derivationPath := fmt.Sprintf(validatorKeyPath, index)
	err = w.StoreValidatorKey(validatorKey, derivationPath)
	if err != nil {
		return nil, fmt.Errorf("error saving keystore: %w", err)
	}

	// Return response
	return &response, nil
}
