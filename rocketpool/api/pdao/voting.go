package pdao

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/network"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
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
	gasLimits, err := network.EstimateSetVotingDelegateGas(rp, address, opts)
	if err != nil {
		return nil, err
	}
	response.GasLimits = gasLimits

	// Return response
	return &response, nil

}

func setVotingDelegate(c *cli.Command, address common.Address, t *snroute.TransactOpts) (*api.PDAOSetVotingDelegateResponse, error) {
	opts := t.Opts()

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

func estimateSetVotingDelegateGasHandler(ctx snroute.Context) {
	addr := common.HexToAddress(paramVal(ctx.Request, "address"))
	resp, err := estimateSetVotingDelegateGas(ctx.Command(), addr)
	response.WriteResponse(ctx.Writer, resp, err)
}

func setVotingDelegateHandler(ctx snroute.WriteContext) {
	addr := common.HexToAddress(paramVal(ctx.Request, "address"))
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := setVotingDelegate(ctx.Command(), addr, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}

func getCurrentVotingDelegateHandler(ctx snroute.Context) {
	resp, err := getCurrentVotingDelegate(ctx.Command())
	response.WriteResponse(ctx.Writer, resp, err)
}
