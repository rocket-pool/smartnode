package megapool

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

	// Check if Saturn is already deployed
	saturnResp, err := rp.IsSaturnDeployed()
	if err != nil {
		return err
	}
	if !saturnResp.IsSaturnDeployed {
		fmt.Println("This command is only available after the Saturn upgrade.")
		return nil
	}

	// Check if the node can deploy a megapool
	canDeploy, err := rp.CanDeployMegapool()
	if err != nil {
		return err
	}

	if !canDeploy.CanDeploy {
		if canDeploy.AlreadyDeployed {
			fmt.Println("The node already has a megapool deployed.")
		}
		return nil
	}

	fmt.Println("You're about to deploy a megapool contract. It will be used to manage your validators and be the destination for rewards. Both ongoing and upfront costs are reduced by megapools due to the efficiency savings achieved by using a single smart contract over multiple minipool contracts.")
	fmt.Println()

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm("Are you sure you want to deploy a megapool contract?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canDeploy.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	response, err := rp.DeployMegapool()
	if err != nil {
		return err
	}

	fmt.Printf("Deploying megapool...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	fmt.Printf("Megapool deployed successfully at address %s!", canDeploy.ExpectedAddress)
	return nil
}
