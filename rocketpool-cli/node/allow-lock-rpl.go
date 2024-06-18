package node

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

func setRPLLockingAllowed(c *cli.Context, allowedToLock bool) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get the gas estimate
	canResponse, err := rp.CanSetRPLLockingAllowed(allowedToLock)
	if err != nil {
		return err
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canResponse.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	allowString := "Are you sure you want to allow the node to lock RPL when creating governance proposals or to challenge a proposal? Note that the bond could be lost during the proposal challenge process."
	if !allowedToLock {
		allowString = "Are you sure you want to block the node from locking RPL when creating governance proposals or to challenge a proposal? The node won't be able to create proposals or to create challenges."
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm(allowString)) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Set the allow status
	response, err := rp.SetRPLLockingAllowed(allowedToLock)
	if err != nil {
		return err
	}

	fmt.Printf("Submitting the RPL locking transaction...\n")
	cliutils.PrintTransactionHash(rp, response.SetTxHash)
	if _, err = rp.WaitForTransaction(response.SetTxHash); err != nil {
		return err
	}

	// Log & return
	if allowedToLock {
		fmt.Printf("Successfully allowed the node to lock RPL.\n")
	} else {
		fmt.Printf("Successfully blocked the node from locking RPL.\n")
	}

	return nil
}
