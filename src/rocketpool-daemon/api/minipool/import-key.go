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
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/minipool"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	eth2types "github.com/wealdtech/go-eth2-types/v2"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	nmc_validator "github.com/rocket-pool/node-manager-core/node/validator"
	"github.com/rocket-pool/node-manager-core/node/wallet"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/validator"
)

// ===============
// === Factory ===
// ===============

type minipoolImportKeyContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolImportKeyContextFactory) Create(args url.Values) (*minipoolImportKeyContext, error) {
	c := &minipoolImportKeyContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("address", args, input.ValidateAddress, &c.minipoolAddress),
		server.ValidateArg("mnemonic", args, input.ValidateWalletMnemonic, &c.mnemonic),
	}
	return c, errors.Join(inputErrs...)
}

func (f *minipoolImportKeyContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*minipoolImportKeyContext, types.SuccessData](
		router, "import-key", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolImportKeyContext struct {
	handler     *MinipoolHandler
	rp          *rocketpool.RocketPool
	w           *wallet.Wallet
	vMgr        *validator.ValidatorManager
	nodeAddress common.Address
	mp          minipool.IMinipool

	minipoolAddress common.Address
	mnemonic        string
}

func (c *minipoolImportKeyContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.w = sp.GetWallet()
	c.nodeAddress, _ = c.w.GetAddress()

	// Requirements
	err := sp.RequireWalletReady()
	if err != nil {
		return types.ResponseStatus_WalletNotReady, err
	}

	// Bindings
	mpMgr, err := minipool.NewMinipoolManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating minipool manager binding: %w", err)
	}
	c.mp, err = mpMgr.CreateMinipoolFromAddress(c.minipoolAddress, false, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating minipool binding: %w", err)
	}
	return types.ResponseStatus_Success, nil
}

func (c *minipoolImportKeyContext) GetState(mc *batch.MultiCaller) {
	mpCommon := c.mp.Common()
	eth.AddQueryablesToMulticall(mc,
		mpCommon.NodeAddress,
		mpCommon.Pubkey,
	)
}

func (c *minipoolImportKeyContext) PrepareData(data *types.SuccessData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	// Validate minipool owner
	mpCommon := c.mp.Common()
	if mpCommon.NodeAddress.Get() != c.nodeAddress {
		return types.ResponseStatus_InvalidChainState, fmt.Errorf("minipool %s does not belong to the node", c.minipoolAddress.Hex())
	}

	// Get minipool validator pubkey
	pubkey := mpCommon.Pubkey.Get()
	emptyPubkey := beacon.ValidatorPubkey{}
	if pubkey == emptyPubkey {
		return types.ResponseStatus_InvalidChainState, fmt.Errorf("minipool %s does not have a validator pubkey associated with it", c.minipoolAddress.Hex())
	}

	// Get the index for this validator based on the mnemonic
	index := uint(0)
	validatorKeyPath := validator.ValidatorKeyPath
	var validatorKey *eth2types.BLSPrivateKey
	for index < validatorKeyRetrievalLimit {
		validatorKeyPath := fmt.Sprintf(validatorKeyPath, index)
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

	// Save the keystore to disk
	derivationPath := fmt.Sprintf(validatorKeyPath, index)
	err := c.vMgr.StoreValidatorKey(validatorKey, derivationPath)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error saving keystore: %w", err)
	}
	return types.ResponseStatus_Success, nil
}
