package megapool

import (
	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli"
)

func canDeployMegapool(c *cli.Context) (*api.CanDeployMegapoolResponse, error) {
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
	response := api.CanDeployMegapoolResponse{}

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

	// Check if the megapool is already deployed
	alreadyDeployed, err := megapool.GetMegapoolDeployed(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}
	response.AlreadyDeployed = alreadyDeployed
	if alreadyDeployed {
		response.CanDeploy = false
		return &response, nil
	}

	expectedAddress, err := megapool.GetMegapoolExpectedAddress(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}
	response.ExpectedAddress = expectedAddress

	// Check if the node can deploy a megapool
	gasInfo, err := node.EstimateDeployMegapool(rp, opts)
	if err != nil {
		return nil, err
	}

	// Return response
	response.CanDeploy = !alreadyDeployed
	response.GasInfo = gasInfo
	return &response, nil
}

func deployMegapool(c *cli.Context) (*api.DeployMegapoolResponse, error) {
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
	response := api.DeployMegapoolResponse{}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Deploy megapool
	txHash, err := node.DeployMegapool(rp, opts)
	if err != nil {
		return nil, err
	}

	// Return response
	response.TxHash = txHash
	return &response, nil
}
