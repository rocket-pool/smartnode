package megapool

import (
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canDistributeMegapool(c *cli.Command) (*api.CanDistributeMegapoolResponse, error) {
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

	// Response
	response := api.CanDistributeMegapoolResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Load the megapool details
	details, err := services.GetNodeMegapoolDetails(rp, bc, nodeAccount.Address, nil, false)
	if err != nil {
		return nil, err
	}

	if !details.Deployed {
		response.CanDistribute = false
		response.MegapoolNotDeployed = true
		return &response, nil
	}

	response.MegapoolAddress = details.Address
	response.Details = details

	// Load the megapool
	mp, err := megapool.NewMegaPoolV1(rp, response.MegapoolAddress, nil)
	if err != nil {
		return nil, err
	}

	response.LastDistributionTime = details.LastDistributionTime

	if response.LastDistributionTime == 0 {
		response.CanDistribute = false
		return &response, nil
	}

	if details.LockedValidatorCount > 0 {
		response.CanDistribute = false
		response.LockedValidatorCount = details.LockedValidatorCount
		return &response, nil
	}

	if details.ExitingValidatorCount > 0 {
		response.ExitingValidatorCount = details.ExitingValidatorCount
		response.CanDistribute = false
		return &response, nil
	}

	gasLimits, err := mp.EstimateDistributeGas(opts)
	if err != nil {
		return nil, err
	}
	// Return response
	response.CanDistribute = true
	response.GasLimits = gasLimits
	return &response, nil
}

func distributeMegapool(c *cli.Command, t *snroute.TransactOpts) (*api.DistributeMegapoolResponse, error) {
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
	response := api.DistributeMegapoolResponse{}

	megapoolAddress, err := megapool.GetMegapoolExpectedAddress(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	// Load the megapool
	mp, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Distribute
	hash, err := mp.Distribute(opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil
}

func canDistributeHandler(ctx snroute.Context) {
	resp, err := canDistributeMegapool(ctx.Command())
	response.WriteResponse(ctx.Writer, resp, err)
}

func distributeHandler(ctx snroute.WriteContext) {
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := distributeMegapool(ctx.Command(), opts)
	response.WriteResponse(ctx.Writer, resp, err)
}
