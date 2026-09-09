package megapool

import (
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/units"
)

func canReduceBond(c *cli.Command, amount units.Wei) (*api.CanReduceBondResponse, error) {

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
	details, err := services.GetNodeMegapoolDetails(rp, bc, nodeAccount.Address, nil, false)
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
	gasLimits, err := mp.EstimateReduceBondGas(amount, opts)
	if err != nil {
		return nil, err
	}
	response.GasLimits = gasLimits
	response.CanReduceBond = true

	return &response, nil

}

func reduceBond(c *cli.Command, amount units.Wei, t *snroute.TransactOpts) (*api.ReduceBondResponse, error) {
	opts := t.Opts()

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

	// Reduce bond
	hash, err := mp.ReduceBond(amount, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}

func canReduceBondHandler(ctx snroute.Context) {
	amount, err := parseWei(ctx.Request, "amountWei")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := canReduceBond(ctx.Command(), amount)
	response.WriteResponse(ctx.Writer, resp, err)
}

func reduceBondHandler(ctx snroute.WriteContext) {
	amount, err := parseWei(ctx.Request, "amountWei")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := reduceBond(ctx.Command(), amount, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}
