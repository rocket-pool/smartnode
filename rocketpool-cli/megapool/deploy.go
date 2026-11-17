package megapool

import (
	"fmt"

	cliutils "github.com/rocket-pool/smartnode/rocketpool-cli/cli"
	"github.com/rocket-pool/smartnode/rocketpool-cli/cli/color"
	"github.com/rocket-pool/smartnode/rocketpool-cli/cli/prompt"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
)

func deployMegapool(yes bool) error {
	rp, err := rocketpool.NewClient().WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	canDeploy, err := rp.CanDeployMegapool()
	if err != nil {
		return err
	}
	if canDeploy.AlreadyDeployed {
		fmt.Printf("The node already has a megapool deployed at %s.\n", color.LightBlue(canDeploy.MegapoolAddress.Hex()))
		return nil
	}
	if !canDeploy.CanDeploy {
		fmt.Println("The node cannot deploy a megapool.")
		return nil
	}

	fmt.Printf("This will deploy the node's megapool at %s.\n", color.LightBlue(canDeploy.MegapoolAddress.Hex()))
	if err = gas.AssignMaxFeeAndLimit(canDeploy.GasLimits, rp, yes); err != nil {
		return err
	}
	if prompt.Declined(yes, "Are you sure you want to deploy your megapool?") {
		fmt.Println("Cancelled.")
		return nil
	}

	response, err := rp.DeployMegapool()
	if err != nil {
		return err
	}
	fmt.Println("Deploying megapool...")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}
	fmt.Printf("Successfully deployed the megapool at %s.\n", color.LightBlue(canDeploy.MegapoolAddress.Hex()))
	return nil
}
