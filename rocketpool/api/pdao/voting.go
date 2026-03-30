package pdao

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/network"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func estimateSetVotingDelegateGas(c *cli.Command, address common.Address) (*api.PDAOCanSetVotingDelegateResponse, error) {

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

	response := api.PDAOCanSetVotingDelegateResponse{}

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

func setVotingDelegate(c *cli.Command, address common.Address, opts *bind.TransactOpts) (*api.PDAOSetVotingDelegateResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	// Response
	response := api.PDAOSetVotingDelegateResponse{}

	// Set the delegate
	tx, err := network.SetVotingDelegate(rp, address, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = tx

	// Return response
	return &response, nil

}

func getCurrentVotingDelegate(c *cli.Command) (*api.PDAOCurrentVotingDelegateResponse, error) {

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
	response := api.PDAOCurrentVotingDelegateResponse{}
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
