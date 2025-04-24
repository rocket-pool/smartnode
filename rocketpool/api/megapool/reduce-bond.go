package megapool

import (
	"fmt"
	"math/big"

	"github.com/rocket-pool/rocketpool-go/megapool"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)

func canReduceBond(c *cli.Context, amount *big.Int) (*api.CanReduceBondResponse, error) {

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
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	// Get the node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.CanReduceBondResponse{}

	// Check if the megapool is deployed
	megapoolDeployed, err := megapool.GetMegapoolDeployed(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}
	if !megapoolDeployed {
		response.CanReduceBond = false
		return &response, nil
	}

	// Get the megapool address
	megapoolAddress, err := megapool.GetMegapoolExpectedAddress(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	// Load the megapool details
	details, err := GetNodeMegapoolDetails(rp, bc, nodeAccount.Address)
	if err != nil {
		return nil, err
	}

	mp, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Check if the node bond is greater than the current required bond
	if details.NodeBond.Cmp(details.BondRequirement) <= 0 {
		response.CanReduceBond = false
		response.NotEnoughBond = true
		return &response, nil
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := mp.EstimateReduceBondGas(amount, opts)
	if err != nil {
		return nil, err
	}
	response.GasInfo = gasInfo
	response.CanReduceBond = true

	return &response, nil

}

func reduceBond(c *cli.Context, amount *big.Int) (*api.ReduceBondResponse, error) {

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
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ReduceBondResponse{}

	// Get the megapool address
	megapoolAddress, err := megapool.GetMegapoolExpectedAddress(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	// Load the megapool
	mp, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
	if err != nil {
		return nil, err
	}

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

	// Reduce bond
	hash, err := mp.ReduceBond(amount, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}
