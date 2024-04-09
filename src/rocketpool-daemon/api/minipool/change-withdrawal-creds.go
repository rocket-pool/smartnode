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
	"github.com/rocket-pool/rocketpool-go/v2/minipool"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	eth2types "github.com/wealdtech/go-eth2-types/v2"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/beacon"
	nmc_validator "github.com/rocket-pool/node-manager-core/node/validator"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/validator"
)

// ===============
// === Factory ===
// ===============

type minipoolChangeCredsContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolChangeCredsContextFactory) Create(args url.Values) (*minipoolChangeCredsContext, error) {
	c := &minipoolChangeCredsContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("address", args, input.ValidateAddress, &c.minipoolAddress),
		server.ValidateArg("mnemonic", args, input.ValidateWalletMnemonic, &c.mnemonic),
	}
	return c, errors.Join(inputErrs...)
}

func (f *minipoolChangeCredsContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*minipoolChangeCredsContext, types.SuccessData](
		router, "change-withdrawal-creds", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolChangeCredsContext struct {
	handler     *MinipoolHandler
	rp          *rocketpool.RocketPool
	bc          beacon.IBeaconClient
	nodeAddress common.Address

	mnemonic        string
	minipoolAddress common.Address
	mp              minipool.IMinipool
}

func (c *minipoolChangeCredsContext) Initialize() (types.ResponseStatus, error) {
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
	c.mp, err = mpMgr.CreateMinipoolFromAddress(c.minipoolAddress, false, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating minipool binding from address: %w", err)
	}
	return types.ResponseStatus_Success, nil
}

func (c *minipoolChangeCredsContext) GetState(mc *batch.MultiCaller) {
	c.mp.Common().Pubkey.AddToQuery(mc)
}

func (c *minipoolChangeCredsContext) PrepareData(data *types.SuccessData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	ctx := c.handler.ctx
	// Get minipool validator pubkey
	pubkey := c.mp.Common().Pubkey.Get()

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

	// Get beacon head
	head, err := c.bc.GetBeaconHead(ctx)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting Beacon head: %w", err)
	}

	// Get the BlsToExecutionChange signature domain
	signatureDomain, err := c.bc.GetDomainData(ctx, eth2types.DomainBlsToExecutionChange[:], head.Epoch, true)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting signature domain: %w", err)
	}

	// Get validator index
	validatorIndex, err := c.bc.GetValidatorIndex(ctx, pubkey)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting validator index: %w", err)
	}

	// Get signed withdrawal creds change message
	signature, err := nmc_validator.GetSignedWithdrawalCredsChangeMessage(withdrawalKey, validatorIndex, c.minipoolAddress, signatureDomain)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting signed withdrawal credentials change message: %w", err)
	}

	// Broadcast withdrawal creds change message
	withdrawalPubkey := beacon.ValidatorPubkey(withdrawalKey.PublicKey().Marshal())
	if err := c.bc.ChangeWithdrawalCredentials(ctx, validatorIndex, withdrawalPubkey, c.minipoolAddress, signature); err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error submitting withdrawal credentials change message: %w", err)
	}
	return types.ResponseStatus_Success, nil
}
