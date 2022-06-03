package node

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

func joinSmoothingPool(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c)
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check and assign the EC status
	err = cliutils.CheckExecutionClientStatus(rp)
	if err != nil {
		return err
	}

	// Get the node's registration status
	status, err := rp.NodeGetSmoothingPoolRegistrationStatus()
	if err != nil {
		return err
	}

	if status.NodeRegistered {
		fmt.Println("The node is already joined to the Smoothing Pool.")
		return nil
	}

	if status.TimeLeftUntilChangeable > 0 {
		fmt.Printf("You have recently left the Smoothing Pool. You must wait %s until you can join it again.\n", status.TimeLeftUntilChangeable)
		return nil
	}

	// Print some info
	fmt.Println("You are about to opt into the Smoothing Pool.\nYour fee recipient will be changed to the Smoothing Pool contract.\nAll priority fees and MEV you earn via proposals will be shared equally with other members of the Smoothing Pool.\n\nIf you desire, you can opt back out after one full rewards interval has passed.\n")

	// Get the gas estimate
	canResponse, err := rp.CanNodeSetSmoothingPoolStatus(true)
	if err != nil {
		return err
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canResponse.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	fmt.Printf("%sNOTE: This process will restart your node's validator client.\nYou may miss an attestation if you are currently scheduled to produce one.%s\n\n", colorYellow, colorReset)

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm("Are you sure you want to join the Smoothing Pool?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Set the fee recipient to the Smoothing Pool
	response, err := rp.NodeSetSmoothingPoolStatus(true)
	if err != nil {
		return err
	}

	fmt.Printf("Joining the Smoothing Pool...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Println("Successfully joined the Smoothing Pool.")
	return nil

}

func leaveSmoothingPool(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c)
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check and assign the EC status
	err = cliutils.CheckExecutionClientStatus(rp)
	if err != nil {
		return err
	}

	// Get the node's registration status
	status, err := rp.NodeGetSmoothingPoolRegistrationStatus()
	if err != nil {
		return err
	}

	if !status.NodeRegistered {
		fmt.Println("The node is not currently joined to the Smoothing Pool.")
		return nil
	}

	if status.TimeLeftUntilChangeable > 0 {
		fmt.Printf("You have recently joined the Smoothing Pool. You must wait %s until you can leave it.\n", status.TimeLeftUntilChangeable)
		return nil
	}

	// Print some info
	fmt.Println("You are about to opt out of the Smoothing Pool.\nYour fee recipient will be changed back to your node's distributor contract.\nAll priority fees and MEV you earn via proposals will go directly to your distributor and will not be shared by the Smoothing Pool members.\n\nIf you desire, you can opt back in after one full rewards interval has passed.\n")

	// Get the gas estimate
	canResponse, err := rp.CanNodeSetSmoothingPoolStatus(false)
	if err != nil {
		return err
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canResponse.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm("Are you sure you want to leave the Smoothing Pool?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Set the fee recipient to the Smoothing Pool
	response, err := rp.NodeSetSmoothingPoolStatus(false)
	if err != nil {
		return err
	}

	fmt.Printf("Leaving the Smoothing Pool...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Println("Successfully left the Smoothing Pool.")
	fmt.Printf("%sNOTE: Your validator client will restart soon to change its fee recipient back to your node's distributor.\nYou may miss an attestation when this happens; this is normal.%s\n", colorYellow, colorReset)
	return nil

}
