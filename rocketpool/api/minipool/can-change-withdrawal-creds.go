package minipool

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/rocketpool/common/beacon"
	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	sharedtypes "github.com/rocket-pool/smartnode/shared/types"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/validator"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
	util "github.com/wealdtech/go-eth2-util"
)

// ===============
// === Factory ===
// ===============

type minipoolCanChangeCredsContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolCanChangeCredsContextFactory) Create(vars map[string]string) (*minipoolCanChangeCredsContext, error) {
	c := &minipoolCanChangeCredsContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("address", vars, cliutils.ValidateAddress, &c.minipoolAddress),
		server.ValidateArg("mnemonic", vars, cliutils.ValidateWalletMnemonic, &c.mnemonic),
	}
	return c, errors.Join(inputErrs...)
}

func (f *minipoolCanChangeCredsContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*minipoolCanChangeCredsContext, api.MinipoolCanChangeWithdrawalCredentialsData](
		router, "change-withdrawal-creds/verify", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolCanChangeCredsContext struct {
	handler     *MinipoolHandler
	rp          *rocketpool.RocketPool
	bc          beacon.Client
	nodeAddress common.Address

	mnemonic        string
	minipoolAddress common.Address
	mpv3            *minipool.MinipoolV3
}

func (c *minipoolCanChangeCredsContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.bc = sp.GetBeaconClient()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	err := errors.Join(
		sp.RequireNodeRegistered(),
		sp.RequireBeaconClientSynced(),
	)
	if err != nil {
		return err
	}

	// Bindings
	mp, err := minipool.CreateMinipoolFromAddress(c.rp, c.minipoolAddress, false, nil)
	if err != nil {
		return fmt.Errorf("error creating minipool binding from address: %w", err)
	}
	mpv3, success := minipool.GetMinipoolAsV3(mp)
	if !success {
		return fmt.Errorf("minipool delegate is too old - it must be upgraded before you can change the withdrawal credentials to this minipool")
	}
	c.mpv3 = mpv3
	return nil
}

func (c *minipoolCanChangeCredsContext) GetState(mc *batch.MultiCaller) {
	c.mpv3.GetVacant(mc)
	c.mpv3.GetStatus(mc)
	c.mpv3.GetNodeAddress(mc)
	c.mpv3.GetPubkey(mc)
}

func (c *minipoolCanChangeCredsContext) PrepareData(data *api.MinipoolCanChangeWithdrawalCredentialsData, opts *bind.TransactOpts) error {
	// Validate minipool owner
	if c.mpv3.GetCommonDetails().NodeAddress != c.nodeAddress {
		return fmt.Errorf("minipool %s does not belong to the node", c.minipoolAddress.Hex())
	}

	// Check minipool status
	if !c.mpv3.Details.IsVacant {
		return fmt.Errorf("minipool %s is not vacant", c.minipoolAddress.Hex())
	}
	if c.mpv3.GetCommonDetails().Status.Formatted() != types.Prelaunch {
		return fmt.Errorf("minipool %s is not in prelaunch state", c.minipoolAddress.Hex())
	}

	// Check the validator's status and current creds
	pubkey := c.mpv3.GetCommonDetails().Pubkey
	beaconStatus, err := c.bc.GetValidatorStatus(pubkey, nil)
	if err != nil {
		return fmt.Errorf("error getting Beacon status for minipool %s (pubkey %s): %w", c.minipoolAddress.Hex(), pubkey.Hex(), err)
	}
	if beaconStatus.Status != sharedtypes.ValidatorState_ActiveOngoing {
		return fmt.Errorf("minipool %s (pubkey %s) was in state %v, but is required to be active_ongoing for migration", c.minipoolAddress.Hex(), pubkey.Hex(), beaconStatus.Status)
	}
	if beaconStatus.WithdrawalCredentials[0] != 0x00 {
		return fmt.Errorf("minipool %s (pubkey %s) has already been migrated - its withdrawal credentials are %s", c.minipoolAddress.Hex(), pubkey.Hex(), beaconStatus.WithdrawalCredentials.Hex())
	}

	// Get the index for this validator based on the mnemonic
	index := uint(0)
	validatorKeyPath := validator.ValidatorKeyPath
	var validatorKey *eth2types.BLSPrivateKey
	for index < validatorKeyRetrievalLimit {
		key, err := validator.GetPrivateKey(c.mnemonic, index, validatorKeyPath)
		if err != nil {
			return fmt.Errorf("error deriving key for index %d: %w", index, err)
		}
		candidatePubkey := key.PublicKey().Marshal()
		if bytes.Equal(pubkey[:], candidatePubkey) {
			validatorKey = key
			break
		}
		index++
	}
	if validatorKey == nil {
		return fmt.Errorf("couldn't find the validator key for this mnemonic after %d tries", validatorKeyRetrievalLimit)
	}

	// Get the withdrawal creds from this index
	withdrawalKey, err := validator.GetWithdrawalKey(c.mnemonic, index, validatorKeyPath)
	if err != nil {
		return fmt.Errorf("error getting withdrawal key for validator: %w", err)
	}
	withdrawalPubkey := withdrawalKey.PublicKey().Marshal()
	withdrawalPubkeyHashBytes := util.SHA256(withdrawalPubkey) // Withdrawal creds use sha256, *not* Keccak
	withdrawalPubkeyHash := common.BytesToHash(withdrawalPubkeyHashBytes)
	withdrawalPubkeyHash[0] = 0x00 // BLS prefix

	// Make sure they match what's on Beacon
	if beaconStatus.WithdrawalCredentials != withdrawalPubkeyHash {
		return fmt.Errorf("withdrawal credentials mismatch for minipool %s (pubkey %s): should be %s but matching index %d provided %s", c.minipoolAddress.Hex(), pubkey.Hex(), beaconStatus.WithdrawalCredentials.Hex(), index, withdrawalPubkeyHash.Hex())
	}

	// Update & return response
	data.CanChange = true
	return nil
}
