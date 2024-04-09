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

// ===============
// === Factory ===
// ===============

type walletTestSearchAndRecoverContextFactory struct {
	handler *WalletHandler
}

func (f *walletTestSearchAndRecoverContextFactory) Create(args url.Values) (*walletTestSearchAndRecoverContext, error) {
	c := &walletTestSearchAndRecoverContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("mnemonic", args, input.ValidateWalletMnemonic, &c.mnemonic),
		server.ValidateArg("address", args, input.ValidateAddress, &c.address),
		server.ValidateOptionalArg("skip-validator-key-recovery", args, input.ValidateBool, &c.skipValidatorKeyRecovery, nil),
	}
	return c, errors.Join(inputErrs...)
}

func (f *walletTestSearchAndRecoverContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*walletTestSearchAndRecoverContext, api.WalletSearchAndRecoverData](
		router, "test-search-and-recover", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type walletTestSearchAndRecoverContext struct {
	handler                  *WalletHandler
	skipValidatorKeyRecovery bool
	mnemonic                 string
	address                  common.Address
}

func (c *walletTestSearchAndRecoverContext) PrepareData(data *api.WalletSearchAndRecoverData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	rs := sp.GetNetworkResources()
	vMgr := sp.GetValidatorManager()

	if !c.skipValidatorKeyRecovery {
		status, err := sp.RequireRocketPoolContracts(c.handler.ctx)
		if err != nil {
			return status, err
		}
	}

	// Try each derivation path across all of the iterations
	var recoveredWallet *wallet.Wallet
	paths := []string{
		wallet.DefaultNodeKeyPath,
		wallet.LedgerLiveNodeKeyPath,
		wallet.MyEtherWalletNodeKeyPath,
	}
	for i := uint(0); i < findIterations; i++ {
		for j := 0; j < len(paths); j++ {
			var err error
			derivationPath := paths[j]
			recoveredWallet, err = wallet.TestRecovery(derivationPath, i, c.mnemonic, rs.ChainID)
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
	data.AccountAddress, _ = recoveredWallet.GetAddress()

	// Recover validator keys
	if !c.skipValidatorKeyRecovery {
		var err error
		data.ValidatorKeys, err = vMgr.RecoverMinipoolKeys(true)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error recovering minipool keys: %w", err)
		}
	}

	return types.ResponseStatus_Success, nil
}
