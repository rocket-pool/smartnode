package service

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Deletes the contents of the data folder including the wallet file, password file, and all validator keys.
// Don't use this unless you have a very good reason to do it (such as switching from Prater to Mainnet).
func createFeeRecipientFile(c *cli.Context) (*api.CreateFeeRecipientFileResponse, error) {

	// Get services
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.CreateFeeRecipientFileResponse{}

	// Regenerate the fee recipient file
	distributor, err := w.StoreFeeRecipientFile(rp)
	if err != nil {
		return nil, fmt.Errorf("error regenerating fee recipient file: %w", err)
	}

	response.Distributor = distributor

	// Return response
	return &response, nil

}
