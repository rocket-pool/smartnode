package node

import (
	"math/big"

	"github.com/rocket-pool/smartnode/shared/services"
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

	// Response
	response := api.CheckCollateralResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Check collateral
	ethMatched, ethMatchedLimit, pendingMatchAmount, err := rputils.CheckCollateral(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}
	response.EthMatched.Set(ethMatched)
	response.EthMatchedLimit.Set(ethMatchedLimit)
	response.PendingMatchAmount.Set(pendingMatchAmount)

	// Check if there's sufficient collateral including pending bond reductions
	remainingMatch := big.NewInt(0).Sub(&response.EthMatchedLimit, &response.EthMatched)
	remainingMatch.Sub(remainingMatch, &response.PendingMatchAmount)
	response.InsufficientCollateral = (remainingMatch.Cmp(big.NewInt(0)) < 0)

	return &response, nil
}
