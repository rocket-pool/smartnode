package minipool

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/v2/minipool"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/beacon"
	nmc_validator "github.com/rocket-pool/node-manager-core/node/validator"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/validator"
	"github.com/rocket-pool/smartnode/v2/shared/keys"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
)

// ===============
// === Factory ===
// ===============

type minipoolExitContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolExitContextFactory) Create(args url.Values) (*minipoolExitContext, error) {
	c := &minipoolExitContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArgBatch("addresses", args, minipoolAddressBatchSize, input.ValidateAddress, &c.minipoolAddresses),
	}
	return c, errors.Join(inputErrs...)
}

func (f *minipoolExitContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*minipoolExitContext, types.SuccessData](
		router, "exit", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolExitContext struct {
	handler *MinipoolHandler
	rp      *rocketpool.RocketPool
	vMgr    *validator.ValidatorManager
	bc      beacon.IBeaconClient
	mps     []minipool.IMinipool

	minipoolAddresses []common.Address
}

func (c *minipoolExitContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.vMgr = sp.GetValidatorManager()
	c.bc = sp.GetBeaconClient()

	// Requirements
	err := sp.RequireBeaconClientSynced(c.handler.ctx)
	if err != nil {
		return types.ResponseStatus_ClientsNotSynced, err
	}
	err = sp.RequireWalletReady()
	if err != nil {
		return types.ResponseStatus_WalletNotReady, err
	}
	mpMgr, err := minipool.NewMinipoolManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating minipool manager binding: %w", err)
	}
	c.mps, err = mpMgr.CreateMinipoolsFromAddresses(c.minipoolAddresses, false, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating minipool bindings: %w", err)
	}
	return types.ResponseStatus_Success, nil
}

func (c *minipoolExitContext) GetState(mc *batch.MultiCaller) {
	for _, mp := range c.mps {
		mp.Common().Pubkey.AddToQuery(mc)
	}
}

func (c *minipoolExitContext) PrepareData(data *types.SuccessData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	ctx := c.handler.ctx
	// Get beacon head
	head, err := c.bc.GetBeaconHead(ctx)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting beacon head: %w", err)
	}

	// Get voluntary exit signature domain
	signatureDomain, err := c.bc.GetDomainData(ctx, eth2types.DomainVoluntaryExit[:], head.Epoch, false)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting beacon domain data: %w", err)
	}

	for _, mp := range c.mps {
		mpCommon := mp.Common()
		minipoolAddress := mpCommon.Address
		validatorPubkey := mpCommon.Pubkey.Get()

		// Get validator private key
		validatorKey, err := c.vMgr.LoadValidatorKey(validatorPubkey)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting private key for minipool %s (pubkey %s): %w", minipoolAddress.Hex(), validatorPubkey.Hex(), err)
		}

		// Get validator index
		validatorIndex, err := c.bc.GetValidatorIndex(ctx, validatorPubkey)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting index of minipool %s (pubkey %s): %w", minipoolAddress.Hex(), validatorPubkey.Hex(), err)
		}

		// Get signed voluntary exit message
		signature, err := nmc_validator.GetSignedExitMessage(validatorKey, validatorIndex, head.Epoch, signatureDomain)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting exit message signature for minipool %s (pubkey %s): %w", minipoolAddress.Hex(), validatorPubkey.Hex(), err)
		}

		// Broadcast voluntary exit message
		if err := c.bc.ExitValidator(ctx, validatorIndex, head.Epoch, signature); err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error submitting exit message for minipool %s (pubkey %s): %w", minipoolAddress.Hex(), validatorPubkey.Hex(), err)
		}
		c.handler.logger.Info("Minipool exit submitted",
			slog.String(keys.MinipoolKey, minipoolAddress.Hex()),
			slog.String(keys.PubkeyKey, validatorPubkey.Hex()),
		)
	}
	return types.ResponseStatus_Success, nil
}
