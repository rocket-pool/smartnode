package node

import (
	"math/big"

	"github.com/rocket-pool/smartnode/shared/services"
	updateCheck "github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/types/api"
	rputils "github.com/rocket-pool/smartnode/shared/utils/rp"
	"github.com/urfave/cli"
)

func checkCollateral(c *cli.Context) (*api.CheckCollateralResponse, error) {
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

	// Check if Saturn is already deployed
	saturnDeployed, err := updateCheck.IsSaturnDeployed(rp, nil)
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
	response.EthMatched, response.EthMatchedLimit, response.PendingMatchAmount, err = rputils.CheckCollateral(saturnDeployed, rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	// Check if there's sufficient collateral including pending bond reductions
	remainingMatch := big.NewInt(0).Sub(response.EthMatchedLimit, response.EthMatched)
	remainingMatch.Sub(remainingMatch, response.PendingMatchAmount)
	response.InsufficientCollateral = (remainingMatch.Cmp(big.NewInt(0)) < 0)

	return &response, nil
}
