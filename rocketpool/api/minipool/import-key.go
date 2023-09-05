package minipool

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/urfave/cli"
	eth2types "github.com/wealdtech/go-eth2-types/v2"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/validator"
)

func importKey(c *cli.Context, minipoolAddress common.Address, mnemonic string) (*api.ApiResponse, error) {
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
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ApiResponse{}

	// Create minipool
	mp, err := minipool.CreateMinipoolFromAddress(rp, minipoolAddress, false, nil)
	if err != nil {
		return nil, err
	}
	mpCommon := mp.GetMinipoolCommon()

	// Get the relevant details
	err = rp.Query(func(mc *batch.MultiCaller) error {
		mpCommon.GetNodeAddress(mc)
		mpCommon.GetPubkey(mc)
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool details: %w", err)
	}

	// Validate minipool owner
	if mpCommon.Details.NodeAddress != nodeAccount.Address {
		return nil, fmt.Errorf("minipool %s does not belong to the node", minipoolAddress.Hex())
	}

	// Get minipool validator pubkey
	pubkey := mpCommon.Details.Pubkey
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
