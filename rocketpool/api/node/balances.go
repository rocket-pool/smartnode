package node

import (
	"context"
	"fmt"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli"
)

func getNodeEthBalance(c *cli.Context) (*api.NodeEthBalanceResponse, error) {
	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	ec, err := services.GetEthClient(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.NodeEthBalanceResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	response.Balance, err = ec.BalanceAt(context.Background(), nodeAccount.Address, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting ETH balance of node %s: %w", nodeAccount.Address.Hex(), err)
	}

	return &response, nil
}
