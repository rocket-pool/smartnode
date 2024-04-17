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

type walletInitializeContextFactory struct {
	handler *WalletHandler
}

func (f *walletInitializeContextFactory) Create(args url.Values) (*walletInitializeContext, error) {
	c := &walletInitializeContext{
		handler: f.handler,
	}
	server.GetOptionalStringFromVars("derivation-path", args, &c.derivationPath)
	inputErrs := []error{
		server.ValidateOptionalArg("index", args, input.ValidateUint, &c.index, nil),
		server.ValidateArg("password", args, input.ValidateNodePassword, &c.password),
		server.ValidateArg("save-wallet", args, input.ValidateBool, &c.saveWallet),
		server.ValidateArg("save-password", args, input.ValidateBool, &c.savePassword),
	}
	return c, errors.Join(inputErrs...)
}

func (f *walletInitializeContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*walletInitializeContext, api.WalletInitializeData](
		router, "initialize", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type walletInitializeContext struct {
	handler        *WalletHandler
	derivationPath string
	index          uint64
	password       string
	savePassword   bool
	saveWallet     bool
}

func (c *walletInitializeContext) PrepareData(data *api.WalletInitializeData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider

	// Parse the derivation path
	path, err := nodewallet.GetDerivationPath(wallet.DerivationPath(c.derivationPath))
	if err != nil {
		return types.ResponseStatus_InvalidArguments, err
	}

	var w *nodewallet.Wallet
	var mnemonic string
	if !c.saveWallet {
		// Make a dummy wallet for the sake of creating a mnemonic and derived address
		mnemonic, err = nodewallet.GenerateNewMnemonic()
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error generating new mnemonic: %w", err)
		}

		w, err = nodewallet.TestRecovery(path, uint(c.index), mnemonic, 0)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error generating wallet from new mnemonic: %w", err)
		}
	} else {
		// Initialize the daemon wallet
		w = sp.GetWallet()

		// Requirements
		status, err := w.GetStatus()
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting wallet status: %w", err)
		}
		if status.Wallet.IsOnDisk {
			return types.ResponseStatus_ResourceConflict, fmt.Errorf("a wallet is already present")
		}

		// Create the new wallet
		mnemonic, err = w.CreateNewLocalWallet(path, uint(c.index), c.password, c.savePassword)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error initializing new wallet: %w", err)
		}
	}

	data.Mnemonic = mnemonic
	data.AccountAddress, _ = w.GetAddress()
	return types.ResponseStatus_Success, nil
}
