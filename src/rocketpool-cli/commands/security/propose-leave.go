package security

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
)

func proposeLeave(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Build the TX
	response, err := rp.Api.Security.ProposeLeave()
	if err != nil {
		return err
	}

	// Run the TX
	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		"Are you sure you want to propose leaving the Security Council?",
		"proposing security council leave",
		"Proposing leaving the Security Council...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Println("Successfully submitted a proposal to leave the Security Council.")
	return nil
}
