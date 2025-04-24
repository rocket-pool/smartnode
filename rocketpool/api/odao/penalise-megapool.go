package odao

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/megapool"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)

func canPenaliseMegapool(c *cli.Context, megapoolAddress common.Address, block *big.Int, amount *big.Int) (*api.CanPenaliseMegapoolResponse, error) {

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
	response := api.CanPenaliseMegapoolResponse{}

	// Check if the megapool is deployed
	megapoolDeployed, err := megapool.GetMegapoolDeployed(rp, megapoolAddress, nil)
	if err != nil {
		return nil, err
	}
	if !megapoolDeployed {
		response.CanPenalise = false
		return &response, nil
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := megapool.EstimatePenaliseGas(rp, megapoolAddress, block, amount, opts)
	if err != nil {
		return nil, err
	}
	response.GasInfo = gasInfo
	response.CanPenalise = true

	return &response, nil

}

func penaliseMegapool(c *cli.Context, megapoolAddress common.Address, block *big.Int, amount *big.Int) (*api.PenaliseMegapoolResponse, error) {

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
	response := api.PenaliseMegapoolResponse{}

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

	// Repay debt
	hash, err := megapool.Penalise(rp, megapoolAddress, block, amount, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}
