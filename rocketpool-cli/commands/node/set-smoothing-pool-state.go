package node

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/terminal"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
)

func setSmoothingPoolState(c *cli.Context, optIn bool) error {
	// Get RP client
	rp := client.NewClientFromCtx(c)

	// Build the TX
	response, err := rp.Api.Node.SetSmoothingPoolRegistrationState(optIn)
	if err != nil {
		return err
	}

	// Verify
	if optIn && response.Data.NodeRegistered {
		fmt.Println("The node is already joined to the Smoothing Pool.")
		return nil
	} else if !optIn && !response.Data.NodeRegistered {
		fmt.Println("The node is not currently joined to the Smoothing Pool.")
		return nil
	}
	if response.Data.TimeLeftUntilChangeable > 0 {
		if optIn {
			fmt.Printf("You have recently left the Smoothing Pool. You must wait %s until you can join it again.\n", response.Data.TimeLeftUntilChangeable)
		} else {
			fmt.Printf("You have recently joined the Smoothing Pool. You must wait %s until you can leave it.\n", response.Data.TimeLeftUntilChangeable)
		}
		return nil
	}

	// Get messages
	var confirmMsg string
	var identifierMsg string
	var submitMsg string
	if optIn {
		fmt.Println("You are about to opt into the Smoothing Pool.")
		fmt.Println("Your fee recipient will be changed to the Smoothing Pool contract.")
		fmt.Println("All priority fees and MEV you earn via proposals will be shared equally with other members of the Smoothing Pool.")
		fmt.Println()
		fmt.Println("If you desire, you can opt back out after one full rewards interval has passed.")
		fmt.Println()
		confirmMsg = "Are you sure you want to join the Smoothing Pool?"
		identifierMsg = "joining Smoothing Pool"
		submitMsg = "Joining the Smoothing Pool..."
	} else {
		fmt.Println("You are about to opt out of the Smoothing Pool.")
		fmt.Println("Your fee recipient will be changed back to your node's distributor contract once the next Epoch has been finalized.")
		fmt.Println("All priority fees and MEV you earn via proposals will go directly to your distributor and will not be shared by the Smoothing Pool members.")
		fmt.Println()
		fmt.Println("If you desire, you can opt back in after one full rewards interval has passed.")
		fmt.Println()
		confirmMsg = "Are you sure you want to leave the Smoothing Pool?"
		identifierMsg = "leaving Smoothing Pool"
		submitMsg = "Leaving the Smoothing Pool..."
	}

	// Run the TX
	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		confirmMsg,
		identifierMsg,
		submitMsg,
	)
	if err != nil {
		if optIn {
			fmt.Println()
			return fmt.Errorf("%w\nYour fee recipient will be automatically reset to your node's distributor in a few minutes, and your validator client will restart.", err)
		}
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	if optIn {
		fmt.Println("Successfully joined the Smoothing Pool.")
	} else {
		fmt.Println("Successfully left the Smoothing Pool.")
		fmt.Printf("%sNOTE: Your validator client will restart to change its fee recipient back to your node's distributor once the next Epoch has been finalized.\nYou may miss an attestation when this happens (or multiple if you have Doppelganger Protection enabled); this is normal.%s\n", terminal.ColorYellow, terminal.ColorReset)
	}
	return nil
}
