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
		router, "test-recover", f, f.handler.serviceProvider,
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

func (c *walletTestRecoverContext) PrepareData(data *api.WalletRecoverData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	cfg := sp.GetConfig()
	rp := sp.GetRocketPool()

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
	w, err := wallet.TestRecovery(path, uint(c.index), c.mnemonic, cfg.Smartnode.GetChainID())
	if err != nil {
		return fmt.Errorf("error recovering wallet: %w", err)
	}
	data.AccountAddress, _ = w.GetAddress()

	// Recover validator keys
	if !c.skipValidatorKeyRecovery {
		data.ValidatorKeys, err = wallet.RecoverMinipoolKeys(cfg, rp, w, true)
		if err != nil {
			return fmt.Errorf("error recovering minipool keys: %w", err)
		}
	}

	return nil
}
