package wallet

import (
	"fmt"

	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func rebuildWallet(c *cli.Context) (*api.RebuildWalletResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.RebuildWalletResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

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

	// Save wallet
	if err := w.Save(); err != nil {
		return nil, err
	}

	// Regenerate the fee recipient file
	_, err = w.StoreFeeRecipientFile(rp)
	if err != nil {
		return nil, fmt.Errorf("error regenerating fee recipient file: %w", err)
	}

	// Return response
	return &response, nil

}
