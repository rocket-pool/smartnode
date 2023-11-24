package node

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/storage"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)

func canSetRPLWithdrawalAddress(c *cli.Context, withdrawalAddress common.Address, confirm bool) (*api.CanSetNodeRPLWithdrawalAddressResponse, error) {
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
	response := api.CanSetNodeRPLWithdrawalAddressResponse{}

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

	// Data
	var wg errgroup.Group
	var primaryWithdrawalAddress common.Address
	var isRPLWithdrawalAddressSet bool
	var rplWithdrawalAddress common.Address
	var rplStake *big.Int

	// Get the primary withdrawal address
	wg.Go(func() error {
		var err error
		primaryWithdrawalAddress, err = storage.GetNodeWithdrawalAddress(rp, nodeAccount.Address, nil)
		return err
	})

	// Check if the RPL withdrawal address is set
	wg.Go(func() error {
		var err error
		isRPLWithdrawalAddressSet, err = node.GetNodeRPLWithdrawalAddressIsSet(rp, nodeAccount.Address, nil)
		return err
	})

	// Get the RPL withdrawal address
	wg.Go(func() error {
		var err error
		rplWithdrawalAddress, err = node.GetNodeRPLWithdrawalAddress(rp, nodeAccount.Address, nil)
		return err
	})

	// Get the RPL stake amount
	wg.Go(func() error {
		var err error
		rplStake, err = node.GetNodeRPLStake(rp, nodeAccount.Address, nil)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Check data
	response.RPLStake = rplStake
	response.PrimaryAddressDiffers = (nodeAccount.Address != primaryWithdrawalAddress || isRPLWithdrawalAddressSet)
	response.RPLAddressDiffers = (isRPLWithdrawalAddressSet && nodeAccount.Address != rplWithdrawalAddress)
	response.CanSet = !(response.PrimaryAddressDiffers || response.RPLAddressDiffers)
	if !response.CanSet {
		return &response, nil
	}

	// Check withdrawal address setting
	gasInfo, err := node.EstimateSetRPLWithdrawalAddressGas(rp, nodeAccount.Address, withdrawalAddress, confirm, opts)
	if err != nil {
		return nil, err
	}
	response.GasInfo = gasInfo

	// Return response
	return &response, nil
}

func setRPLWithdrawalAddress(c *cli.Context, withdrawalAddress common.Address, confirm bool) (*api.SetNodeRPLWithdrawalAddressResponse, error) {
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
	response := api.SetNodeRPLWithdrawalAddressResponse{}

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

	// Get the node's account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Set withdrawal address
	hash, err := node.SetRPLWithdrawalAddress(rp, nodeAccount.Address, withdrawalAddress, confirm, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil
}

func canConfirmRPLWithdrawalAddress(c *cli.Context) (*api.CanConfirmNodeRPLWithdrawalAddressResponse, error) {
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
	response := api.CanConfirmNodeRPLWithdrawalAddressResponse{}

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

	// Make sure the pending withdrawal address is set to the node address
	pendingAddress, err := node.GetNodePendingRPLWithdrawalAddress(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	response.CanConfirm = (pendingAddress != nodeAccount.Address)
	if !response.CanConfirm {
		return &response, nil
	}

	// Check withdrawal address setting
	gasInfo, err := node.EstimateConfirmRPLWithdrawalAddressGas(rp, nodeAccount.Address, opts)
	if err != nil {
		return nil, err
	}
	response.GasInfo = gasInfo

	// Return response
	return &response, nil
}

func confirmRPLWithdrawalAddress(c *cli.Context) (*api.ConfirmNodeRPLWithdrawalAddressResponse, error) {
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
	response := api.ConfirmNodeRPLWithdrawalAddressResponse{}

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

	// Get the node's account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Set withdrawal address
	hash, err := node.ConfirmRPLWithdrawalAddress(rp, nodeAccount.Address, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil
}
