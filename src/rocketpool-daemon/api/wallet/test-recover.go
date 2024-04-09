package wallet

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	nodewallet "github.com/rocket-pool/node-manager-core/node/wallet"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/node-manager-core/wallet"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type walletTestRecoverContextFactory struct {
	handler *WalletHandler
}

func (f *walletTestRecoverContextFactory) Create(args url.Values) (*walletTestRecoverContext, error) {
	c := &walletTestRecoverContext{
		handler: f.handler,
	}
	server.GetOptionalStringFromVars("derivation-path", args, &c.derivationPath)
	inputErrs := []error{
		server.ValidateArg("mnemonic", args, input.ValidateWalletMnemonic, &c.mnemonic),
		server.ValidateOptionalArg("skip-validator-key-recovery", args, input.ValidateBool, &c.skipValidatorKeyRecovery, nil),
		server.ValidateOptionalArg("index", args, input.ValidateUint, &c.index, nil),
	}
	return c, errors.Join(inputErrs...)
}

func (f *walletTestRecoverContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*walletTestRecoverContext, api.WalletRecoverData](
		router, "test-recover", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type walletTestRecoverContext struct {
	handler                  *WalletHandler
	skipValidatorKeyRecovery bool
	mnemonic                 string
	derivationPath           string
	index                    uint64
}

func (c *walletTestRecoverContext) PrepareData(data *api.WalletRecoverData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	rs := sp.GetNetworkResources()
	vMgr := sp.GetValidatorManager()

	if !c.skipValidatorKeyRecovery {
		status, err := sp.RequireRocketPoolContracts(c.handler.ctx)
		if err != nil {
			return status, err
		}
	}

	// Parse the derivation path
	path, err := nodewallet.GetDerivationPath(wallet.DerivationPath(c.derivationPath))
	if err != nil {
		return types.ResponseStatus_InvalidArguments, err
	}

	// Recover the wallet
	w, err := nodewallet.TestRecovery(path, uint(c.index), c.mnemonic, rs.ChainID)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error recovering wallet: %w", err)
	}
	data.AccountAddress, _ = w.GetAddress()

	// Recover validator keys
	if !c.skipValidatorKeyRecovery {
		data.ValidatorKeys, err = vMgr.RecoverMinipoolKeys(true)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error recovering minipool keys: %w", err)
		}
	}
	return types.ResponseStatus_Success, nil
}
