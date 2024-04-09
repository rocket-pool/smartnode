package security

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
)

func join(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Build the TX
	response, err := rp.Api.Security.Join()
	if err != nil {
		return err
	}

	// Verify
	if !response.Data.CanJoin {
		fmt.Println("Cannot join the Security Council:")
		if response.Data.ProposalExpired {
			fmt.Println("The proposal for you to join the Security Council does not exist or has expired.")
		}
		if response.Data.AlreadyMember {
			fmt.Println("The node is already a member of the Security Council.")
		}
		return nil
	}

	// Run the Join TX
	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		"Are you sure you want to join the Security Council?",
		"joining Security Council",
		"Joining the Security Council...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Println("Successfully joined the security council.")
	return nil
}
