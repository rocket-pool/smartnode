package megapool

import (
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canDeploy(c *cli.Command) (*api.CanDeployMegapoolResponse, error) {
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
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	response := api.CanDeployMegapoolResponse{}
	deployed, err := megapool.GetMegapoolDeployed(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}
	address, err := megapool.GetMegapoolExpectedAddress(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}
	response.AlreadyDeployed = deployed
	response.MegapoolAddress = address
	if deployed {
		return &response, nil
	}

	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasLimits, err := node.EstimateDeployMegapoolGas(rp, opts)
	if err != nil {
		return nil, err
	}
	response.GasLimits = gasLimits
	response.CanDeploy = true
	return &response, nil
}

func deploy(c *cli.Command, t *snroute.TransactOpts) (*api.DeployMegapoolResponse, error) {
	opts := t.Opts()
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	hash, err := node.DeployMegapool(rp, opts)
	if err != nil {
		return nil, err
	}
	return &api.DeployMegapoolResponse{TxHash: hash}, nil
}

func canDeployHandler(ctx snroute.Context) {
	resp, err := canDeploy(ctx.Command())
	response.WriteResponse(ctx.Writer, resp, err)
}

func deployHandler(ctx snroute.WriteContext) {
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := deploy(ctx.Command(), opts)
	response.WriteResponse(ctx.Writer, resp, err)
}
