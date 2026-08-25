package odao

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"

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

	gasLimits, err := megapool.EstimatePenaliseGas(rp, megapoolAddress, block, amount, opts)
	if err != nil {
		return nil, err
	}
	response.GasLimits = gasLimits
	response.CanPenalise = true

	return &response, nil

}

func penaliseMegapool(c *cli.Command, megapoolAddress common.Address, block *big.Int, amount *big.Int, t *snroute.TransactOpts) (*api.PenaliseMegapoolResponse, error) {
	opts := t.Opts()

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

func canPenaliseMegapoolHandler(ctx snroute.Context) {
	megapool, block, amount, err := parsePenaliseParams(ctx.Request)
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := canPenaliseMegapool(ctx.Command(), megapool, block, amount)
	response.WriteResponse(ctx.Writer, resp, err)
}

func penaliseMegapoolHandler(ctx snroute.WriteContext) {
	megapool, block, amount, err := parsePenaliseParams(ctx.Request)
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := penaliseMegapool(ctx.Command(), megapool, block, amount, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}
