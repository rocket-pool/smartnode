package node

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)

func canSetStakeRplForAllowed(c *cli.Context, caller common.Address, allowed bool) (*api.CanSetStakeRplForAllowedResponse, error) {

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
	response := api.CanSetStakeRplForAllowedResponse{}

	// Get gas estimates
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := node.EstimateSetStakeRPLForAllowedGas(rp, caller, allowed, opts)
	if err != nil {
		return nil, err
	}
	response.GasInfo = gasInfo

	// Update & return response
	response.CanSet = true
	return &response, nil

}

func setStakeRplForAllowed(c *cli.Context, caller common.Address, allowed bool) (*api.SetStakeRplForAllowedResponse, error) {

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
	response := api.SetStakeRplForAllowedResponse{}

	// Stake RPL
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}
	hash, err := node.SetStakeRPLForAllowed(rp, caller, allowed, opts)
	if err != nil {
		return nil, err
	}

	response.SetTxHash = hash

	// Return response
	return &response, nil

}
