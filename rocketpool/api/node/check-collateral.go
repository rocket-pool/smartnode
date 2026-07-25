package node

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
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
	response.EthBorrowed, response.EthBorrowedLimit, response.PendingBorrowAmount, err = CheckCollateral(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	// Check if there's sufficient collateral including pending bond reductions
	remainingBorrow := big.NewInt(0).Sub(response.EthBorrowedLimit, response.EthBorrowed)
	remainingBorrow.Sub(remainingBorrow, response.PendingBorrowAmount)
	response.InsufficientCollateral = (remainingBorrow.Cmp(big.NewInt(0)) < 0)

	return &response, nil
}

// Checks the given node's current borrowed ETH, its limit on borrowed ETH, and how much ETH is preparing to be borrowed by pending bond reductions
func CheckCollateral(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (ethBorrowed *big.Int, ethBorrowedLimit *big.Int, pendingBorrowAmount *big.Int, err error) {
	ethBorrowed, err = node.GetNodeETHBorrowed(rp, nodeAddress, opts)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error getting node's borrowed ETH amount: %w", err)
	}

	return
}
