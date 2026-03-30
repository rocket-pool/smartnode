package odao

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canPenaliseMegapool(c *cli.Command, megapoolAddress common.Address, block *big.Int, amount *big.Int) (*api.CanPenaliseMegapoolResponse, error) {

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

func penaliseMegapool(c *cli.Command, megapoolAddress common.Address, block *big.Int, amount *big.Int, opts *bind.TransactOpts) (*api.PenaliseMegapoolResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.PenaliseMegapoolResponse{}

	// Repay debt
	hash, err := megapool.Penalise(rp, megapoolAddress, block, amount, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}
