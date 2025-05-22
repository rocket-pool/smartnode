package wallet

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/types/api"
	walletutils "github.com/rocket-pool/smartnode/shared/utils/wallet"
)

const (
	findIterations uint = 100000
)

func recoverWallet(c *cli.Context, mnemonic string) (*api.RecoverWalletResponse, error) {

	// Get services
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	var rp *rocketpool.RocketPool
	if !c.Bool("skip-validator-key-recovery") {
		if err := services.RequireRocketStorage(c); err != nil {
			return nil, err
		}
		rp, err = services.GetRocketPool(c)
		if err != nil {
			return nil, err
		}
	}

	// Response
	response := api.RecoverWalletResponse{}

	// Check if wallet is already initialized
	if w.IsInitialized() {
		return nil, errors.New("the wallet is already initialized")
	}

	// Get the derivation path
	path := c.String("derivation-path")
	switch path {
	case "":
		path = wallet.DefaultNodeKeyPath
	case "ledgerLive":
		path = wallet.LedgerLiveNodeKeyPath
	case "mew":
		path = wallet.MyEtherWalletNodeKeyPath
	}

	// Get the wallet index
	walletIndex := c.Uint("wallet-index")

	// Recover wallet
	if err := w.Recover(path, walletIndex, mnemonic); err != nil {
		return nil, err
	}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}
	response.AccountAddress = nodeAccount.Address

	if !c.Bool("skip-validator-key-recovery") {
		response.ValidatorKeys, err = walletutils.RecoverNodeKeys(c, rp, nodeAccount.Address, w, false)
		if err != nil {
			return nil, err
		}
	}

	// Save wallet
	if err := w.Save(); err != nil {
		return nil, err
	}

	// Return response
	return &response, nil

}

func searchAndRecoverWallet(c *cli.Context, mnemonic string, address common.Address) (*api.SearchAndRecoverWalletResponse, error) {

	// Get services
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	var rp *rocketpool.RocketPool
	if !c.Bool("skip-validator-key-recovery") {
		if err := services.RequireRocketStorage(c); err != nil {
			return nil, err
		}
		rp, err = services.GetRocketPool(c)
		if err != nil {
			return nil, err
		}
	}

	// Response
	response := api.SearchAndRecoverWalletResponse{}

	// Check if wallet is already initialized
	if w.IsInitialized() {
		return nil, errors.New("the wallet is already initialized")
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
			recoveredWallet, err := wallet.NewWallet("", "", uint(w.GetChainID().Uint64()), nil, nil, 0, nil, nil)
			if err != nil {
				return nil, fmt.Errorf("error generating new wallet: %w", err)
			}
			err = recoveredWallet.TestRecovery(derivationPath, i, mnemonic)
			if err != nil {
				return nil, fmt.Errorf("error recovering wallet with path [%s], index [%d]: %w", derivationPath, i, err)
			}

			// Get recovered account
			recoveredAccount, err := recoveredWallet.GetNodeAccount()
			if err != nil {
				return nil, fmt.Errorf("error getting recovered account: %w", err)
			}
			if recoveredAccount.Address == address {
				// We found the correct derivation path and index
				response.FoundWallet = true
				response.DerivationPath = derivationPath
				response.Index = i
				break
			}
		}
		if response.FoundWallet {
			break
		}
	}

	if !response.FoundWallet {
		return nil, fmt.Errorf("exhausted all derivation paths and indices from 0 to %d, wallet not found", findIterations)
	}

	// Recover wallet
	if err := w.Recover(response.DerivationPath, response.Index, mnemonic); err != nil {
		return nil, err
	}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}
	response.AccountAddress = nodeAccount.Address

	if !c.Bool("skip-validator-key-recovery") {
		response.ValidatorKeys, err = walletutils.RecoverNodeKeys(c, rp, nodeAccount.Address, w, false)
		if err != nil {
			return nil, err
		}
	}

	// Save wallet
	if err := w.Save(); err != nil {
		return nil, err
	}

	// Return response
	return &response, nil

}
