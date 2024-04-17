package minipool

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/eth"
	nmc_validator "github.com/rocket-pool/node-manager-core/node/validator"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/rocketpool-go/v2/minipool"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/v2/types"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/validator"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
	util "github.com/wealdtech/go-eth2-util"
)

// ===============
// === Factory ===
// ===============

type minipoolCanChangeCredsContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolCanChangeCredsContextFactory) Create(args url.Values) (*minipoolCanChangeCredsContext, error) {
	c := &minipoolCanChangeCredsContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("address", args, input.ValidateAddress, &c.minipoolAddress),
		server.ValidateArg("mnemonic", args, input.ValidateWalletMnemonic, &c.mnemonic),
	}
	return c, errors.Join(inputErrs...)
}

func (f *minipoolCanChangeCredsContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*minipoolCanChangeCredsContext, api.MinipoolCanChangeWithdrawalCredentialsData](
		router, "change-withdrawal-creds/verify", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolCanChangeCredsContext struct {
	handler     *MinipoolHandler
	rp          *rocketpool.RocketPool
	bc          beacon.IBeaconClient
	nodeAddress common.Address

	mnemonic        string
	minipoolAddress common.Address
	mpv3            *minipool.MinipoolV3
}

func (c *minipoolCanChangeCredsContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.bc = sp.GetBeaconClient()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireBeaconClientSynced(c.handler.ctx)
	if err != nil {
		return types.ResponseStatus_ClientsNotSynced, err
	}

	// Bindings
	mpMgr, err := minipool.NewMinipoolManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating minipool manager binding: %w", err)
	}
	mp, err := mpMgr.CreateMinipoolFromAddress(c.minipoolAddress, false, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating minipool binding from address: %w", err)
	}
	mpv3, success := minipool.GetMinipoolAsV3(mp)
	if !success {
		return types.ResponseStatus_InvalidChainState, fmt.Errorf("minipool delegate is too old - it must be upgraded before you can change the withdrawal credentials to this minipool")
	}
	c.mpv3 = mpv3
	return types.ResponseStatus_Success, nil
}

func (c *minipoolCanChangeCredsContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.mpv3.IsVacant,
		c.mpv3.Status,
		c.mpv3.NodeAddress,
		c.mpv3.Pubkey,
	)
}

func (c *minipoolCanChangeCredsContext) PrepareData(data *api.MinipoolCanChangeWithdrawalCredentialsData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	ctx := c.handler.ctx

	// Validate minipool owner
	if c.mpv3.Common().NodeAddress.Get() != c.nodeAddress {
		return types.ResponseStatus_InvalidChainState, fmt.Errorf("minipool %s does not belong to the node", c.minipoolAddress.Hex())
	}

	// Check minipool status
	if !c.mpv3.IsVacant.Get() {
		return types.ResponseStatus_InvalidChainState, fmt.Errorf("minipool %s is not vacant", c.minipoolAddress.Hex())
	}
	if c.mpv3.Common().Status.Formatted() != rptypes.MinipoolStatus_Prelaunch {
		return types.ResponseStatus_InvalidChainState, fmt.Errorf("minipool %s is not in prelaunch state", c.minipoolAddress.Hex())
	}

	// Check the validator's status and current creds
	pubkey := c.mpv3.Common().Pubkey.Get()
	beaconStatus, err := c.bc.GetValidatorStatus(ctx, pubkey, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting Beacon status for minipool %s (pubkey %s): %w", c.minipoolAddress.Hex(), pubkey.Hex(), err)
	}
	if beaconStatus.Status != beacon.ValidatorState_ActiveOngoing {
		return types.ResponseStatus_InvalidChainState, fmt.Errorf("minipool %s (pubkey %s) was in state %v, but is required to be active_ongoing for migration", c.minipoolAddress.Hex(), pubkey.Hex(), beaconStatus.Status)
	}
	if beaconStatus.WithdrawalCredentials[0] != 0x00 {
		return types.ResponseStatus_InvalidChainState, fmt.Errorf("minipool %s (pubkey %s) has already been migrated - its withdrawal credentials are %s", c.minipoolAddress.Hex(), pubkey.Hex(), beaconStatus.WithdrawalCredentials.Hex())
	}

	// Get the index for this validator based on the mnemonic
	index := uint(0)
	validatorKeyPath := fmt.Sprintf(validator.ValidatorKeyPath, index)
	var validatorKey *eth2types.BLSPrivateKey
	for index < validatorKeyRetrievalLimit {
		key, err := nmc_validator.GetPrivateKey(c.mnemonic, validatorKeyPath)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error deriving key for index %d: %w", index, err)
		}
		candidatePubkey := key.PublicKey().Marshal()
		if bytes.Equal(pubkey[:], candidatePubkey) {
			validatorKey = key
			break
		}
		index++
	}
	if validatorKey == nil {
		return types.ResponseStatus_ResourceNotFound, fmt.Errorf("couldn't find the validator key for this mnemonic after %d tries", validatorKeyRetrievalLimit)
	}

	// Get the withdrawal creds from this index
	withdrawalKey, err := nmc_validator.GetWithdrawalKey(c.mnemonic, validatorKeyPath)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting withdrawal key for validator: %w", err)
	}
	withdrawalPubkey := withdrawalKey.PublicKey().Marshal()
	withdrawalPubkeyHashBytes := util.SHA256(withdrawalPubkey) // Withdrawal creds use sha256, *not* Keccak
	withdrawalPubkeyHash := common.BytesToHash(withdrawalPubkeyHashBytes)
	withdrawalPubkeyHash[0] = 0x00 // BLS prefix

	// Make sure they match what's on Beacon
	if beaconStatus.WithdrawalCredentials != withdrawalPubkeyHash {
		return types.ResponseStatus_InvalidChainState, fmt.Errorf("withdrawal credentials mismatch for minipool %s (pubkey %s): should be %s but matching index %d provided %s", c.minipoolAddress.Hex(), pubkey.Hex(), beaconStatus.WithdrawalCredentials.Hex(), index, withdrawalPubkeyHash.Hex())
	}

	// Update & return response
	data.CanChange = true
	return types.ResponseStatus_Success, nil
}
