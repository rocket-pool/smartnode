package node

import (
	"fmt"

	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)

func canSetRplLockAllowed(c *cli.Context, allowed bool) (*api.CanSetRplLockingAllowedResponse, error) {

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
	// Get the node account
	account, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.CanSetRplLockingAllowedResponse{}

	isAllowed, err := node.GetRPLLockedAllowed(rp, account.Address, nil)
	if err != nil {
		return nil, err
	}

	// Get gas estimates
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := node.EstimateSetStakeRPLForAllowedGas(rp, account.Address, allowed, opts)
	if err != nil {
		return nil, err
	}
	response.GasInfo = gasInfo

	// Update & return response
	response.CanSet = (!isAllowed && allowed) || (isAllowed && !allowed)
	if !response.CanSet {
		if allowed {
			response.Error = "RPL locking is already allowed"
		} else {
			response.Error = "RPL locking is already denied"
		}
	}
	return &response, nil

}

func setRplLockAllowed(c *cli.Context, allowed bool) (*api.SetRplLockingAllowedResponse, error) {

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
	// Get the node account
	account, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.SetRplLockingAllowedResponse{}

	// Stake RPL
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}
	hash, err := node.SetRPLLockingAllowed(rp, account.Address, allowed, opts)
	if err != nil {
		return nil, err
	}

	response.SetTxHash = hash

	// Return response
	return &response, nil

}
