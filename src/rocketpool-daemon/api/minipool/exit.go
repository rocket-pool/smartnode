package minipool

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"

	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/node/validator/utils"
	"github.com/rocket-pool/node-manager-core/node/wallet"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/server"
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

func (f *minipoolExitContextFactory) GetCancelContext() context.Context {
	return f.handler.context
}

func (f *minipoolExitContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterMinipoolRoute[*minipoolExitContext, types.SuccessData](
		router, "exit", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolExitContext struct {
	handler *MinipoolHandler
	rp      *rocketpool.RocketPool
	w       *wallet.Wallet
	bc      beacon.IBeaconClient

	minipoolAddresses []common.Address
}

func (c *minipoolExitContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.w = sp.GetWallet()
	c.bc = sp.GetBeaconClient()

	// Requirements
	err := errors.Join(
		sp.RequireNodeRegistered(c.handler.context),
		sp.RequireBeaconClientSynced(c.handler.context),
		sp.RequireWalletReady(),
	)
	if err != nil {
		return err
	}
	return nil
}

func (c *minipoolExitContext) GetState(node *node.Node, mc *batch.MultiCaller) {
}

func (c *minipoolExitContext) CheckState(node *node.Node, response *types.SuccessData) bool {
	return true
}

func (c *minipoolExitContext) GetMinipoolDetails(mc *batch.MultiCaller, mp minipool.IMinipool, index int) {
	mp.Common().Pubkey.AddToQuery(mc)
}

func (c *minipoolExitContext) PrepareData(addresses []common.Address, mps []minipool.IMinipool, data *types.SuccessData) error {
	// Get beacon head
	head, err := c.bc.GetBeaconHead(c.handler.context)
	if err != nil {
		return fmt.Errorf("error getting beacon head: %w", err)
	}

	// Get voluntary exit signature domain
	signatureDomain, err := c.bc.GetDomainData(c.handler.context, eth2types.DomainVoluntaryExit[:], head.Epoch, false)
	if err != nil {
		return fmt.Errorf("error getting beacon domain data: %w", err)
	}

	for _, mp := range mps {
		mpCommon := mp.Common()
		minipoolAddress := mpCommon.Address
		validatorPubkey := mpCommon.Pubkey.Get()

		// Get validator private key
		validatorKey, err := c.w.GetValidatorKeyByPubkey(validatorPubkey)
		if err != nil {
			return fmt.Errorf("error getting private key for minipool %s (pubkey %s): %w", minipoolAddress.Hex(), validatorPubkey.Hex(), err)
		}

		// Get validator index
		validatorIndex, err := c.bc.GetValidatorIndex(c.handler.context, validatorPubkey)
		if err != nil {
			return fmt.Errorf("error getting index of minipool %s (pubkey %s): %w", minipoolAddress.Hex(), validatorPubkey.Hex(), err)
		}

		// Get signed voluntary exit message
		signature, err := utils.GetSignedExitMessage(validatorKey, validatorIndex, head.Epoch, signatureDomain)
		if err != nil {
			return fmt.Errorf("error getting exit message signature for minipool %s (pubkey %s): %w", minipoolAddress.Hex(), validatorPubkey.Hex(), err)
		}

		// Broadcast voluntary exit message
		if err := c.bc.ExitValidator(c.handler.context, validatorIndex, head.Epoch, signature); err != nil {
			return fmt.Errorf("error submitting exit message for minipool %s (pubkey %s): %w", minipoolAddress.Hex(), validatorPubkey.Hex(), err)
		}
	}
	return nil
}
