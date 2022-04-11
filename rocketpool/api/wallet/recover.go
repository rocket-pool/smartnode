package wallet

import (
	"errors"

	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func recoverWallet(c *cli.Context, mnemonic string) (*api.RecoverWalletResponse, error) {

	// Get services
	if err := services.RequireNodePassword(c); err != nil {
		return nil, err
	}
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
		return nil, errors.New("The wallet is already initialized")
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

	// Recover wallet
	if err := w.Recover(path, mnemonic); err != nil {
		return nil, err
	}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}
	response.AccountAddress = nodeAccount.Address

	if !c.Bool("skip-validator-key-recovery") {
		// Get node's validating pubkeys
		pubkeys, err := minipool.GetNodeValidatingMinipoolPubkeys(rp, nodeAccount.Address, nil)
		if err != nil {
			return nil, err
		}
		response.ValidatorKeys = pubkeys

		// Recover validator keys
		for _, pubkey := range pubkeys {
			if err := w.RecoverValidatorKey(pubkey); err != nil {
				return nil, err
			}
		}
	}

	// Save wallet
	if err := w.Save(); err != nil {
		return nil, err
	}

	// Return response
	return &response, nil

}
