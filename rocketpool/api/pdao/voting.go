package pdao

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"

	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)

func estimateSetVotingDelegateGas(c *cli.Context, address common.Address) (*api.NetworkCanSetVotingDelegateResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	// Response

	response := api.NetworkCanSetVotingDelegateResponse{}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Get the gas info
	gasInfo, err := network.EstimateSetVotingDelegateGas(rp, address, opts)
	if err != nil {
		return nil, err
	}
	response.GasInfo = gasInfo

	// Return response
	return &response, nil

}

func setVotingDelegate(c *cli.Context, address common.Address) (*api.NetworkSetVotingDelegateResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.NetworkSetVotingDelegateResponse{}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Set the delegate
	tx, err := network.SetVotingDelegate(rp, address, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = tx

	// Return response
	return &response, nil

}

func getCurrentVotingDelegate(c *cli.Context) (*api.NetworkCurrentVotingDelegateResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.NetworkCurrentVotingDelegateResponse{}
	response.AccountAddress = nodeAccount.Address

	// Set the delegate
	delegate, err := network.GetCurrentVotingDelegate(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}
	response.VotingDelegate = delegate

	// Return response
	return &response, nil

}
