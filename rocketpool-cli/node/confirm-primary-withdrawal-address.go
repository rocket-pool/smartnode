package node

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

func confirmPrimaryWithdrawalAddress(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check if the withdrawal address can be confirmed
	canResponse, err := rp.CanConfirmNodePrimaryWithdrawalAddress()
	if err != nil {
		return err
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canResponse.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm("Are you sure you want to confirm your node's address as the new primary withdrawal address?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Confirm node's withdrawal address
	response, err := rp.ConfirmNodePrimaryWithdrawalAddress()
	if err != nil {
		return err
	}

	fmt.Printf("Confirming new primary withdrawal address...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("The node's primary withdrawal address was successfully set to the node address.\n")
	return nil

}
