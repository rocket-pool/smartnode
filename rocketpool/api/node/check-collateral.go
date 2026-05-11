package node

import (
	"math/big"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	rputils "github.com/rocket-pool/smartnode/shared/utils/rp"
)

func checkCollateral(c *cli.Command) (*api.CheckCollateralResponse, error) {
	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
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
	response := api.CheckCollateralResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Check collateral
	response.EthBorrowed, response.EthBorrowedLimit, response.PendingBorrowAmount, err = rputils.CheckCollateral(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	// Check if there's sufficient collateral including pending bond reductions
	remainingBorrow := big.NewInt(0).Sub(response.EthBorrowedLimit, response.EthBorrowed)
	remainingBorrow.Sub(remainingBorrow, response.PendingBorrowAmount)
	response.InsufficientCollateral = (remainingBorrow.Cmp(big.NewInt(0)) < 0)

	return &response, nil
}
