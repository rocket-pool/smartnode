package wallet

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/rocketpool/common/wallet"
	"github.com/rocket-pool/smartnode/shared/types"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
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
		server.ValidateOptionalArg("password", args, input.ValidateNodePassword, &c.password, &c.passwordExists),
		server.ValidateOptionalArg("save-password", args, input.ValidateBool, &c.savePassword, nil),
	}
	return c, errors.Join(inputErrs...)
}

func (f *walletInitializeContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*walletInitializeContext, api.WalletInitializeData](
		router, "initialize", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type walletInitializeContext struct {
	handler        *WalletHandler
	derivationPath string
	index          uint64
	password       []byte
	passwordExists bool
	savePassword   bool
}

func (c *walletInitializeContext) PrepareData(data *api.WalletInitializeData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	w := sp.GetWallet()

	// Requirements
	status := w.GetStatus()
	if status.HasKeystore {
		return fmt.Errorf("a wallet is already present")
	}

	_, hasPassword := w.GetPassword()
	if !hasPassword && !c.passwordExists {
		return fmt.Errorf("you must set a password before initializing the wallet, or provide one in this call")
	}
	w.RememberPassword(c.password)
	if c.savePassword {
		err := w.SavePassword()
		if err != nil {
			return fmt.Errorf("error saving wallet password to disk: %w", err)
		}
	}

	// Parse the derivation path
	pathType := types.DerivationPath(c.derivationPath)
	var path string
	switch pathType {
	case types.DerivationPath_Default:
		path = wallet.DefaultNodeKeyPath
	case types.DerivationPath_LedgerLive:
		path = wallet.LedgerLiveNodeKeyPath
	case types.DerivationPath_Mew:
		path = wallet.MyEtherWalletNodeKeyPath
	default:
		return fmt.Errorf("[%s] is not a valid derivation path type", c.derivationPath)
	}

	// Create the new wallet
	mnemonic, err := w.CreateNewWallet(path, uint(c.index))
	if err != nil {
		return fmt.Errorf("error initializing new wallet: %w", err)
	}
	data.Mnemonic = mnemonic
	data.AccountAddress, _ = w.GetAddress()
	return nil
}
