package node

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/urfave/cli"
)

func provisionExpressTickets(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check if the node can provision express tickets
	canProvision, err := rp.CanProvisionExpressTickets()
	if err != nil {
		return err
	}

	if !canProvision.CanProvision {
		if canProvision.AlreadyProvisioned {
			fmt.Println("The node has already provisioned express tickets.")
		}
		return nil
	}

	// Provision express tickets
	response, err := rp.ProvisionExpressTickets()
	if err != nil {
		return err
	}

	fmt.Printf("Provisioning express tickets...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("The node's express tickets were successfully provisioned.\n")
	return nil
}
