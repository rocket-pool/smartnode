package node

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/tx"
)

func nodeClearVotingDelegate(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Build the TX
	response, err := rp.Api.Node.ClearSnapshotDelegate()
	if err != nil {
		return err
	}

	// Run the TX
	err = tx.HandleTx(c, rp, response.Data.TxInfo,
		"Are you sure you remove your node's current delegate address for voting on governance proposals?",
		"removing delegate",
		"Removing delegate...",
	)
	if err != nil {
		return err
	}

	// Log & return
	fmt.Println("The node's voting delegate has been removed.")
	return nil
}
