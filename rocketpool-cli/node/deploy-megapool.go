package node

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/urfave/cli"
)

func deployMegapool(c *cli.Context) error {
	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check if the node can deploy a megapool
	canDeploy, err := rp.CanDeployMegapool()
	if err != nil {
		return err
	}

	// Check if Saturn is already deployed
	saturnResp, err := rp.IsSaturnDeployed()
	if err != nil {
		return err
	}
	if !saturnResp.IsSaturnDeployed {
		fmt.Println("This command is only available after the Saturn upgrade.")
		return nil
	}

	if !canDeploy.CanDeploy {
		if canDeploy.AlreadyDeployed {
			fmt.Println("The node already has a Megapool deployed.")
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canDeploy.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm("Are you sure you want to deploy a Megapool?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	response, err := rp.DeployMegapool()
	if err != nil {
		return err
	}

	fmt.Printf("Deploying Megapool...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	fmt.Println("Megapool deployed successfully!")
	return nil
}
