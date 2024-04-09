package node

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
)

func setRplLockingAllowed(c *cli.Context, allowedToLock bool) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get the gas estimate
	response, err := rp.Api.Node.SetRplLockingAllowed(allowedToLock)
	if err != nil {
		return err
	}

	// Verify
	if !response.Data.CanSet {
		fmt.Println("Cannot set RPL locking status:")
		if response.Data.DifferentRplAddress {
			fmt.Println("Locking can only be modified by the node's RPL withdrawal address.")
		}
		return nil
	}

	// Run the TX
	var confirmMsg string
	var submissionMsg string
	if allowedToLock {
		confirmMsg = "Are you sure you want to allow the node to lock RPL when creating governance proposals or to challenge a proposal? Note that the bond could be lost during the proposal challenge process."
		submissionMsg = "Enabling RPL locking..."
	} else {
		confirmMsg = "Are you sure you want to block the node from locking RPL when creating governance proposals or to challenge a proposal? The node won't be able to create proposals or to create challenges."
		submissionMsg = "Disabling RPL locking..."
	}
	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		confirmMsg,
		"modifying RPL locking status",
		submissionMsg,
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	if allowedToLock {
		fmt.Printf("Successfully allowed the node to lock RPL.\n")
	} else {
		fmt.Printf("Successfully blocked the node from locking RPL.\n")
	}
	return nil
}
