package wallet

import (
	"fmt"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func testRecoverWalletWithParams(c *cli.Command, mnemonic string, skipValidatorKeyRecovery bool, derivationPath string, walletIndex uint) (*api.RecoverWalletResponse, error) {

	// Get services
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}
	// Chain clients are only needed to discover and filter validator pubkeys
	var rp *rocketpool.RocketPool
	var bc beacon.Client
	if !skipValidatorKeyRecovery {
		if err := services.RequireRocketStorage(c); err != nil {
			return nil, err
		}
		rp, err = services.GetRocketPool(c)
		if err != nil {
			return nil, err
		}
		bc, err = services.GetBeaconClient(c)
		if err != nil {
			return nil, err
		}
	}

	// Create a blank wallet
	chainId := cfg.Smartnode.GetChainID()
	w, err := wallet.NewWallet("", "", chainId, nil, nil, 0, nil, nil)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.RecoverWalletResponse{}

	// Get the derivation path
	path := derivationPath
	switch path {
	case "":
		path = wallet.DefaultNodeKeyPath
	case "ledgerLive":
		path = wallet.LedgerLiveNodeKeyPath
	case "mew":
		path = wallet.MyEtherWalletNodeKeyPath
	}

	// Recover wallet
	if err := w.TestRecovery(path, walletIndex, mnemonic); err != nil {
		return nil, err
	}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}
	response.AccountAddress = nodeAccount.Address

	if !skipValidatorKeyRecovery {
		response.ValidatorKeys, err = recoverNodeKeys(c, rp, bc, nodeAccount.Address, w, true)
		if err != nil {
			return nil, err
		}
	}

	// Return response
	return &response, nil

}

func testSearchAndRecoverWalletWithParams(c *cli.Command, mnemonic string, address common.Address, skipValidatorKeyRecovery bool) (*api.SearchAndRecoverWalletResponse, error) {

	// Get services
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}
	// Chain clients are only needed to discover and filter validator pubkeys
	var rp *rocketpool.RocketPool
	var bc beacon.Client
	if !skipValidatorKeyRecovery {
		if err := services.RequireRocketStorage(c); err != nil {
			return nil, err
		}
		rp, err = services.GetRocketPool(c)
		if err != nil {
			return nil, err
		}
		bc, err = services.GetBeaconClient(c)
		if err != nil {
			return nil, err
		}
	}

	// Create a blank wallet
	chainId := cfg.Smartnode.GetChainID()
	w, err := wallet.NewWallet("", "", chainId, nil, nil, 0, nil, nil)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.SearchAndRecoverWalletResponse{}

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
	if err := w.TestRecovery(response.DerivationPath, response.Index, mnemonic); err != nil {
		return nil, err
	}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}
	response.AccountAddress = nodeAccount.Address

	if !skipValidatorKeyRecovery {
		response.ValidatorKeys, err = recoverNodeKeys(c, rp, bc, nodeAccount.Address, w, true)
		if err != nil {
			return nil, err
		}
	}

	// Return response
	return &response, nil

}

func testRecoverHandler(ctx snroute.Context) {
	mnemonic := ctx.Request.FormValue("mnemonic")
	skipRecovery := ctx.Request.FormValue("skipValidatorKeyRecovery") == "true"
	derivationPath := ctx.Request.FormValue("derivationPath")
	walletIndex, _ := strconv.ParseUint(ctx.Request.FormValue("walletIndex"), 10, 64)
	resp, err := withRecoveryLock("wallet test-recovery", func() (*api.RecoverWalletResponse, error) {
		return testRecoverWalletWithParams(ctx.Command(), mnemonic, skipRecovery, derivationPath, uint(walletIndex))
	})
	response.WriteResponse(ctx.Writer, resp, err)
}

func testSearchAndRecoverHandler(ctx snroute.Context) {
	mnemonic := ctx.Request.FormValue("mnemonic")
	address := common.HexToAddress(ctx.Request.FormValue("address"))
	skipRecovery := ctx.Request.FormValue("skipValidatorKeyRecovery") == "true"
	resp, err := withRecoveryLock("wallet test-recovery --address", func() (*api.SearchAndRecoverWalletResponse, error) {
		return testSearchAndRecoverWalletWithParams(ctx.Command(), mnemonic, address, skipRecovery)
	})
	response.WriteResponse(ctx.Writer, resp, err)
}
