package node

import (
	"fmt"
	"math/big"

	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)

func canNodeWithdrawEth(c *cli.Context, amountWei *big.Int) (*api.CanNodeWithdrawEthResponse, error) {

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
	response := api.CanNodeWithdrawEthResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Check for Houston
	isHoustonDeployed, err := state.IsHoustonDeployed(rp, nil)
	if err != nil {
		return nil, fmt.Errorf("error checking if Houston has been deployed: %w", err)
	}

	if !isHoustonDeployed {
		return nil, fmt.Errorf("can only withdraw ETH staked on the node after Houston")
	}

	// Data
	var wg errgroup.Group
	var nodeDetails node.NodeDetails
	var ethStaked *big.Int

	// Get node details
	wg.Go(func() error {
		var err error
		nodeDetails, err = node.GetNodeDetails(rp, nodeAccount.Address, false, nil)
		return err
	})

	// Get RPL stake
	wg.Go(func() error {
		var err error
		ethStaked, err = node.GetNodeEthBalance(rp, nodeAccount.Address, nil)
		return err
	})

	// Get gas estimate
	wg.Go(func() error {
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return err
		}
		gasInfo, err := node.EstimateWithdrawEthGas(rp, nodeAccount.Address, amountWei, opts)
		if err == nil {
			response.GasInfo = gasInfo
		}
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Check data
	response.InsufficientBalance = (amountWei.Cmp(ethStaked) > 0)
	response.HasDifferentWithdrawalAddress = (nodeAccount.Address != nodeDetails.PrimaryWithdrawalAddress)

	// Update & return response
	response.CanWithdraw = !(response.InsufficientBalance || response.HasDifferentWithdrawalAddress)
	return &response, nil

}

func nodeWithdrawEth(c *cli.Context, amountWei *big.Int) (*api.NodeWithdrawRplResponse, error) {

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
	response := api.NodeWithdrawRplResponse{}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Withdraw ETH
	tx, err := node.WithdrawEth(rp, nodeAccount.Address, amountWei, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = tx.Hash()

	// Return response
	return &response, nil

}
