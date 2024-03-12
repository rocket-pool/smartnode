package wallet

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"

	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/server"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/wallet"
	"github.com/rocket-pool/smartnode/shared/types"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
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
		server.ValidateOptionalArg("password", args, input.ValidateNodePassword, &c.password, &c.passwordExists),
		server.ValidateOptionalArg("save-password", args, input.ValidateBool, &c.savePassword, nil),
	}
	return c, errors.Join(inputErrs...)
}

func (f *walletRecoverContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*walletRecoverContext, api.WalletRecoverData](
		router, "recover", f, f.handler.serviceProvider,
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
	password                 []byte
	passwordExists           bool
	savePassword             bool
}

func (c *walletRecoverContext) PrepareData(data *api.WalletRecoverData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	cfg := sp.GetConfig()
	rp := sp.GetRocketPool()
	w := sp.GetWallet()

	// Requirements
	status := w.GetStatus()
	if status.HasKeystore {
		return fmt.Errorf("a wallet is already present")
	}

	// Use the provided password if there is one
	if c.passwordExists {
		w.RememberPassword(c.password)
		if c.savePassword {
			err := w.SavePassword()
			if err != nil {
				return fmt.Errorf("error saving wallet password to disk: %w", err)
			}
		}
	} else {
		_, hasPassword := w.GetPassword()
		if !hasPassword {
			return fmt.Errorf("you must set a password before recovering a wallet, or provide one in this call")
		}
	}

	if !c.skipValidatorKeyRecovery {
		err := sp.RequireEthClientSynced()
		if err != nil {
			return err
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

	// Recover the wallet
	err := w.Recover(path, uint(c.index), c.mnemonic)
	if err != nil {
		return fmt.Errorf("error recovering wallet: %w", err)
	}
	data.AccountAddress, _ = w.GetAddress()

	// Recover validator keys
	if !c.skipValidatorKeyRecovery {
		data.ValidatorKeys, err = wallet.RecoverMinipoolKeys(cfg, rp, w, false)
		if err != nil {
			return fmt.Errorf("error recovering minipool keys: %w", err)
		}
	}

	return nil
}
