package wallet

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/node/wallet"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
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
		server.ValidateArg("password", args, input.ValidateNodePassword, &c.password),
		server.ValidateArg("save-password", args, input.ValidateBool, &c.savePassword),
	}
	return c, errors.Join(inputErrs...)
}

func (f *walletSearchAndRecoverContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*walletSearchAndRecoverContext, api.WalletSearchAndRecoverData](
		router, "search-and-recover", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
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
	password                 string
	passwordExists           bool
	savePassword             bool
}

func (c *walletSearchAndRecoverContext) PrepareData(data *api.WalletSearchAndRecoverData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	w := sp.GetWallet()
	rs := sp.GetNetworkResources()
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

	// Try each derivation path across all of the iterations
	paths := []string{
		wallet.DefaultNodeKeyPath,
		wallet.LedgerLiveNodeKeyPath,
		wallet.MyEtherWalletNodeKeyPath,
	}
	for i := uint(0); i < findIterations; i++ {
		for j := 0; j < len(paths); j++ {
			derivationPath := paths[j]
			recoveredWallet, err := wallet.TestRecovery(derivationPath, i, c.mnemonic, rs.ChainID)
			if err != nil {
				return types.ResponseStatus_Error, fmt.Errorf("error recovering wallet with path [%s], index [%d]: %w", derivationPath, i, err)
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
		return types.ResponseStatus_ResourceNotFound, fmt.Errorf("exhausted all derivation paths and indices from 0 to %d, wallet not found", findIterations)
	}

	// Recover the wallet
	err = w.Recover(data.DerivationPath, data.Index, c.mnemonic, c.password, c.savePassword, false)
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
