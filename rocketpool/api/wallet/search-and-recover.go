package wallet

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/rocketpool/common/wallet"
	"github.com/rocket-pool/smartnode/shared/types"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
)

const (
	findIterations uint = 100000
)

// ===============
// === Factory ===
// ===============

type walletSearchAndRecoverContextFactory struct {
	handler *WalletHandler
}

func (f *walletSearchAndRecoverContextFactory) Create(args url.Values) (*walletSearchAndRecoverContext, error) {
	c := &walletSearchAndRecoverContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("mnemonic", args, input.ValidateWalletMnemonic, &c.mnemonic),
		server.ValidateArg("address", args, input.ValidateAddress, &c.address),
		server.ValidateOptionalArg("skip-validator-key-recovery", args, input.ValidateBool, &c.skipValidatorKeyRecovery, nil),
		server.ValidateOptionalArg("password", args, input.ValidateNodePassword, &c.password, &c.passwordExists),
		server.ValidateOptionalArg("save-password", args, input.ValidateBool, &c.savePassword, nil),
	}
	return c, errors.Join(inputErrs...)
}

func (f *walletSearchAndRecoverContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*walletSearchAndRecoverContext, api.WalletSearchAndRecoverData](
		router, "search-and-recover", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type walletSearchAndRecoverContext struct {
	handler                  *WalletHandler
	skipValidatorKeyRecovery bool
	mnemonic                 string
	address                  common.Address
	password                 []byte
	passwordExists           bool
	savePassword             bool
}

func (c *walletSearchAndRecoverContext) PrepareData(data *api.WalletSearchAndRecoverData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	cfg := sp.GetConfig()
	rp := sp.GetRocketPool()
	w := sp.GetWallet()

	// Requirements
	switch w.GetStatus() {
	case types.WalletStatus_Ready, types.WalletStatus_KeystoreMismatch:
		return fmt.Errorf("a wallet is already present")
	default:
		_, hasPassword := w.GetPassword()
		if !hasPassword && !c.passwordExists {
			return fmt.Errorf("you must set a password before recovering a wallet, or provide one in this call")
		}
		w.RememberPassword(c.password)
		if c.savePassword {
			err := w.SavePassword()
			if err != nil {
				return fmt.Errorf("error saving wallet password to disk: %w", err)
			}
		}
	}
	if !c.skipValidatorKeyRecovery {
		err := sp.RequireEthClientSynced()
		if err != nil {
			return err
		}
	}

	// Try each derivation path across all of the iterations
	paths := []string{
		wallet.DefaultNodeKeyPath,
		wallet.LedgerLiveNodeKeyPath,
		wallet.MyEtherWalletNodeKeyPath,
	}
	for i := uint(0); i < findIterations; i++ {
		for j := 0; j < len(paths); j++ {
			derivationPath := paths[j]
			recoveredWallet, err := wallet.TestRecovery(derivationPath, i, c.mnemonic, cfg.Smartnode.GetChainID())
			if err != nil {
				return fmt.Errorf("error recovering wallet with path [%s], index [%d]: %w", derivationPath, i, err)
			}

			// Get recovered account
			recoveredAddress, _ := recoveredWallet.GetAddress()
			if recoveredAddress == c.address {
				// We found the correct derivation path and index
				data.FoundWallet = true
				data.DerivationPath = derivationPath
				data.Index = i
				break
			}
		}
		if data.FoundWallet {
			break
		}
	}

	if !data.FoundWallet {
		return fmt.Errorf("exhausted all derivation paths and indices from 0 to %d, wallet not found", findIterations)
	}

	// Recover the wallet
	err := w.Recover(data.DerivationPath, data.Index, c.mnemonic)
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
