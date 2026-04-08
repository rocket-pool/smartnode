package node

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/storage"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canSetPrimaryWithdrawalAddress(c *cli.Command, withdrawalAddress common.Address, confirm bool) (*api.CanSetNodePrimaryWithdrawalAddressResponse, error) {

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
	response := api.CanSetNodePrimaryWithdrawalAddressResponse{}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Get the node's account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Check withdrawal address setting
	gasInfo, err := storage.EstimateSetWithdrawalAddressGas(rp, nodeAccount.Address, withdrawalAddress, confirm, opts)
	if err != nil {
		return nil, err
	}
	response.GasInfo = gasInfo

	// Return response
	response.CanSet = true
	return &response, nil
}

func setPrimaryWithdrawalAddress(c *cli.Command, withdrawalAddress common.Address, confirm bool, opts *bind.TransactOpts) (*api.SetNodePrimaryWithdrawalAddressResponse, error) {

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
	response := api.SetNodePrimaryWithdrawalAddressResponse{}

	// Get the node's account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Make sure the current withdrawal address is set to the node address
	currentAddress, err := storage.GetNodeWithdrawalAddress(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}
	if currentAddress != nodeAccount.Address {
		return nil, fmt.Errorf("This wallet's current withdrawal address is %s, "+
			"so you cannot call set-withdrawal-address from the node.", currentAddress.String())
	}

	// Set withdrawal address
	hash, err := storage.SetWithdrawalAddress(rp, nodeAccount.Address, withdrawalAddress, confirm, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}

func canConfirmPrimaryWithdrawalAddress(c *cli.Command) (*api.CanConfirmNodePrimaryWithdrawalAddressResponse, error) {

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
	response := api.CanConfirmNodePrimaryWithdrawalAddressResponse{}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Get the node's account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Make sure the current withdrawal address is set to the node address
	pendingAddress, err := storage.GetNodePendingWithdrawalAddress(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	// Check withdrawal address setting
	gasInfo, err := storage.EstimateConfirmWithdrawalAddressGas(rp, nodeAccount.Address, opts)
	if err != nil {
		return nil, err
	}
	response.GasInfo = gasInfo

	// Return response
	response.CanConfirm = (pendingAddress != nodeAccount.Address)
	return &response, nil
}

func confirmPrimaryWithdrawalAddress(c *cli.Command, opts *bind.TransactOpts) (*api.ConfirmNodePrimaryWithdrawalAddressResponse, error) {

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
	response := api.ConfirmNodePrimaryWithdrawalAddressResponse{}

	// Get the node's account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Make sure the current withdrawal address is set to the node address
	pendingAddress, err := storage.GetNodePendingWithdrawalAddress(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}
	if pendingAddress != nodeAccount.Address {
		return nil, fmt.Errorf("This wallet's pending withdrawal address is %s, "+
			"which is not the node address.", pendingAddress.String())
	}

	// Set withdrawal address
	hash, err := storage.ConfirmWithdrawalAddress(rp, nodeAccount.Address, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}
