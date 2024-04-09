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

type walletRecoverContextFactory struct {
	handler *WalletHandler
}

func (f *walletRecoverContextFactory) Create(args url.Values) (*walletRecoverContext, error) {
	c := &walletRecoverContext{
		handler: f.handler,
	}
	server.GetOptionalStringFromVars("derivation-path", args, &c.derivationPath)
	inputErrs := []error{
		server.ValidateArg("mnemonic", args, input.ValidateWalletMnemonic, &c.mnemonic),
		server.ValidateOptionalArg("skip-validator-key-recovery", args, input.ValidateBool, &c.skipValidatorKeyRecovery, nil),
		server.ValidateOptionalArg("index", args, input.ValidateUint, &c.index, nil),
		server.ValidateArg("password", args, input.ValidateNodePassword, &c.password),
		server.ValidateArg("save-password", args, input.ValidateBool, &c.savePassword),
	}
	return c, errors.Join(inputErrs...)
}

func (f *walletRecoverContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*walletRecoverContext, api.WalletRecoverData](
		router, "recover", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type walletRecoverContext struct {
	handler                  *WalletHandler
	skipValidatorKeyRecovery bool
	mnemonic                 string
	derivationPath           string
	index                    uint64
	password                 string
	savePassword             bool
}

func (c *walletRecoverContext) PrepareData(data *api.WalletRecoverData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	w := sp.GetWallet()
	vMgr := sp.GetValidatorManager()

	// Requirements
	status, err := w.GetStatus()
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting wallet status: %w", err)
	}
	if status.Wallet.IsOnDisk {
		return types.ResponseStatus_ResourceConflict, fmt.Errorf("a wallet is already present")
	}
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
	err = w.Recover(path, uint(c.index), c.mnemonic, c.password, c.savePassword, false)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error recovering wallet: %w", err)
	}
	data.AccountAddress, _ = w.GetAddress()

	// Recover validator keys
	if !c.skipValidatorKeyRecovery {
		data.ValidatorKeys, err = vMgr.RecoverMinipoolKeys(false)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error recovering minipool keys: %w", err)
		}
	}

	return types.ResponseStatus_Success, nil
}
