package odao

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


func proposeLeave(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Check if proposal can be made
    canPropose, err := rp.CanProposeLeaveTNDAO()
    if err != nil {
        return err
    }
    if !canPropose.CanPropose {
        fmt.Println("Cannot propose leaving:")
        if canPropose.ProposalCooldownActive {
            fmt.Println("The node must wait for the proposal cooldown period to pass before making another proposal.")
        }
        if canPropose.InsufficientMembers {
            fmt.Println("There are not enough members in the oracle DAO to allow a member to leave.")
        }
        return nil
    }

    // Submit proposal
    response, err := rp.ProposeLeaveTNDAO()
    if err != nil {
        return err
    }

    fmt.Printf("Proposing a leave from the oracle DAO...\n")
    cliutils.PrintTransactionHash(c, response.TxHash)
    if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
        return err
    }

    // Log & return
    fmt.Printf("Successfully submitted a leave proposal with ID %d.\n", response.ProposalId)
    return nil

}

